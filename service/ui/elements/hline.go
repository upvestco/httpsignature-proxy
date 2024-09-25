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
	"github.com/upvestco/httpsignature-proxy/service/ui/console"
	"github.com/upvestco/httpsignature-proxy/service/ui/window"
)

type HLine struct {
	window.View
	style window.FrameStyle
}

func NewHLine(areaTransformer window.AreaTransformer) *HLine {
	e := &HLine{
		style: window.NormalFrameStyle,
	}
	e.InitView(areaTransformer)

	return e
}

func (e *HLine) Draw(c *console.Console) {
	printHorizontalLine(c, e.GetArea(), e.GetColor(), e.style)
}

func printHorizontalLine(c *console.Console, area window.Area, color window.Color, style window.FrameStyle) {
	if !area.Valid() {
		return
	}
	x := area.X1
	y := area.Y1
	w := area.Width()

	if c.GetChar(x, y) == style.Middle[0] {
		c.SetCharWithAttributes(x, y, style.Intersect[0], color.FG, color.BG)
	} else {
		c.SetCharWithAttributes(x, y, style.Middle[1], color.FG, color.BG)
	}
	for i := x + 1; i < x+w-1; i++ {
		if c.GetChar(i, y) == style.Middle[0] {
			c.SetCharWithAttributes(i, y, style.Intersect[1], color.FG, color.BG)
		} else {
			c.SetCharWithAttributes(i, y, style.Middle[1], color.FG, color.BG)
		}
	}
	if c.GetChar(x+w-1, y) == style.Middle[0] {
		c.SetCharWithAttributes(x+w-1, y, style.Intersect[2], color.FG, color.BG)
	} else {
		c.SetCharWithAttributes(x+w-1, y, style.Middle[1], color.FG, color.BG)
	}

}
