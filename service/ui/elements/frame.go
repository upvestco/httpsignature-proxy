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

func NewFrame(areaTransformer window.AreaTransformer, style window.FrameStyle) *Frame {
	e := &Frame{
		style: style,
	}
	e.InitView(areaTransformer)
	return e
}

type Frame struct {
	window.View
	style window.FrameStyle
}

func (e *Frame) Draw(c *console.Console) {
	printFrame(c, e.GetArea(), e.style, e.GetColor())
}

func printFrame(c *console.Console, area window.Area, style window.FrameStyle, color window.Color) {
	c.SetCharWithAttributes(area.X1, area.Y1, style.Top[0], color.FG, color.BG)
	c.SetCharWithAttributes(area.X2, area.Y1, style.Top[2], color.FG, color.BG)
	c.SetCharWithAttributes(area.X1, area.Y2, style.Bottom[0], color.FG, color.BG)
	c.SetCharWithAttributes(area.X2, area.Y2, style.Bottom[2], color.FG, color.BG)

	for i := 1; i < area.Width()-1; i++ {
		r := style.Middle[1]
		if c.GetChar(area.X1+i, area.Y1) == style.Middle[0] {
			r = style.Top[1]
		}
		c.SetCharWithAttributes(area.X1+i, area.Y1, r, color.FG, color.BG)

		r = style.Middle[1]
		if c.GetChar(area.X1+i, area.Y2) == style.Middle[0] {
			r = style.Bottom[1]
		}
		c.SetCharWithAttributes(area.X1+i, area.Y2, r, color.FG, color.BG)
	}

	for i := 1; i < area.Height()-1; i++ {
		r := style.Middle[0]
		if c.GetChar(area.X1, i) == style.Middle[1] {
			r = style.Intersect[0]
		}
		c.SetCharWithAttributes(area.X1, area.Y1+i, r, color.FG, color.BG)

		r = style.Middle[0]
		if c.GetChar(area.X2, i) == style.Middle[1] {
			r = style.Intersect[2]
		}
		c.SetCharWithAttributes(area.X2, area.Y1+i, r, color.FG, color.BG)
	}

}
