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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/upvestco/httpsignature-proxy/config"
	"github.com/upvestco/httpsignature-proxy/service/signer"
	"github.com/upvestco/httpsignature-proxy/service/signer/logger"
	"github.com/upvestco/httpsignature-proxy/service/signer/material"
	"github.com/upvestco/httpsignature-proxy/service/signer/request"
)

var (
	hostHeader           = http.CanonicalHeaderKey("host")
	acceptEncodingHeader = http.CanonicalHeaderKey("accept-encoding")
	acceptHeader         = http.CanonicalHeaderKey("accept")
	connectionHeader     = http.CanonicalHeaderKey("connection")
	userAgentHeader      = http.CanonicalHeaderKey("user-agent")

	upvestClientID = "upvest-client-id"
	tokenEndpoint  = "/auth/token"

	excludedHeaders = []string{hostHeader, acceptEncodingHeader, connectionHeader, userAgentHeader}
)

type Handler struct {
	signerConfigs map[string]SignerConfig
	cfg           *config.Config
	requestSigner request.Signer
	log           logger.Logger
}

func newHandler(cfg *config.Config, signerConfigs map[string]SignerConfig, log logger.Logger) *Handler {
	return &Handler{
		cfg:           cfg,
		log:           log,
		requestSigner: request.New(log),
		signerConfigs: signerConfigs,
	}
}

func (h *Handler) writeResponse(rw http.ResponseWriter, code int, headers map[string][]string, resp []byte) {
	excludedHeaders := map[string]struct{}{
		material.SignatureHeader:      {},
		material.SignatureInputHeader: {},
		"Host":                        {},
	}

	for name, values := range headers {
		if _, ok := excludedHeaders[name]; ok {
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

func (h *Handler) copyHeaders(in *http.Request, out *http.Request) {
	for headerName, value := range in.Header {
		if h.excludeHeader(headerName) {
			continue
		}
		out.Header.Add(headerName, strings.Join(value, ","))
	}
}

func (h *Handler) addRequiredHeaders(req *http.Request) {
	if res := req.Header.Get(acceptHeader); res == "" {
		req.Header.Add(acceptHeader, `*/*`)
		fmt.Printf(" - Header '%s' added with value '%s'\n", acceptHeader, `*/*`)
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

func (h *Handler) getSignerConfig(clientID string) (SignerConfig, error) {
	signerCfg, ok := h.signerConfigs[clientID]
	if !ok {
		return h.getDefaultSigner()
	}
	h.log.LogF(" - Used signer for clientID %s\n", clientID)
	return signerCfg, nil
}
func (h *Handler) getDefaultSigner() (SignerConfig, error) {
	signerCfg, ok := h.signerConfigs[config.DefaultClientKey]
	if !ok {
		return SignerConfig{}, errors.New("unknown clientID, please, check your signing proxy configuration")
	}
	h.log.Log(" - Used default signer\n")
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
	h.log.Log("\nSend request:\n")
	ctx, cancel := context.WithTimeout(inReq.Context(), h.cfg.DefaultTimeout)
	defer cancel()

	clientID, err := h.getClientID(inReq)
	if err != nil {
		err = errors.Wrap(err, "invalid clientID, please, check your signing proxy configuration")
		h.writeError(rw, http.StatusInternalServerError, err)
		return
	}
	h.log.LogF(" - For Client ID: %s\n", clientID)

	signerCfg, err := h.getSignerConfig(clientID)
	if err != nil {
		h.log.Log("Signer not found")
		h.writeError(rw, http.StatusInternalServerError, err)
		return
	}

	toUrl := fmt.Sprintf("%s%s", signerCfg.KeyConfig.BaseUrl, inReq.URL.Path)
	h.log.LogF(" - To url '%s'\n", toUrl)

	body, err := ioutil.ReadAll(inReq.Body)
	if err != nil {
		h.writeError(rw, http.StatusInternalServerError, err)
		return
	}

	outReq, err := http.NewRequestWithContext(ctx, inReq.Method, toUrl, bytes.NewBuffer(body))
	if err != nil {
		h.writeError(rw, http.StatusInternalServerError, err)
		return
	}

	h.copyHeaders(inReq, outReq)

	h.addRequiredHeaders(outReq)

	sign := signerCfg.SignBuilder.GetDefaultPrivateKey()

	httpClient := signer.NewHTTPClient(h.requestSigner, sign)

	resp, err := httpClient.Do(outReq)
	if err != nil {
		h.writeError(rw, http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.writeError(rw, http.StatusInternalServerError, err)
		return
	}

	h.log.Log("Response:\n")
	h.log.LogF(" - Status '%d'\n", resp.StatusCode)
	h.log.LogF(" - Headers:\n")
	for key := range resp.Header {
		h.log.LogF("    %s:%s\n", key, resp.Header[key])
	}

	h.writeResponse(rw, resp.StatusCode, resp.Header, data)
}
