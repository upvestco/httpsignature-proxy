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

package schema

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/upvestco/httpsignature-proxy/service/logger"

	"github.com/upvestco/httpsignature-proxy/service/signer/material"
)

type SigningSchemeBuilder interface {
	GetDefaultPrivateKey() *Sign
}

const (
	algoEd25519 = "Ed25519"
	algoECDSA   = "ECDSA"
)

const (
	AlgoECDSA = "ECDSA"
	EcKeyType = "EC PRIVATE KEY"
)

const (
	signingVersion = "15"
)

var (
	errUnsupportedAlgorithm = errors.New("unsupported algorithm")
	errWrongPrivateKey      = errors.New("wrong private key")
)

type Sign struct {
	KeyID string
	Algo  string
	Pk    interface{}
	Pub   interface{}
}

func (e *Sign) SignRequest(m *material.Material, r *http.Request, log logger.Logger) error {
	return errors.Wrap(e.sign(m, r.Header, log), "sign")
}

func (e *Sign) sign(m *material.Material, headers http.Header, log logger.Logger) error {
	const sigID = "sig1"
	body, signatureParams, err := m.GetBody(e.KeyID)

	if err != nil {
		return errors.Wrap(err, "GetBody")
	}
	signBytes, err := e.calculateSignBytes(body)
	if err != nil {
		return errors.Wrap(err, "calculateSignBytes")
	}
	hash := b64.StdEncoding.EncodeToString(signBytes)

	headers.Set(material.SignatureInputHeader, fmt.Sprintf("%s=%s", sigID, signatureParams))
	headers.Set(material.SignatureHeader, fmt.Sprintf("%s=:%s:", sigID, hash))
	headers.Set(material.SignatureHeader, fmt.Sprintf("%s=:%s:", sigID, hash))
	headers.Set(material.SigningVersionHeader, signingVersion)

	log.LogF(" - Header '%s' added with value '%s'", material.SignatureInputHeader, signatureParams)
	log.LogF(" - Header '%s' added with value '%s'", material.SignatureHeader, hash)
	log.LogF(" - Header '%s' added with value '%s'", material.SigningVersionHeader, signingVersion)

	log.Log(" - Headers list:")
	for key, vals := range headers {
		for _, val := range vals {
			log.LogF("   - %s: '%v'", key, val)
		}
	}
	log.LogF(" - Body for signing: \n'%s'", body)
	log.LogF(" - Body with escaped \\n :\n'%s'", strings.ReplaceAll(string(body), "\n", "\\n"))

	return nil
}

func (e *Sign) calculateSignBytes(message []byte) ([]byte, error) {
	var signBytes []byte
	var err error
	switch e.Algo {
	case algoEd25519:
		pk, ok := e.Pk.(*ed25519.PrivateKey)
		if !ok {
			return nil, errors.Wrap(errWrongPrivateKey, "not a ed25519.PrivateKey")
		}
		signBytes = ed25519.Sign(*pk, message)
	case algoECDSA:
		pk, ok := e.Pk.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.Wrap(errWrongPrivateKey, "not a ecdsa.PrivateKey")
		}
		hash := sha512.Sum512(message)
		signBytes, err = ecdsa.SignASN1(rand.Reader, pk, hash[:])
		if err != nil {
			return nil, errors.Wrap(err, "signature SignASN1")
		}
	default:
		return nil, errUnsupportedAlgorithm
	}
	return signBytes, nil
}
