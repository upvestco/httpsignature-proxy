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

type VLine struct {
	window.View
	style window.FrameStyle
}

func NewVLine(areaTransformer window.AreaTransformer) *VLine {
	e := &VLine{
		style: window.NormalFrameStyle,
	}
	e.InitView(areaTransformer)

	return e
}

func (e *VLine) Draw(c *console.Console) {
	printVerticalLine(c, e.GetArea(), e.GetColor(), e.style)
}

func printVerticalLine(c *console.Console, area window.Area, color window.Color, style window.FrameStyle) {
	x := area.X1
	y := area.Y1
	h := area.Height()
	if h <= 0 {
		return
	}
	if c.GetChar(x, y) == style.Middle[1] {
		c.SetCharWithAttributes(x, y, style.Top[1], color.FG, color.BG)
	} else {
		c.SetCharWithAttributes(x, y, style.Middle[0], color.FG, color.BG)
	}
	for i := y + 1; i < h-1; i++ {
		if c.GetChar(x, i) == style.Middle[1] {
			c.SetCharWithAttributes(x, i, style.Intersect[1], color.FG, color.BG)
		} else {
			c.SetCharWithAttributes(x, i, style.Middle[0], color.FG, color.BG)
		}
	}
	if c.GetChar(x, y+h-1) == style.Middle[1] {
		c.SetCharWithAttributes(x, y+h-1, style.Bottom[1], color.FG, color.BG)
	} else {
		c.SetCharWithAttributes(x, y+h-1, style.Middle[0], color.FG, color.BG)
	}

}
