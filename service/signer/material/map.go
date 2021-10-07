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

	"github.com/pkg/errors"
)

func parseMap(src string) (interface{}, error) {
	res := map[string]string{}
	for i := 0; i < len(src); {
		key, err := getKey(src, i)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get key from %s", src)
		}
		i = i + len(key) + 1
		var v []byte
		if src[i] == '"' { //nolint
			v, err = extractQuotedInnerValue(src[i:])
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get quoted inner value from %s", src[i:])
			}
		} else if src[i] == '(' {
			v, err = extractInnerList(src[i:])
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get inner list  from %s", src[i:])
			}
			v = []byte(fmt.Sprintf("(%s)", string(v)))
		} else {
			v, err = extractInnerValue(src[i:])
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get inner list  from %s", src[i:])
			}
		}
		i += len(v)
		res[key] = string(v)
		if endOfTheItem(src, i) {
			i = getNextSpacePosition(src, i+1)
		} else {
			i++
		}

	}
	return res, nil
}

func getKey(src string, sp int) (string, error) {
	key := make([]byte, 0)
	for i := sp; i < len(src); i++ {
		s := src[i]
		if allowedForKey(s) { // nolint
			key = append(key, s)
		} else if s == '=' {
			return string(key), nil
		} else {
			return "", errors.WithMessage(ErrWrongKeySymbol, src[sp:])
		}
	}
	return "", errors.WithMessage(ErrWrongKeySymbol, src[sp:])
}

func extractQuotedInnerValue(src string) ([]byte, error) {
	cv := make([]byte, 0)
	cv = append(cv, src[0])

	for i := 1; i < len(src); i++ {
		s := src[i]
		if s == '"' {
			cv = append(cv, s)
			return cv, nil
		}
		cv = append(cv, s)
		if s == '\\' {
			if !endOfTheItem(src, i) {
				cv = append(cv, src[i+1])
				i++
			} else {
				return nil, ErrWrongValueSymbol
			}
		}
	}
	return nil, ErrWrongValueSymbol
}

func extractInnerValue(src string) ([]byte, error) {
	cv := make([]byte, 0)

	for i := 0; i < len(src); i++ {
		s := src[i]
		if endOfTheItem(src, i) {
			return cv, nil
		}
		if !allowedForValue(s) {
			return nil, ErrWrongValueSymbol
		}
		cv = append(cv, s)
	}
	return cv, nil
}

func allowedForKey(n byte) bool {
	return n == '_' || n == '-' || n == '.' || n == '*' || (n >= 'a' && n <= 'z') || (n >= '0' && n <= '9')
}

func allowedForValue(n byte) bool {
	return n >= 32 && n <= 0x7f
}

func getNextSpacePosition(src string, start int) int {
	var k int
	for k = start; k < len(src); k++ {
		if src[k] != ' ' {
			break
		}
	}
	return k
}

func endOfTheItem(src string, i int) bool {
	return i+1 < len(src) && src[i] == ',' && src[i+1] == ' '
}
