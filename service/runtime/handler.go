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
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/upvestco/httpsignature-proxy/config"
	"github.com/upvestco/httpsignature-proxy/service/signer"
	"github.com/upvestco/httpsignature-proxy/service/signer/logger"
	"github.com/upvestco/httpsignature-proxy/service/signer/material"
	"github.com/upvestco/httpsignature-proxy/service/signer/request"
	"github.com/upvestco/httpsignature-proxy/service/signer/schema"
)

var (
	hostHeader           = http.CanonicalHeaderKey("host")
	acceptEncodingHeader = http.CanonicalHeaderKey("accept-encoding")
	acceptHeader         = http.CanonicalHeaderKey("accept")
	connectionHeader     = http.CanonicalHeaderKey("connection")
	userAgentHeader      = http.CanonicalHeaderKey("user-agent")

	excludedHeaders = []string{hostHeader, acceptEncodingHeader, connectionHeader, acceptHeader, userAgentHeader}
)

type Handler struct {
	cfg           *config.Config
	requestSigner request.Signer
	ssBuilder     schema.SigningSchemeBuilder
	log           logger.Logger
}

func newHandler(cfg *config.Config, ssBuilder schema.SigningSchemeBuilder, log logger.Logger) *Handler {
	return &Handler{
		cfg:           cfg,
		log:           log,
		requestSigner: request.New(log),
		ssBuilder:     ssBuilder,
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

func (h *Handler) ServeHTTP(rw http.ResponseWriter, inReq *http.Request) {
	h.log.Log("\nSend request:\n")
	ctx, cancel := context.WithTimeout(inReq.Context(), h.cfg.DefaultTimeout)
	defer cancel()

	url := fmt.Sprintf("%s%s", h.cfg.BaseUrl, inReq.URL.Path)
	h.log.LogF(" - To url '%s'\n", url)

	body, err := ioutil.ReadAll(inReq.Body)
	if err != nil {
		h.writeError(rw, http.StatusInternalServerError, err)
		return
	}

	outReq, err := http.NewRequestWithContext(ctx, inReq.Method, url, bytes.NewBuffer(body))
	if err != nil {
		h.writeError(rw, http.StatusInternalServerError, err)
		return
	}

	h.copyHeaders(inReq, outReq)

	h.addRequiredHeaders(outReq)

	sign := h.ssBuilder.GetDefaultPrivateKey()

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
