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
	"net"
	"net/http"

	"github.com/pkg/errors"
	"github.com/upvestco/httpsignature-proxy/config"
	"github.com/upvestco/httpsignature-proxy/service/logger"
	"github.com/upvestco/httpsignature-proxy/service/signer/schema"
)

type Proxy struct {
	cfg               *config.Config
	signerConfigs     map[string]SignerConfig
	logger            logger.Logger
	server            *http.Server
	userCredentialsCh chan UserCredentials
}

type SignerConfig struct {
	SignBuilder schema.SigningSchemeBuilder
	KeyConfig   config.BaseConfig
}

func NewProxy(cfg *config.Config, signerConfigs map[string]SignerConfig, userCredentialsCh chan UserCredentials, logger logger.Logger) Proxy {
	return Proxy{
		cfg:               cfg,
		logger:            logger,
		signerConfigs:     signerConfigs,
		userCredentialsCh: userCredentialsCh,
	}
}

func (r *Proxy) Run() error {
	addr := net.JoinHostPort("localhost", fmt.Sprintf("%d", r.cfg.Port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "Listen")
	}
	handler := newHandler(r.cfg, r.signerConfigs, r.userCredentialsCh, r.logger)
	r.server = &http.Server{
		Handler: handler,
	}
	go func() {
		if err := r.server.Serve(listener); err != nil {
			r.logger.PrintLn(err.Error())
		}
	}()
	return nil
}
