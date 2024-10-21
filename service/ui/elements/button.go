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

	tb "github.com/nsf/termbox-go"
)

var noOp = func() {
}

type Button struct {
	window.View
	pressed       bool
	text          string
	pressedStyle  window.ButtonStyle
	releasedStyle window.ButtonStyle
	customStyle   tb.Attribute
	onPress       func()
	onRelease     func()
}

func NewButton(mw *window.Window, text string, area window.AreaTransformer) *Button {
	e := &Button{
		onRelease:     noOp,
		onPress:       noOp,
		text:          text,
		pressedStyle:  mw.DefaultButtonStyle(true),
		releasedStyle: mw.DefaultButtonStyle(false),
	}
	e.InitView(area)
	return e
}

func (e *Button) SetCustomStyle(customStyle tb.Attribute) {
	e.customStyle = customStyle
}

func (e *Button) SetInactive() {
	e.LostFocus()
	e.Release()
}

func (e *Button) OnPress(onPress func()) *Button {
	e.onPress = onPress
	return e
}
func (e *Button) OnRelease(onRelease func()) *Button {
	e.onRelease = onRelease
	return e
}

func (e *Button) IsPressed() bool {
	return e.pressed
}

func (e *Button) SetPressedStyle(style window.ButtonStyle) {
	e.pressedStyle = style
}
func (e *Button) SetReleasedStyle(style window.ButtonStyle) {
	e.releasedStyle = style
}
func (e *Button) Press() {
	e.pressed = true
	e.onPress()
}
func (e *Button) Release() {
	e.pressed = false
	e.onRelease()
}

func (e *Button) SetText(text string) {
	e.text = text
}

func (e *Button) GetText() string {
	return e.text
}

func (e *Button) OnEvent(event tb.Event) {
	switch event.Type {
	case tb.EventMouse:
		switch event.Key {
		case tb.MouseLeft:
			e.Press()
		case tb.MouseRelease:
			e.Release()
		default:
		}

	default:
	}
}

func (e *Button) Draw(c *console.Console) {
	printButton(c, e.GetArea(), e.pressed, e.pressedStyle, e.releasedStyle, e.text, e.customStyle)
}

func printButton(c *console.Console, area window.Area, pressed bool, prStyle, relStyle window.ButtonStyle, text string, customStyle tb.Attribute) {
	style := relStyle
	if pressed {
		style = prStyle
	}
	if style.FrameStyle != nil {
		printFrame(c, area, *style.FrameStyle, *style.FrameColor)
	}
	if len(text) > 0 {
		width := area.Width()
		x := area.X1
		y := area.Y1 + area.Height()/2
		if style.FrameStyle != nil {
			width -= 2
			x++
		}
		if width <= 0 {
			return
		}
		s := text
		if len(s) > width {
			s = s[:width]
		} else {
			pads := (width - len(s)) / 2
			s = strings.Repeat(" ", pads) + s
		}
		c.PrintStringWithAttributes(x, y, s, style.TextColor.BG, style.TextColor.FG|style.TextAttribute|customStyle)
	}
}
