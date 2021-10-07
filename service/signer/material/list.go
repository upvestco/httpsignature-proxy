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

func parseList(src string) (interface{}, error) { //nolint
	list := make([]string, 0)
	quoted := false
	cv := make([]byte, 0)
	for i := 0; i < len(src); i++ {
		s := src[i]
		if quoted { //nolint
			if s == '"' {
				quoted = false
				cv = append(cv, s)
				list = append(list, string(cv))
				cv = cv[:0]
				i += 1
				if endOfTheItem(src, i) {
					i += 1
				}
				continue
			}
			cv = append(cv, s)
			if s == '\\' {
				i++
				if i < len(src) {
					cv = append(cv, src[i])
				} else {
					return nil, errors.WithMessage(ErrImbalancedQuotes, src)
				}
			}
		} else {
			if s == '"' {
				quoted = true
				cv = append(cv, s)
				continue
			} else {
				if endOfTheItem(src, i) {
					list = append(list, string(cv))
					cv = cv[:0]
					i = getNextSpacePosition(src, i+1) - 1
					continue
				}
				if !(allowedForValue(s)) {
					return nil, errors.WithMessage(ErrWrongValueSymbol, src)
				}
				cv = append(cv, src[i])
			}
		}
	}
	if len(cv) > 0 {
		if quoted {
			return nil, errors.WithMessage(ErrImbalancedQuotes, src)
		}
		list = append(list, string(cv))
	}
	return list, nil
}
