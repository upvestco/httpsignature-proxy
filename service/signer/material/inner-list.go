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
	"github.com/pkg/errors"
)

func parseInnerList(src string) (interface{}, error) {
	cv, err := extractInnerList(src)
	if err != nil {
		return nil, errors.Wrap(err, "fail to extract inner lst")
	}
	return parseList(string(cv))
}

func extractInnerList(src string) ([]byte, error) {
	quoted := false
	cv := make([]byte, 0)

	for i := 1; i < len(src); i++ {
		s := src[i]
		if quoted { //nolint
			if s == '"' {
				quoted = false
			}
			if s == '\\' {
				cv = append(cv, s)
				if !endOfTheItem(src, i) {
					i = i + 1
					s = src[i]
				}
			}
		} else {
			if s == '"' {
				quoted = true
			}
			if s == ')' {
				return cv, nil
			}
		}
		cv = append(cv, s)
	}
	return nil, errors.WithMessage(ErrWrongValueSymbol, src)
}
