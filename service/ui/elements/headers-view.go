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

package elements

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/upvestco/httpsignature-proxy/service/ui/window"

	tb "github.com/nsf/termbox-go"
)

type HeadersView struct {
	SelectView
	keyColor   tb.Attribute
	valueColor tb.Attribute
}

func NewHeadersView(areaTransformer window.AreaTransformer) *HeadersView {
	e := &HeadersView{}
	e.InitSelectView(areaTransformer, e.Foregrounds)
	return e
}

func (e *HeadersView) SetKeyColor(color tb.Attribute) {
	e.keyColor = color
}

func (e *HeadersView) SetValueColor(color tb.Attribute) {
	e.valueColor = color
}

func (e *HeadersView) SetHeaders(headers http.Header) {
	lines := FormatHeaders(headers)
	items := make([]SelectViewItem, len(headers))
	for i, line := range lines {
		items[i] = &headerView{line: line}
	}
	e.Set(items)
}

func FormatHeaders(headers http.Header) []string {
	var keys []string
	maxL := 0
	for key := range headers {
		if l := len(key); l > maxL {
			maxL = l
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	items := make([]string, len(headers))
	for i, key := range keys {
		values := headers[key]
		s := fmt.Sprintf("%s%s : %s", strings.Repeat(" ", maxL-len(key)), key, strings.Join(values, ","))
		items[i] = s
	}
	return items
}

func (e *HeadersView) Foregrounds(selected, _ bool, s string) ([]tb.Attribute, tb.Attribute) {
	var fgs []tb.Attribute
	color := e.GetColor()
	p := strings.Index(s, ":")
	m := tb.Attribute(0)
	if selected {
		m = tb.AttrBold
	}
	for i := 0; i < len(s); i++ {
		if i < p {
			fgs = append(fgs, e.keyColor|m)
		} else {
			fgs = append(fgs, e.valueColor|m)
		}
	}
	return fgs, color.BG
}

type headerView struct {
	line string
}

func (h headerView) String() string {
	return h.line
}
