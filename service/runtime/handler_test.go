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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upvestco/httpsignature-proxy/config"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"github.com/upvestco/httpsignature-proxy/service/signer"
	"github.com/upvestco/httpsignature-proxy/service/tunnels"
)

func newTestHandler(t *testing.T, backendURL string, ch chan tunnels.UserCredentials) (*Handler, uuid.UUID) {
	t.Helper()
	clientID := uuid.New()
	keyCfg := config.KeyConfig{
		BaseConfig: config.BaseConfig{
			BaseUrl:  backendURL,
			Password: testPass,
			KeyID:    testKeyID,
		},
		ClientID: clientID.String(),
	}
	builder, err := signer.NewLocalPrivateSchemeBuilderFromSeed(privateTestKey, &keyCfg)
	require.NoError(t, err)

	signerConfigs := map[string]SignerConfig{
		clientID.String(): {SignBuilder: builder, KeyConfig: keyCfg.BaseConfig},
	}
	cfg := &config.Config{DefaultTimeout: 30 * time.Second}
	return newHandler(cfg, signerConfigs, ch, logger.New(false)), clientID
}

func TestHandler_AuthToken_NoReaderOnChannel(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	ch := make(chan tunnels.UserCredentials) // unbuffered, no reader
	h, clientID := newTestHandler(t, backend.URL, ch)

	body := fmt.Sprintf("client_id=%s&client_secret=secret", clientID)
	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(body))
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		h.ServeHTTP(rec, req)
		close(done)
	}()

	select {
	case <-done:
		assert.Equal(t, http.StatusOK, rec.Code)
	case <-time.After(3 * time.Second):
		t.Fatal("ServeHTTP blocked: channel send deadlock")
	}
}

func TestHandler_AuthToken_WithReaderOnChannel(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	ch := make(chan tunnels.UserCredentials, 1)
	h, clientID := newTestHandler(t, backend.URL, ch)

	secret := "test-secret"
	body := fmt.Sprintf("client_id=%s&client_secret=%s", clientID, secret)
	req := httptest.NewRequest(http.MethodPost, "/auth/token", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	select {
	case uc := <-ch:
		assert.Equal(t, clientID.String(), uc.ClientID)
		assert.Equal(t, secret, uc.ClientSecret)
	default:
		t.Fatal("expected credentials on channel")
	}
}
