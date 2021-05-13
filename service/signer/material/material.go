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
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrWrongValueSymbol = errors.New("wrong value symbol")
	ErrImbalancedQuotes = errors.New("imbalanced quotes")
	ErrWrongKeySymbol   = errors.New("wrong key symbol")

	SignatureInput = textproto.CanonicalMIMEHeaderKey("Signature-Input")
	Signature      = textproto.CanonicalMIMEHeaderKey("Signature")
)

func NewMaterial() *Material {
	return &Material{
		list:  make(map[string]string),
		Names: make([]string, 0),
	}
}

type Material struct {
	list    map[string]string
	Names   []string
	Created string
	Data    []byte
}

func (e *Material) AppendHeaders(headers http.Header) error {
	for k, v := range headers {
		if err := e.Append(k, v); err != nil {
			return err
		}
	}
	return nil
}
func (e *Material) Append(k string, v []string) error {
	if k == Signature || k == SignatureInput {
		return nil
	}
	nk, nv, err := e.normalize(k, v)
	if err != nil {
		return errors.Wrap(err, "normalization error")
	}
	for i := 0; i < len(nk); i++ {
		e.Append0(nk[i], nv[i])
	}
	return nil
}

func (e *Material) Append0(k, v string) {
	e.list[k] = v
	e.Names = append(e.Names, k)

}
func (e *Material) PrepareWithBodyForSign(body []byte) {
	e.Created = fmt.Sprintf("%d", time.Now().Unix())
	e.Append0("*created", e.Created)
	buf := new(bytes.Buffer)
	for _, s := range e.Names {
		buf.WriteString(e.format(s, e.list[s]))
		buf.WriteByte('\n')
	}
	buf.Write(body)
	e.Data = buf.Bytes()
}

func (e *Material) format(k, v string) string {
	return fmt.Sprintf("%s: %s", k, v)
}

func (e *Material) normalize(k string, v []string) ([]string, []string, error) {
	keyList := make([]string, 0)
	valueList := make([]string, 0)
	nk := strings.TrimSpace(strings.ToLower(k))
	for i := range nk {
		if !allowedForKey(nk[i]) {
			return nil, nil, errors.Wrap(ErrWrongKeySymbol, nk)
		}
	}
	trimmed := make([]string, len(v))
	for i := range v {
		trimmed[i] = strings.TrimSpace(v[i])
	}
	nv := strings.Join(trimmed, ", ")
	if len(nv) == 0 {
		return []string{nk}, []string{""}, nil
	}
	vt, obj, err := parseHeaderValue(nv)
	if err != nil {
		return nil, nil, errors.Wrap(err, nv)
	}
	switch vt {
	case vTypeValue:
		list := obj.([]string)
		keyList = append(keyList, nk)
		valueList = append(valueList, list[0])
	case vTypeList:
		list := obj.([]string)
		keyList = append(keyList, nk)
		valueList = append(valueList, strings.Join(list, ", "))
	case vTypeInnerList:
		list := obj.([]string)
		keyList = append(keyList, nk+":0")
		valueList = append(valueList, "()")
		for i := range list {
			keyList = append(keyList, fmt.Sprintf(nk+":%d", i+1))
			valueList = append(valueList, fmt.Sprintf("(%s)", strings.Join(list[0:i+1], ", ")))
		}
	case vTypeMap:
		list := obj.(map[string]string)
		for k, v := range list {
			keyList = append(keyList, fmt.Sprintf(nk+":%s", k))
			valueList = append(valueList, fmt.Sprintf("%v", v))
		}
	}
	return keyList, valueList, nil
}

func endOfTheItem(src string, i int) bool {
	return i+1 < len(src) && src[i] == ',' && src[i+1] == ' '
}

func allowedForKey(n byte) bool {
	return n == '_' || n == '-' || n == '.' || n == '*' || (n >= 'a' && n <= 'z') || (n >= '0' && n <= '9')
}
func allowedForValue(n byte) bool {
	return n >= 32 && n <= 0x7f
}

func scipSpaces(src string, start int) int {
	var k int
	for k = start; k < len(src); k++ {
		if src[k] != ' ' {
			break
		}
	}
	return k
}
