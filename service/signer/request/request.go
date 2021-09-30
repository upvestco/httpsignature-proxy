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

package request

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/upvestco/httpsignature-proxy/service/signer/material"
	"github.com/upvestco/httpsignature-proxy/service/signer/schema"
)

var ErrNoKey = errors.New("No key")

type Signer interface {
	Sign(req *http.Request) error
}

func New(ssBuilder schema.SigningSchemeBuilder) Signer {
	return &requestSigner{ssBuilder}
}

type requestSigner struct {
	ssBuilder schema.SigningSchemeBuilder
}

func (e *requestSigner) Sign(req *http.Request) error {
	m := material.NewMaterial()
	if err := m.AppendHeaders(req.Header); err != nil {
		return errors.Wrap(err, "Sign:Append error")
	}

	target := req.Method + " " + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target = target + "?" + req.URL.RawQuery
	}

	m.Append0("*request-target", strings.ToLower(target))
	reqBody, err := req.GetBody()
	if err != nil {
		return errors.Wrap(err, "Sign: Get Body error")
	}
	defer reqBody.Close()

	body, err := ioutil.ReadAll(reqBody)
	if err != nil {
		return errors.Wrap(err, "Sign: ReadAll error")
	}

	m.PrepareWithBodyForSign(body)

	signer := e.ssBuilder.GetDefaultPrivateKey()
	return signer.SignRequest(m, req)
}
