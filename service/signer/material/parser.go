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
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type vType int

/*
This parser is based on https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-message-signatures-06 */

const (
	vTypeInnerList vType = iota + 1
	vTypeList
	vTypeValue
	vTypeMap
)

type parser func(string) (interface{}, error)

var (
	ErrWrongValueSymbol     = errors.New("wrong value symbol")
	ErrImbalancedQuotes     = errors.New("imbalanced quotes")
	ErrWrongKeySymbol       = errors.New("wrong key symbol")
	ErrUnknownMaterialValue = errors.New("unknown material value")
)

var parsers = map[vType]parser{
	vTypeInnerList: parseInnerList,
	vTypeList:      parseList,
	vTypeMap:       parseMap,
	vTypeValue:     parseList,
}

func parseHeaderValue(src string) (vType, interface{}, error) {
	vt := getValueType(src, 0)
	v, err := parsers[vt](src)
	return vt, v, err
}

func getValueType(src string, p int) vType {
	if src[p] == '(' {
		return vTypeInnerList
	}
	if src[p] == '"' {
		return vTypeList
	}
	for e := p + 1; e < len(src); e++ {
		if src[e] == '=' { // nolint
			return vTypeMap
		} else if src[e] == ',' {
			return vTypeList
		} else if !allowedForKey(src[e]) {
			return vTypeValue
		}
	}
	return vTypeValue
}

func Normalise(k string, v []string) ([]string, []string, error) {
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
		list, ok := obj.([]string)
		if !ok {
			return nil, nil, ErrUnknownMaterialValue
		}
		keyList = append(keyList, nk)
		valueList = append(valueList, list[0])
	case vTypeList:
		list, ok := obj.([]string)
		if !ok {
			return nil, nil, ErrUnknownMaterialValue
		}
		keyList = append(keyList, nk)
		valueList = append(valueList, strings.Join(list, ", "))
	case vTypeInnerList:
		list, ok := obj.([]string)
		if !ok {
			return nil, nil, ErrUnknownMaterialValue
		}
		keyList = append(keyList, nk+":0")
		valueList = append(valueList, "()")
		for i := range list {
			keyList = append(keyList, fmt.Sprintf(nk+":%d", i+1))
			valueList = append(valueList, fmt.Sprintf("(%s)", strings.Join(list[0:i+1], ", ")))
		}
	case vTypeMap:
		list, ok := obj.(map[string]string)
		if !ok {
			return nil, nil, ErrUnknownMaterialValue
		}
		for k, v := range list {
			keyList = append(keyList, fmt.Sprintf(nk+":%s", k))
			valueList = append(valueList, fmt.Sprintf("%v", v))
		}
	}
	return keyList, valueList, nil
}
