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
	"golang.org/x/exp/utf8string"

	tb "github.com/nsf/termbox-go"
)

type SelectViewItem interface {
	String() string
}

type viewItem struct {
	item    SelectViewItem
	visited bool
}

type SelectView struct {
	window.View
	lines      []*viewItem
	maxLength  int
	top        int
	left       int
	notVisited int

	marker int

	hlColor      tb.Attribute
	visitedColor tb.Attribute
	onSelect     func(interface{})
	onChange     func()
	foregrounds  func(selected, visited bool, s string) ([]tb.Attribute, tb.Attribute)
}

func NewSelectView(areaTransformer window.AreaTransformer) *SelectView {
	e := &SelectView{}
	e.InitSelectView(areaTransformer, e.defaultForegrounds)
	return e
}

func (e *SelectView) InitSelectView(areaTransformer window.AreaTransformer, foregrounds func(bool, bool, string) ([]tb.Attribute, tb.Attribute)) {
	e.onSelect = func(item interface{}) {

	}
	e.onChange = func() {

	}
	e.hlColor = tb.ColorCyan
	e.visitedColor = tb.ColorDefault
	e.marker = -1
	e.InitView(areaTransformer)
	if foregrounds == nil {
		foregrounds = e.defaultForegrounds
	}
	e.foregrounds = foregrounds
}

func (e *SelectView) SetVisitedColor(color tb.Attribute) {
	e.visitedColor = color
}
func (e *SelectView) GetSize() int {
	return len(e.lines)
}
func (e *SelectView) GetNotVisited() int {
	return e.notVisited
}

func (e *SelectView) ResetMarker() {
	e.marker = -1
	e.top = 0
}

func (e *SelectView) SetHighlightColor(hlColor tb.Attribute) {
	e.hlColor = hlColor
}

func (e *SelectView) add(content []SelectViewItem) {
	maxLength := e.maxLength
	lines := e.lines
	for _, l := range content {
		if m := len(l.String()); m > maxLength {
			maxLength = m
		}
		lines = append(lines, &viewItem{
			item:    l,
			visited: false,
		})
	}
	e.notVisited += len(content)
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
		e.ScrollDown(false)
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

func (e *SelectView) onMouseLeft(y int) {
	marker := y - e.GetArea().Y1
	if marker+e.top > len(e.lines) {
		marker = len(e.lines) - e.top - 1
	}
	e.marker = marker
	e.selectItem(true)
}

func (e *SelectView) OnEnd() {
	e.left = 0
	height := e.GetArea().Height()
	if len(e.lines) > height {
		e.top = len(e.lines) - height
		e.marker = height - 1
	} else {
		e.marker = len(e.lines) - 1
	}
	e.selectItem(true)

}

func (e *SelectView) Home() {
	e.top = 0
	e.left = 0
	e.marker = 0
	e.selectItem(true)
}

func (e *SelectView) ScrollRight() {
	if e.left > 0 {
		e.left--
	}
}

func (e *SelectView) ScrollUp(visit bool) {
	if e.marker == 0 {
		if e.top == 0 {
			return
		} else {
			e.top--
		}
	} else {
		e.marker--
	}
	e.selectItem(visit)

}
func (e *SelectView) ScrollDown(visit bool) {
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
	e.selectItem(visit)
}

func (e *SelectView) selectItem(visit bool) {
	r := e.row()
	if r >= 0 && r < len(e.lines) {
		item := e.lines[r]
		if visit && !item.visited {
			e.notVisited--
			item.visited = true
		}
		e.onSelect(item.item)
	}
}

func (e *SelectView) MarkAllVisited() {
	for i := 0; i < len(e.lines); i++ {
		e.lines[i].visited = true
	}
	e.notVisited = 0
}

func (e *SelectView) row() int {
	return e.top + e.marker
}

func (e *SelectView) OnEvent(event tb.Event) {
	switch event.Key {
	case tb.KeyArrowUp:
		e.ScrollUp(true)
	case tb.KeyArrowDown:
		e.ScrollDown(true)
	case tb.KeyArrowLeft:
		e.ScrollRight()
	case tb.KeyArrowRight:
		e.ScrollLeft()
	case tb.KeyHome:
		e.Home()
	case tb.KeyEnd:
		e.OnEnd()
	case tb.MouseLeft:
		e.onMouseLeft(event.MouseY)
	case tb.MouseWheelUp:
		e.ScrollDown(false)
	case tb.MouseWheelDown:
		e.ScrollUp(false)
	default:

	}
}

func (e *SelectView) GetItems() []SelectViewItem {
	var r []SelectViewItem
	for _, item := range e.lines {
		r = append(r, item.item)
	}
	return r
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
		s := lines.item.String()
		fgs, bg := e.foregrounds(y == e.marker, lines.visited, s)
		if len(s) > e.left {
			s = s[e.left:]
			fgs = fgs[e.left:]
		} else {
			s = ""
			fgs = fgs[:0]
		}
		if len(s) > width {
			s = s[:width]
			fgs = fgs[:width]
		} else {
			s += strings.Repeat(" ", width-len(s))
		}
		for len(fgs) < len(s) {
			fgs = append(fgs, bg)
		}
		e.printLine(c, area.X1, area.Y1+y, s, fgs, bg)
	}
}

func (e *SelectView) defaultForegrounds(selected, visited bool, s string) ([]tb.Attribute, tb.Attribute) {
	color := e.GetColor()
	bg := color.BG
	m := tb.Attribute(0)
	if selected {
		if e.HasFocus() {
			bg = e.hlColor
		} else {
			m = tb.AttrBold
		}
	}
	fgs := make([]tb.Attribute, len(s))
	fg := color.FG
	if visited {
		fg = e.visitedColor
	}
	for i := 0; i < len(s); i++ {
		fgs[i] = fg | m
	}
	return fgs, bg
}

func (e *SelectView) printLine(c *console.Console, x, y int, s string, fgs []tb.Attribute, bg tb.Attribute) {
	rs := utf8string.NewString(s)
	for i := 0; i < rs.RuneCount(); i++ {
		c.SetCharWithAttributes(x+i, y, rs.At(i), fgs[i], bg)
	}
}
