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
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/upvestco/httpsignature-proxy/config"
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
	suite.Run(t, &TestRuntimeSuite{})
}

type TestRuntimeSuite struct {
	suite.Suite
	testServ *testService
}

func (s *TestRuntimeSuite) SetupSuite() {
	cfg := &config.Config{
		Port:               runtimePort,
		BaseUrl:            testBaseUrl,
		PrivateKeyFileName: "",
		Password:           testPass,
		DefaultTimeout:     30 * time.Second,
		KeyID:              testKeyID,
	}
	s.setupTestService()
	s.setupRuntime(cfg)
}

func (s *TestRuntimeSuite) setupTestService() {
	s.testServ = &testService{}
	s.testServ.Start(s.T())
}

func (s *TestRuntimeSuite) setupRuntime(cfg *config.Config) {
	privateSchemeBuilder, err := signer.NewLocalPrivateSchemeBuilderFromSeed(privateTestKey, cfg)
	assert.NoError(s.T(), err)
	r := NewRuntime(cfg, privateSchemeBuilder)
	go r.Run()
}

func (s *TestRuntimeSuite) Test_RuntimeRun() {
	url := fmt.Sprintf("http://localhost:%d/%s", runtimePort, "endpoint")
	pl := []byte("This is the body")
	body := bytes.NewBuffer(pl)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, body)
	require.NoError(s.T(), err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	assert.Exactly(s.T(), http.StatusOK, resp.StatusCode)
	respBody, err := ioutil.ReadAll(resp.Body)
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
		Handler: h}

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
	body, err := ioutil.ReadAll(r.Body)
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
	body, err := ioutil.ReadAll(r.Body)
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
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	e.router.ServeHTTP(w, r)
}
