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

package signer

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/upvestco/httpsignature-proxy/service/logger"

	"github.com/upvestco/httpsignature-proxy/service/signer/request"
)

var _ http.RoundTripper = (*RoundTripper)(nil)

var ErrSigning = errors.New("signing proxy: unable to sign request")

// NewHTTPClient will create a new http.Client and add the signing transport to it.
func NewHTTPClient(signer request.Signer, signingKey request.RequestSigner, log logger.Logger) *http.Client {
	return &http.Client{
		Transport: NewTransport(http.DefaultTransport, signer, signingKey, log),
	}
}

// NewTransport will create a new http.RoundTripper that can be used in http.Client to sign requests transparently.
// Underlying http.RoundTripper cannot be nil, if unsure, you can use http.DefaultTransport.
func NewTransport(inner http.RoundTripper, signer request.Signer, signingKey request.RequestSigner, log logger.Logger) *RoundTripper {
	return &RoundTripper{
		inner:      inner,
		signer:     signer,
		signingKey: signingKey,
		log:        log,
	}
}

// RoundTripper implements a http transport middleware for signing outgoing http requests.
type RoundTripper struct {
	inner      http.RoundTripper
	signer     request.Signer
	signingKey request.RequestSigner
	log        logger.Logger
}

// RoundTrip does the actual signing and sending.
func (r RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	err := r.signer.Sign(req, r.signingKey)
	if err != nil {
		r.log.LogF("signing error: %v", err)
		return nil, ErrSigning
	}

	rsp, err := r.inner.RoundTrip(req)
	if err != nil {
		return nil, errors.Wrap(err, "signing proxy: unable to perform request")
	}

	return rsp, nil
}
