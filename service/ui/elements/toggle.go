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
	"github.com/upvestco/httpsignature-proxy/service/ui/window"

	tb "github.com/nsf/termbox-go"
)

type Toggle struct {
	Button
}

func NewToggle(mw *window.Window, text string, areaTransformer window.AreaTransformer) *Toggle {
	e := &Toggle{}
	e.onPress = noOp
	e.onRelease = noOp
	e.pressedStyle = mw.DefaultButtonStyle(true)
	e.releasedStyle = mw.DefaultButtonStyle(false)
	e.text = text
	e.Enabled()
	e.InitView(areaTransformer)
	return e
}

func (e *Toggle) OnEvent(event tb.Event) {
	switch event.Type {
	case tb.EventMouse:
		switch event.Key {
		case tb.MouseLeft:
			e.pressed = !e.pressed
			if e.pressed {
				e.Press()
			} else {
				e.Release()
			}
		default:
		}
	case tb.EventKey:
		switch event.Key {
		case tb.KeySpace:
			e.pressed = !e.pressed
			if e.pressed {
				e.Press()
			} else {
				e.Release()
			}
		default:
		}
	default:
	}
}
