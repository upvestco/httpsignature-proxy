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
	"crypto/rand"
	"crypto/sha512"
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/upvestco/httpsignature-proxy/service/signer/material"
)

type SigningSchemeBuilder interface {
	GetDefaultPrivateKey() *Sign
}

type SigningSchema interface {
	Sign(m *material.Material, w http.ResponseWriter) error
}

const (
	AlgoECDSA = "ECDSA"
	EcKeyType = "EC PRIVATE KEY"
)

var (
	errUnsupportedAlgorithm = errors.New("unsupported algorithm")
	errWrongPrivateKey      = errors.New("wrong private key")
)

type Sign struct {
	KeyId string
	Algo  string
	Pk    interface{}
}

func (e *Sign) SignRequest(m *material.Material, r *http.Request) error {
	signBytes, err := e.signature(m)
	if err != nil {
		return errors.Wrap(err, "Error creating signature")
	}
	sign := b64.StdEncoding.EncodeToString(signBytes)
	input := e.signatureInput(m)
	r.Header.Add(material.SignatureInput, input)
	r.Header.Add(material.Signature, fmt.Sprintf("sig1=:%s:", sign))
	return nil
}

func (e *Sign) signature(m *material.Material) ([]byte, error) {
	var sign []byte
	var err error
	switch e.Algo {
	case AlgoECDSA:
		pk, ok := e.Pk.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.Wrap(errWrongPrivateKey, "not a ecdsa.PrivateKey")
		}
		hash := sha512.Sum512(m.Data)
		sign, err = ecdsa.SignASN1(rand.Reader, pk, hash[:])
		if err != nil {
			return nil, errors.Wrap(err, "error signing with ecdsa.SignASN1")
		}
	default:
		return nil, errUnsupportedAlgorithm
	}
	return sign, nil
}

func (e *Sign) signatureInput(m *material.Material) string {
	return fmt.Sprintf("sig1=(%s); keyId=\"%s\"; alg=%s; created=%s",
		strings.Join(m.Names, ", "), e.KeyId, e.Algo, m.Created)
}
