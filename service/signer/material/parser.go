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

type vType int

/*
This parser is based on https://tools.ietf.org/id/draft-ietf-httpbis-message-signatures-01.html */

const (
	vTypeInnerList vType = iota + 1
	vTypeList
	vTypeValue
	vTypeMap
)

type parser func(string) (interface{}, error)

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
		if src[e] == '=' {
			return vTypeMap
		} else if src[e] == ',' {
			return vTypeList
		} else if !allowedForKey(src[e]) {
			return vTypeValue
		}
	}
	return vTypeValue
}
