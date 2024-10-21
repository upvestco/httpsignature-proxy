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
	"strings"

	"github.com/upvestco/httpsignature-proxy/service/ui/console"
	"github.com/upvestco/httpsignature-proxy/service/ui/window"
)

func NewTextView(areaTransformer window.AreaTransformer, text string) *TextView {
	e := &TextView{
		text: text,
	}
	e.InitView(areaTransformer)
	return e
}

type TextView struct {
	window.View
	text string
}

func (e *TextView) Draw(c *console.Console) {
	area := e.GetArea()
	lines := strings.Split(e.text, "\n")
	blank := strings.Repeat(" ", area.Width())
	for i := 0; i < area.Height(); i++ {
		s := blank
		if i < len(lines) {
			s = lines[i]
			if len(s) > area.Width() {
				s = s[:area.Width()]
			} else {
				s += strings.Repeat(" ", area.Width()-len(s))
			}
		}
		c.PrintStringWithAttributes(area.X1, area.Y1+i, s, e.GetColor().BG, e.GetColor().FG)
	}
}

func (e *TextView) SetText(text string) {
	e.text = text
}

func (e *TextView) GetText() string {
	return e.text
}
