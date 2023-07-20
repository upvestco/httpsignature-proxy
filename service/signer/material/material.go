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

package material

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/ory/x/randx"
	"github.com/pkg/errors"
)

const (
	ietfMethod          = "@method"
	ietfPath            = "@path"
	ietfQuery           = "@query"
	ietfContentDigest   = "content-digest"
	ietfSignatureParams = "@signature-params"
)

var (
	SignatureHeader      = textproto.CanonicalMIMEHeaderKey("Signature")
	SignatureInputHeader = textproto.CanonicalMIMEHeaderKey("Signature-Input")
	SigningVersionHeader = textproto.CanonicalMIMEHeaderKey("Upvest-Signature-Version")
	ContentDigestHeader  = textproto.CanonicalMIMEHeaderKey("Content-Digest")

	ignoreHeadersWithPrefix = []string{
		"cf-",
		"cdn-",
		"cookie",
		"x-",
	}
)

type Material struct {
	Data           map[string]string
	Names          []string
	Created        string
	Expires        string
	Nonce          string
	Body           []byte
	SourceBody     []byte
	SignatureInput string
}

func newMaterial() *Material {
	return &Material{
		Data:    make(map[string]string),
		Names:   make([]string, 0),
		Created: fmt.Sprintf("%d", time.Now().Unix()),
		Nonce:   randx.MustString(10, randx.Numeric),
		Expires: fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()),
	}
}

func (e *Material) GetBody(keyID string) ([]byte, string, error) {
	quoteNames := make([]string, len(e.Names))
	for i, s := range e.Names {
		quoteNames[i] = fmt.Sprintf("%q", s)
	}
	names := strings.Join(quoteNames, " ")

	signatureParams := fmt.Sprintf("(%s);keyid=%q;created=%s;nonce=%q;expires=%s", names, keyID, e.Created, e.Nonce, e.Expires)

	e.CompleteWithSourceBody(ietfSignatureParams, signatureParams)

	return e.Body, signatureParams, nil
}

func MaterialFromRequest(req *http.Request) (*Material, error) {
	e := newMaterial()

	if err := e.AppendHeaders(req.Header); err != nil {
		return nil, errors.Wrap(err, "appendHeaders")
	}

	e.AppendValue(ietfMethod, req.Method)
	if len(req.URL.Path) > 0 {
		e.AppendValue(ietfPath, req.URL.Path)
	}
	body, err := GetRequestBody(req)
	if err != nil {
		return nil, errors.Wrap(err, "getRequestBody")
	}
	if len(body) > 0 {
		e.addContentDigest(body, req.Header)
	}
	if len(req.URL.RawQuery) > 0 {
		e.AppendValue(ietfQuery, "?"+req.URL.RawQuery)
	}

	return e, nil
}

func (e *Material) addContentDigest(body []byte, headers http.Header) {
	data := sha512.Sum512(body)
	hash := "sha-512=:" + base64.StdEncoding.EncodeToString(data[:]) + ":"
	headers.Set(ContentDigestHeader, hash)
	e.AppendValue(ietfContentDigest, hash)
}

func (e *Material) AppendHeaders(headers http.Header) error {
	for k, v := range headers {
		if err := e.AppendArray(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (e *Material) AppendArray(k string, v []string) error {
	if k == SignatureHeader || k == SignatureInputHeader {
		return nil
	}
	nk, nv, err := Normalise(k, v)
	if err != nil {
		return errors.Wrap(err, "normalisation error")
	}
	for i := 0; i < len(nk); i++ {
		e.AppendValue(nk[i], nv[i])
	}
	return nil
}

func (e *Material) AppendValue(k, v string) {
	for _, prefix := range ignoreHeadersWithPrefix {
		if strings.HasPrefix(k, prefix) {
			return
		}
	}
	e.Data[k] = v
	e.Names = append(e.Names, k)
}

func (e *Material) CompleteWithSourceBody(postBodyData ...string) {
	buf := new(bytes.Buffer)
	for _, s := range e.Names {
		buf.WriteString(Format(s, e.Data[s]))
		buf.WriteByte('\n')
	}
	if e.Body != nil {
		buf.Write(e.Body)
	}
	if len(postBodyData) > 0 {
		if e.Body != nil {
			buf.WriteByte('\n')
		}
		buf.WriteString(Format(postBodyData[0], postBodyData[1]))
	}
	e.Body = buf.Bytes()
}
