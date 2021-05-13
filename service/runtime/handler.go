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
	"github.com/upvestco/httpsignature-proxy/service/signer/material"
	"github.com/upvestco/httpsignature-proxy/service/signer/request"
	"github.com/upvestco/httpsignature-proxy/service/signer/schema"
)

type Handler struct {
	cfg           *config.Config
	requestSigner request.Signer
	httpClient    *http.Client
}

func newHandler(cfg *config.Config, ssBuilder schema.SigningSchemeBuilder) *Handler {
	return &Handler{
		cfg:           cfg,
		requestSigner: request.New(ssBuilder),
		httpClient:    &http.Client{},
	}
}

func (h *Handler) writeResponse(rw http.ResponseWriter, code int, headers map[string][]string, resp []byte) {
	excludedHeaders := map[string]struct{}{
		material.Signature:      {},
		material.SignatureInput: {},
		"Host":                  {},
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
		if headerName == "Host" {
			continue
		}
		out.Header.Add(headerName, strings.Join(value, ","))
	}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, inReq *http.Request) {
	ctx, cancel := context.WithTimeout(inReq.Context(), h.cfg.DefaultTimeout)
	defer cancel()

	url := fmt.Sprintf("%s%s", h.cfg.BaseUrl, inReq.URL.Path)
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

	if err = h.requestSigner.Sign(outReq); err != nil {
		h.writeError(rw, http.StatusInternalServerError, err)
		return
	}

	resp, err := h.httpClient.Do(outReq)
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
	h.writeResponse(rw, resp.StatusCode, resp.Header, data)
}
