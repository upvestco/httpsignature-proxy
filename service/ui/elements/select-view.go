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

	"github.com/nsf/termbox-go"
)

type SelectViewItem interface {
	String() string
}

type SelectView struct {
	window.View
	lines     []SelectViewItem
	maxLength int
	top       int
	left      int

	marker int

	hlColor     termbox.Attribute
	onSelect    func(interface{})
	onChange    func()
	foregrounds func(selected bool, s string) ([]termbox.Attribute, termbox.Attribute)
}

func NewSelectView(areaTransformer window.AreaTransformer) *SelectView {
	e := &SelectView{}
	e.InitSelectView(areaTransformer, e.defaultForegrounds)
	return e
}

func (e *SelectView) InitSelectView(areaTransformer window.AreaTransformer, foregrounds func(selected bool, s string) ([]termbox.Attribute, termbox.Attribute)) {
	e.onSelect = func(item interface{}) {

	}
	e.onChange = func() {

	}
	e.hlColor = termbox.ColorCyan
	e.marker = -1
	e.InitView(areaTransformer)
	if foregrounds == nil {
		foregrounds = e.defaultForegrounds
	}
	e.foregrounds = foregrounds
}
func (e *SelectView) GetLength() int {
	return len(e.lines)
}

func (e *SelectView) ResetMarker() {
	e.marker = -1
	e.top = 0
}

func (e *SelectView) SetHighlightColor(hlColor termbox.Attribute) {
	e.hlColor = hlColor
}

func (e *SelectView) add(content []SelectViewItem) {
	maxLength := e.maxLength
	lines := e.lines
	for _, l := range content {
		if m := len(l.String()); m > maxLength {
			maxLength = m
		}
		lines = append(lines, l)
	}
	e.lines = lines
	e.maxLength = maxLength
	e.onChange()
}

func (e *SelectView) OnSelect(onSelect func(l interface{})) *SelectView {
	e.onSelect = onSelect
	return e
}

func (e *SelectView) OnChange(onChange func()) *SelectView {
	e.onChange = onChange
	return e
}

func (e *SelectView) Append(content SelectViewItem) {
	n := len(e.lines)
	e.add([]SelectViewItem{content})
	if n == 0 && len(e.lines) > 0 && e.HasFocus() {
		e.onSelect(e.lines[0])
	}
	lastLine := n-1 == e.row()
	if lastLine {
		e.ScrollDown()
	}
}

func (e *SelectView) Set(content []SelectViewItem) {
	e.left = 0
	e.lines = e.lines[:0]
	e.maxLength = 0
	e.add(content)
	if e.top+e.marker > len(e.lines) {
		e.top = 0
		e.marker = 0
	}
}

func (e *SelectView) ScrollLeft() {
	area := e.GetArea()
	if e.left < e.maxLength-area.Width() {
		e.left++
	}
}

func (e *SelectView) Home() {
	e.top = 0
	e.left = 0
	e.marker = 0
}

func (e *SelectView) ScrollRight() {
	if e.left > 0 {
		e.left--
	}
}

func (e *SelectView) ScrollUp() {
	if e.marker == 0 {
		if e.top == 0 {
			return
		} else {
			e.top--
		}
	} else {
		e.marker--
	}
	e.onSelect(e.lines[e.row()])
}
func (e *SelectView) ScrollDown() {
	area := e.GetArea()
	if e.marker == area.Height()-1 {
		if e.top+area.Height() >= len(e.lines) {
			return
		} else {
			e.top++
		}
	} else {
		if e.row() == len(e.lines)-1 {
			return
		} else {
			e.marker++
		}
	}
	if len(e.lines) > e.row() {
		e.onSelect(e.lines[e.row()])
	}
}

func (e *SelectView) row() int {
	return e.top + e.marker
}

func (e *SelectView) OnEvent(event termbox.Event) {
	switch event.Key {
	case termbox.KeyArrowUp:
		e.ScrollUp()
	case termbox.KeyArrowDown:
		e.ScrollDown()
	case termbox.KeyArrowLeft:
		e.ScrollRight()
	case termbox.KeyArrowRight:
		e.ScrollLeft()
	case termbox.KeyHome:
		e.Home()
	case termbox.KeyEnd:
		e.left = 0
		height := e.GetArea().Height()
		if len(e.lines) > height {
			e.top = len(e.lines) - height
			e.marker = height - 1
		} else {
			e.marker = len(e.lines) - 1
		}
		e.onSelect(e.lines[e.row()])
	case termbox.MouseLeft:
		marker := event.MouseY - e.GetArea().Y1
		if marker+e.top > len(e.lines) {
			marker = len(e.lines) - e.top - 1
		}
		e.marker = marker
		if e.row() >= 0 && e.row() < len(e.lines) {
			e.onSelect(e.lines[e.row()])
		}
	default:

	}
}

func (e *SelectView) Draw(c *console.Console) {
	area := e.GetArea()
	width := area.Width()
	empty := strings.Repeat(" ", width)
	color := e.GetColor()
	for y := 0; y < area.Height(); y++ {
		if y+e.top >= len(e.lines) {
			c.PrintStringWithAttributes(area.X1, area.Y1+y, empty, color.BG, color.FG)
			continue
		}
		lines := e.lines[y+e.top]
		s := lines.String()
		if len(s) > e.left {
			s = s[e.left:]
		} else {
			s = ""
		}
		if len(s) > width {
			s = s[:width]
		} else {
			s += strings.Repeat(" ", width-len(s))
		}
		fgs, bg := e.foregrounds(y == e.marker, s)
		e.printLine(c, area.X1, area.Y1+y, s, fgs, bg)
	}
}

func (e *SelectView) defaultForegrounds(selected bool, s string) ([]termbox.Attribute, termbox.Attribute) {
	color := e.GetColor()
	bg := color.BG
	m := termbox.Attribute(0)
	if selected {
		if e.HasFocus() {
			bg = e.hlColor
		} else {
			m = termbox.AttrBold
		}
	}
	fgs := make([]termbox.Attribute, len(s))
	for i := 0; i < len(s); i++ {
		fgs[i] = color.FG | m
	}
	return fgs, bg
}

func (e *SelectView) printLine(c *console.Console, x, y int, s string, fgs []termbox.Attribute, bg termbox.Attribute) {
	for i, r := range s {
		c.SetCharWithAttributes(x+i, y, r, fgs[i], bg)
	}
}
