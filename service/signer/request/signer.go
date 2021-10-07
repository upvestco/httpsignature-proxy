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
	"net/http"

	"github.com/pkg/errors"

	"github.com/upvestco/httpsignature-proxy/service/signer/logger"
	"github.com/upvestco/httpsignature-proxy/service/signer/material"
)

type Signer interface {
	Sign(req *http.Request, s RequestSigner) error
}

type RequestSigner interface {
	SignRequest(m *material.Material, r *http.Request, log logger.Logger) error
}

func New(log logger.Logger) Signer {
	return &requestSigner{
		log: log,
	}
}

type requestSigner struct {
	log logger.Logger
}

func (e requestSigner) Sign(req *http.Request, s RequestSigner) error {
	m, err := material.MaterialFromRequest(req)
	if err != nil {
		return errors.Wrap(err, "MaterialFromRequest")
	}
	return errors.Wrap(s.SignRequest(m, req, e.log), "AddSignatureHeaders SignRequest")
}
