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
	"crypto/ecdsa"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	"github.com/upvestco/httpsignature-proxy/config"
	"github.com/upvestco/httpsignature-proxy/service/signer/schema"
)

func NewLocalPrivateSchemeBuilder(cfg *config.BaseConfig) (*LocalPrivateSchemeBuilder, error) {
	body, err := ioutil.ReadFile(cfg.PrivateKeyFileName)
	if err != nil {
		return nil, err
	}
	return createLocalPrivateSchemeBuilder(body, cfg.KeyID, cfg.Password)
}

func NewLocalPrivateSchemeBuilderFromSeed(keyData string, cfg *config.KeyConfig) (*LocalPrivateSchemeBuilder, error) {
	return createLocalPrivateSchemeBuilder([]byte(keyData), cfg.KeyID, cfg.Password)
}

func createLocalPrivateSchemeBuilder(keyData []byte, keyId string, keyPassword string) (*LocalPrivateSchemeBuilder, error) {
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("not a PEM block")
	}

	switch block.Type {
	case schema.EcKeyType:
		rawPk, err := ssh.ParseRawPrivateKeyWithPassphrase(keyData, []byte(keyPassword))
		if err != nil {
			return nil, err
		}

		pk, ok := rawPk.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not ecdsa key")
		}

		s := &schema.Sign{
			KeyID: keyId,
			Algo:  schema.AlgoECDSA,
			Pk:    pk,
		}

		return &LocalPrivateSchemeBuilder{sign: s}, nil

	}
	return nil, errors.Errorf("unsupported private key type")
}

type LocalPrivateSchemeBuilder struct {
	sign *schema.Sign
}

func (b *LocalPrivateSchemeBuilder) GetDefaultPrivateKey() *schema.Sign {
	return b.sign
}
