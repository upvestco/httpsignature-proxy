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
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"github.com/upvestco/httpsignature-proxy/service/signer"
	"github.com/upvestco/httpsignature-proxy/service/tunnels"

	"github.com/upvestco/httpsignature-proxy/config"
	"github.com/upvestco/httpsignature-proxy/service/signer/material"
	"github.com/upvestco/httpsignature-proxy/service/signer/request"
)

var (
	hostHeader           = http.CanonicalHeaderKey("host")
	acceptEncodingHeader = http.CanonicalHeaderKey("accept-encoding")
	connectionHeader     = http.CanonicalHeaderKey("connection")
	userAgentHeader      = http.CanonicalHeaderKey("user-agent")

	acceptHeader   = http.CanonicalHeaderKey("accept")
	upvestClientID = "upvest-client-id"
	tokenEndpoint  = "/auth/token"

	excludedHeaders = []string{hostHeader, acceptEncodingHeader, connectionHeader, userAgentHeader}
)

type Handler struct {
	signerConfigs     map[string]SignerConfig
	cfg               *config.Config
	requestSigner     request.Signer
	log               logger.Logger
	userCredentialsCh chan tunnels.UserCredentials
}

func newHandler(cfg *config.Config, signerConfigs map[string]SignerConfig, userCredentialsCh chan tunnels.UserCredentials, log logger.Logger) *Handler {
	return &Handler{
		cfg:               cfg,
		log:               log,
		requestSigner:     request.New(log),
		signerConfigs:     signerConfigs,
		userCredentialsCh: userCredentialsCh,
	}
}

func (h *Handler) writeResponse(rw http.ResponseWriter, code int, headers map[string][]string, resp []byte) {
	excludedOutputHeaders := map[string]struct{}{
		material.SignatureHeader:      {},
		material.SignatureInputHeader: {},
		"Host":                        {},
	}

	for name, values := range headers {
		if _, ok := excludedOutputHeaders[name]; ok {
			continue
		}
		for _, val := range values {
			rw.Header().Add(name, val)
		}
	}
	rw.WriteHeader(code)

	if resp != nil {
		_, _ = rw.Write(resp)
	}
}

func (h *Handler) writeError(rw http.ResponseWriter, code int, err error) {
	rw.Header().Set("Content-Type", "application/json;charset=UTF-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("Pragma", "no-cache")
	rw.WriteHeader(code)

	errResp := struct {
		Err string `json:"error"`
	}{
		Err: err.Error(),
	}
	respBytes, _ := json.Marshal(errResp)

	if respBytes != nil {
		_, _ = rw.Write(respBytes)
	}
}

func (h *Handler) copyHeaders(in *http.Request, out *http.Request, ll logger.Logger) {
	for headerName, value := range in.Header {
		if h.excludeHeader(headerName) {
			continue
		}
		out.Header.Add(headerName, strings.Join(value, ","))
	}
	ll.Log(" - Headers that will be excluded if presented:")
	for _, excludedHeader := range excludedHeaders {
		ll.LogF("   - %s:", excludedHeader)
	}
}

func (h *Handler) addRequiredHeaders(req *http.Request, ll logger.Logger) {
	if res := req.Header.Get(acceptHeader); res == "" {
		req.Header.Add(acceptHeader, `*/*`)
		ll.LogF(" - Header '%s' added with value '%s'", acceptHeader, `*/*`)
	}
}

func (h *Handler) excludeHeader(headerName string) bool {
	cHeaderName := http.CanonicalHeaderKey(headerName)
	for _, excludedHeader := range excludedHeaders {
		if excludedHeader == cHeaderName {
			return true
		}
	}
	return false
}

func (h *Handler) getSignerConfig(clientID string, ll logger.Logger) (SignerConfig, error) {
	signerCfg, ok := h.signerConfigs[clientID]
	if !ok {
		return h.getDefaultSigner(ll)
	}
	ll.LogF(" - Used signer for clientID %s", clientID)
	return signerCfg, nil
}
func (h *Handler) getDefaultSigner(ll logger.Logger) (SignerConfig, error) {
	signerCfg, ok := h.signerConfigs[config.DefaultClientKey]
	if !ok {
		return SignerConfig{}, errors.New("unknown clientID, please, check your signing proxy configuration")
	}
	ll.Log(" - Used default signer")
	return signerCfg, nil
}

func (h *Handler) getClientID(req *http.Request) (string, error) {
	var clientID string
	var err error
	if req.URL.Path == tokenEndpoint {
		clientID, err = h.getClientIDFromBody(req)
		if err != nil {
			return "", errors.Wrap(err, "failed to get client id from request")
		}
	} else {
		clientID = req.Header.Get(upvestClientID)
	}
	if _, err := uuid.Parse(clientID); err != nil {
		return "", errors.New("failed to get client id from request")
	}
	return clientID, nil
}

func (h *Handler) getClientIDFromBody(req *http.Request) (string, error) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to get info from body")
	}
	queryValues, err := url.ParseQuery(string(data))
	if err != nil {
		return "", errors.New("failed to parse body")
	}
	req.Body = io.NopCloser(bytes.NewReader(data))
	return queryValues.Get("client_id"), nil
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, inReq *http.Request) {
	ll := h.log
	if len(inReq.Header.Get(logger.HttpProxyNoLogging)) > 0 {
		ll = logger.NoVerboseLogger
	}
	requestBody := h.proxy(rw, inReq, ll)
	path := inReq.URL.Path
	if path == tokenEndpoint && requestBody != nil {
		uc := h.parseAuthTokenBody(requestBody)
		if !uc.Empty() && h.userCredentialsCh != nil {
			h.userCredentialsCh <- uc
		}
	}
}

