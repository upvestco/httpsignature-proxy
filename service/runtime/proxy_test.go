/*
Copyright © 2021 Upvest GmbH <support@upvest.co>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/upvestco/httpsignature-proxy/config"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"github.com/upvestco/httpsignature-proxy/service/signer"
	"github.com/upvestco/httpsignature-proxy/service/signer/material"
)

const (
	privateTestKey = `-----BEGIN EC PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: AES-256-CBC,436A5FC28B24B33544562F9556720FD4

YEzPUOC0uwoHh1GiwX/XI6TDull44JydY1okKRFxbU9X8tfTymhFX0QVa9vIVZmf
Z9pVt7ezzsXsTa83aTfeDQjMVkHQFnp7K5V/s4qAElRjSZvKdkGRjvgqAHnz2DYw
qFs3oIGIa4fr1C7SXmMyCohmJznOH3kGu73fV6GJkdc=
-----END EC PRIVATE KEY-----`

	testPass    = "123456"
	testKeyID   = "key_id"
	testBaseUrl = "http://localhost:3001"

	runtimePort  = 3002
	verifierPort = 3001
)

func TestRuntime_Run(t *testing.T) {
	suite.Run(t, &TestProxySuite{})
}

type TestProxySuite struct {
	suite.Suite
	testServ *testService
	clientID uuid.UUID
}

func (s *TestProxySuite) SetupSuite() {
	s.clientID = uuid.New()
	cfg := &config.Config{
		Port:           runtimePort,
		DefaultTimeout: 30 * time.Second,
		PullDelay:      time.Second,
		KeyConfigs: []config.KeyConfig{
			{
				BaseConfig: config.BaseConfig{
					BaseUrl:            testBaseUrl,
					PrivateKeyFileName: "",
					Password:           testPass,
					KeyID:              testKeyID,
				},
				ClientID: s.clientID.String(),
			},
		},
	}
	s.setupTestService()
	s.setupProxy(cfg)
}

func (s *TestProxySuite) setupTestService() {
	s.testServ = &testService{}
	s.testServ.Start(s.T())
}

func (s *TestProxySuite) setupProxy(cfg *config.Config) {
	signerConfigs := make(map[string]SignerConfig)
	for i := range cfg.KeyConfigs {
		privateSchemeBuilder, err := signer.NewLocalPrivateSchemeBuilderFromSeed(privateTestKey, &cfg.KeyConfigs[i])
		require.NoError(s.T(), err)
		signerConfigs[cfg.KeyConfigs[i].ClientID] = SignerConfig{
			SignBuilder: privateSchemeBuilder,
			KeyConfig:   cfg.KeyConfigs[i].BaseConfig,
		}
	}
	r := NewProxy(cfg, signerConfigs, nil, logger.New(false))
	require.NoError(s.T(), r.Run())
	time.Sleep(1 * time.Second)
}

func (s *TestProxySuite) Test_ProxyRun() {
	url := fmt.Sprintf("http://localhost:%d/%s", runtimePort, "endpoint?param=val")
	pl := []byte("This is the body")
	body := bytes.NewBuffer(pl)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, body)
	require.NoError(s.T(), err)
	req.Header.Set("upvest-client-id", s.clientID.String())
	param := req.URL.Query().Get("param")
	require.NotEmpty(s.T(), param)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Exactly(s.T(), http.StatusOK, resp.StatusCode)
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	require.Equal(s.T(), pl, respBody)
}

type testService struct {
	server *http.Server
}

func (s *testService) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = s.server.Shutdown(ctx)
}

func (s *testService) Start(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/endpoint", s.endpoint).Methods(http.MethodPost)

	h := &handler{router: router, t: t}

	s.server = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", verifierPort),
		Handler: h,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if err.Error() != http.ErrServerClosed.Error() {
				log.Fatal(err)
			}
		}
	}()
}

func (s *testService) endpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(body)
}

type handler struct {
	router *mux.Router
	t      *testing.T
}

func (e *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	require.NoError(e.t, err)
	require.True(e.t, len(body) != 0)
	defer r.Body.Close()
	headers := r.Header
	if val, ok := headers[material.SignatureHeader]; !ok || val == nil {
		e.t.Error("request doesn't have signature header")
	}
	if val, ok := headers[material.SignatureInputHeader]; !ok || val == nil {
		e.t.Error("request doesn't have signature input header")
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	e.router.ServeHTTP(w, r)
}
