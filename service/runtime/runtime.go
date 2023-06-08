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
	"log"
	"net/http"

	"github.com/upvestco/httpsignature-proxy/config"
	"github.com/upvestco/httpsignature-proxy/service/signer/logger"
	"github.com/upvestco/httpsignature-proxy/service/signer/schema"
)

type Runtime struct {
	cfg           *config.Config
	signerConfigs map[string]SignerConfig
	logger        logger.Logger
	server        *http.Server
}

type SignerConfig struct {
	SignBuilder schema.SigningSchemeBuilder
	KeyConfig   config.BaseConfig
}

func NewRuntime(cfg *config.Config, signerConfigs map[string]SignerConfig) Runtime {
	return Runtime{
		cfg:           cfg,
		logger:        logger.New(cfg.VerboseMode),
		signerConfigs: signerConfigs,
	}
}

func (r *Runtime) Run() {
	r.installServer()
}

func (r *Runtime) installServer() {
	handler := newHandler(r.cfg, r.signerConfigs, r.logger)
	r.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", r.cfg.Port),
		Handler: handler}
	log.Fatal(r.server.ListenAndServe())
}