func (h *Handler) proxy(rw http.ResponseWriter, inReq *http.Request, ll logger.Logger) []byte {
	ll.Log("\nSend request:")
	ctx, cancel := context.WithTimeout(inReq.Context(), h.cfg.DefaultTimeout)
	defer cancel()

	clientID, err := h.getClientID(inReq)
	if err != nil {
		err = errors.Wrap(err, "invalid clientID, please, check your signing proxy configuration")
		h.writeError(rw, http.StatusInternalServerError, err)
		return nil
	}
	ll.LogF(" - For Client ID: %s", clientID)

	signerCfg, err := h.getSignerConfig(clientID, ll)
	if err != nil {
		ll.Log("Signer not found")
		h.writeError(rw, http.StatusInternalServerError, err)
		return nil
	}

	toUrl, err := url.Parse(signerCfg.KeyConfig.BaseUrl)
	if err != nil {
		ll.Log("Wrong base URL")
		h.writeError(rw, http.StatusInternalServerError, err)
		return nil
	}
	toUrl.Path = inReq.URL.Path

	toUrl.RawQuery = inReq.URL.RawQuery
	ll.LogF(" - To url '%s'", toUrl.String())

	requestBody, err := io.ReadAll(inReq.Body)
	if err != nil {
		h.writeError(rw, http.StatusInternalServerError, err)
		return nil
	}
	outReq, err := http.NewRequestWithContext(ctx, inReq.Method, toUrl.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		h.writeError(rw, http.StatusInternalServerError, err)
		return nil
	}

	h.copyHeaders(inReq, outReq, ll)

	h.addRequiredHeaders(outReq, ll)

	sign := signerCfg.SignBuilder.GetDefaultPrivateKey()

	httpClient := signer.NewHTTPClient(h.requestSigner, sign, ll)

	resp, err := httpClient.Do(outReq)
	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			h.writeError(rw, http.StatusGatewayTimeout, err)
		case errors.Is(err, signer.ErrSigning):
			h.writeError(rw, http.StatusBadRequest, err)
		default:
			h.writeError(rw, http.StatusInternalServerError, err)
		}

		return nil
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		h.writeError(rw, http.StatusInternalServerError, err)
		return nil
	}

	ll.Log("\n=====================")
	ll.Log("Response:")
	ll.LogF(" - Status '%d'", resp.StatusCode)
	ll.LogF(" - Headers:")
	for key := range resp.Header {
		ll.LogF("    %s:%s", key, resp.Header[key])
	}

	h.writeResponse(rw, resp.StatusCode, resp.Header, data)
	return requestBody
}

func (h *Handler) parseAuthTokenBody(body []byte) tunnels.UserCredentials {
	var res tunnels.UserCredentials
	for _, keyValue := range strings.Split(string(body), "&") {
		kv := strings.Split(keyValue, "=")
		if kv[0] == "client_id" {
			res.ClientID = strings.TrimSpace(kv[1])
		}
		if kv[0] == "client_secret" {
			res.ClientSecret = strings.TrimSpace(kv[1])
		}
	}
	return res
}
