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

package window

import (
	tb "github.com/nsf/termbox-go"
	"github.com/upvestco/httpsignature-proxy/service/ui/console"
)

type Window struct {
	View
	C             *console.Console
	toUpdate      chan Drawable
	width, height int
}

func New(fg, bg tb.Attribute) *Window {
	e := &Window{
		C:        console.Create(fg, bg),
		toUpdate: make(chan Drawable, 10),
	}
	e.InitView(WholeArea())
	e.View.parent = e

	return e
}

func (e *Window) GetArea() Area {
	return Area{
		X1: 0,
		Y1: 0,
		X2: e.width - 1,
		Y2: e.height - 1,
	}
}

func (e *Window) Close() {
	e.C.Close()
}

func (e *Window) Update(list ...Drawable) {
	for _, el := range list {
		select {
		case e.toUpdate <- el:
		default:
		}
	}
}

func (e *Window) DrawAll() {
	e.C.Clear()
	e.C.Update()
}
func (e *Window) Size() (int, int) {
	return e.C.Size()
}

func (e *Window) Run(exitKey tb.Key) {

	e.width, e.height = e.C.Size()
	active := getSelected(e, 0, 0)
	if active != nil {
		active.ReceiveFocus()
	}
	e.repaint(e)
	e.C.Update()
	eventCh := make(chan tb.Event, 10)

	go func() {
		for {
			eventCh <- tb.PollEvent()
		}
	}()

	for {
		select {
		case r := <-e.toUpdate:
			e.repaint(r)
			e.C.Update()
		case ev := <-eventCh:

			switch ev.Type {
			case tb.EventKey:
				switch ev.Key {
				case exitKey:
					return
				default:
				}
				if active != nil {
					active.OnEvent(ev)
					active.Draw(e.C)
				}
			case tb.EventMouse:
				switch ev.Key {
				case tb.MouseLeft:
					selected := getSelected(e, ev.MouseX, ev.MouseY)
					if selected == nil {
						if active != nil {
							active.LostFocus()
							active.Draw(e.C)
						}
						active = nil
					} else if selected != active {
						if active != nil {
							active.LostFocus()
							active.Draw(e.C)
						}
						selected.ReceiveFocus()
						active = selected
					}
				default:
				}
				if active != nil {
					active.OnEvent(ev)
					active.Draw(e.C)
				}
			case tb.EventResize:
				e.height = ev.Height
				e.width = ev.Width
				e.resize(e)
				e.C.Clear()
				e.repaint(e)

			default:

			}
			e.C.Update()
		}
	}
}

func getSelected(v Drawable, x, y int) Drawable {
	if len(v.Children()) == 0 {
		return v
	}
	for _, r := range v.Children() {
		if r.IsEnabled() && r.IsVisible() {
			area := r.GetArea()
			if area.Inside(x, y) {
				return getSelected(r, x, y)
			}
		}
	}

	return nil
}

func (e *Window) resize(v Drawable) {
	v.OnResize()
	for _, el := range v.Children() {
		e.resize(el)
	}
}

func (e *Window) repaint(v Drawable) {
	if !v.IsVisible() {
		return
	}
	v.Draw(e.C)
	for _, el := range v.Children() {
		if el.IsVisible() {
			e.repaint(el)
		}
	}
}

type Color struct {
	FG tb.Attribute
	BG tb.Attribute
}

func NewColor(fg tb.Attribute, bg tb.Attribute) Color {
	return Color{
		FG: fg,
		BG: bg,
	}
}

func (e *Window) DefaultColor() Color {
	return Color{
		BG: e.C.Background(),
		FG: e.C.Foreground(),
	}
}

func (e *Window) DefaultButtonStyle(pressed bool) ButtonStyle {
	dc := e.DefaultColor()
	style := ButtonStyle{
		TextColor:  &dc,
		FrameColor: &dc,
	}
	if pressed {
		style.TextAttribute = tb.AttrBold
	} else {
		style.TextAttribute = tb.Attribute(0)
	}
	return style
}
