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
	"bytes"
	"encoding/json"
	"strings"

	"github.com/upvestco/httpsignature-proxy/service/ui/console"
	"github.com/upvestco/httpsignature-proxy/service/ui/window"

	tb "github.com/nsf/termbox-go"
)

var (
	lineEnd      = "\n"
	openTags     = []byte{'{', '['}
	closeTags    = []byte{'}', ']'}
	collapseSign = ".."
)

type ColorJSONScheme struct {
	KeyColor          tb.Attribute
	KeyObjectColor    tb.Attribute
	ValueColor        tb.Attribute
	TagColor          tb.Attribute
	ColonColor        tb.Attribute
	CollapseSignColor tb.Attribute
}

type JSONView struct {
	window.View
	root      *Value
	lines     []string
	maxLength int
	top       int
	left      int
	marker    int

	jsonColors ColorJSONScheme
	hlColor    tb.Attribute
}

func DefaultColorJSONScheme() ColorJSONScheme {
	return ColorJSONScheme{
		KeyColor:          tb.ColorLightBlue,
		KeyObjectColor:    tb.ColorBlue,
		ValueColor:        tb.ColorYellow,
		TagColor:          tb.ColorCyan,
		ColonColor:        tb.ColorWhite,
		CollapseSignColor: tb.ColorLightRed | tb.AttrBold,
	}
}

func (e *JSONView) SetHighlightColor(hlColor tb.Attribute) {
	e.hlColor = hlColor
}
func NewJSONView(areaTransformer window.AreaTransformer) *JSONView {
	e := &JSONView{
		hlColor:    tb.ColorCyan,
		jsonColors: DefaultColorJSONScheme(),
	}
	e.InitView(areaTransformer)
	return e
}

func (e *JSONView) Set(obj interface{}) {
	data, _ := json.MarshalIndent(obj, "", "  ")
	e.updateData(data)
	e.root = createTree(data)
}

func (e *JSONView) updateData(text []byte) {
	maxLength := 0
	var lines []string
	for _, l := range toLines(text) {
		if m := len(l); m > maxLength {
			maxLength = m
		}
		lines = append(lines, l)
	}
	e.lines = lines
	e.maxLength = maxLength
	if e.top+e.marker > len(e.lines) {
		e.top = 0
		e.marker = 0
	}

}
func (e *JSONView) WithColorJSONScheme(jsonColors ColorJSONScheme) *JSONView {
	e.jsonColors = jsonColors
	return e
}

func (e *JSONView) OnEvent(event tb.Event) {
	switch event.Key {
	case tb.KeyArrowUp:
		e.ScrollUp()
	case tb.KeyArrowDown:
		e.ScrollDown()
	case tb.KeyArrowLeft:
		e.ScrollRight()
	case tb.KeyArrowRight:
		e.ScrollLeft()
	case tb.KeyHome:
		e.Home()
	case tb.KeySpace:
		e.OnSpace()
	case tb.MouseLeft:
		marker := event.MouseY - e.GetArea().Y1
		if marker+e.top > len(e.lines) {
			marker = len(e.lines) - e.top - 1
		}
		e.marker = marker
	case tb.MouseWheelUp:
		e.ScrollDown()
	case tb.MouseWheelDown:
		e.ScrollUp()
	default:

	}
}

func (e *JSONView) Colours(colors ColorJSONScheme) *JSONView {
	e.jsonColors = colors
	return e
}

func (e *JSONView) Home() {
	e.top = 0
	e.left = 0
	e.marker = 0
}

func (e *JSONView) OnSpace() {
	child := e.root.GetByRow(e.row())
	if child != nil {
		if child.Collapsed() {
			e.updateData(child.Inflate(e.lines))
		} else {
			e.updateData(child.Collapse(e.lines))
		}
	}
}

func (e *JSONView) ScrollLeft() {
	child := e.root.GetByRow(e.row())
	if child != nil && child.Collapsed() {
		e.updateData(child.Inflate(e.lines))
		return
	}
	area := e.GetArea()
	if e.left < e.maxLength-(area.Width()) {
		e.left++
	}
}

func (e *JSONView) ScrollRight() {
	child := e.root.GetByRow(e.row())
	if child != nil && !child.Collapsed() {
		e.updateData(child.Collapse(e.lines))
		return
	}
	if e.left > 0 {
		e.left--
	}
}

func (e *JSONView) ScrollUp() {
	if e.marker == 0 {
		if e.top == 0 {
			return
		} else {
			e.top--
		}
	} else {
		e.marker--
	}
}

func (e *JSONView) ScrollDown() {
	area := e.GetArea()

	if e.marker == area.Height()-1 {
		if e.top+area.Height() >= len(e.lines) {
			return
		} else {
			e.top++
		}
	} else {
		if e.row() < len(e.lines)-1 {
			e.marker++
		}
	}
}

func (e *JSONView) Draw(c *console.Console) {
	area := e.GetArea()
	if !area.Valid() {
		return
	}
	color := e.GetColor()

	empty := strings.Repeat(" ", area.Width())
	for y := 0; y < area.Height(); y++ {
		if y+e.top >= len(e.lines) {
			c.PrintString(area.X1, area.Y1+y, empty)
			continue
		}
		s := e.lines[y+e.top]
		fgs := e.highlight(s)

		if len(s) > e.left {
			s = s[e.left:]
			fgs = fgs[e.left:]
		} else {
			s = ""
		}

		if l := len(s); l > area.Width() {
			s = s[:area.Width()]
			fgs = fgs[:area.Width()]
		} else {
			s += strings.Repeat(" ", area.Width()-l)
			for i := 0; i < area.Width()-l; i++ {
				fgs = append(fgs, 0)
			}
		}
		bg := color.BG
		if y == e.marker && e.HasFocus() {
			bg = e.hlColor
		}
		e.PrintLine(c, area.X1, area.Y1+y, s, fgs, bg)
	}
}

func (e *JSONView) PrintLine(c *console.Console, x, y int, s string, fgs []tb.Attribute, bg tb.Attribute) {
	for i, r := range s {
		c.SetCharWithAttributes(x+i, y, r, fgs[i], bg)
	}
}

func (e *JSONView) highlight(s string) []tb.Attribute {
	type cs struct {
		start  int
		finish int
	}
	var css []cs

	fgs := make([]tb.Attribute, len(s))
	hasColon := false
	inCommas := false
	objectStart := false
	var start int
	var finish int
	for i := 0; i < len(s); i++ {
		fgs[i] = e.jsonColors.ValueColor
		c := s[i]
		switch {
		case c == '\\':
			i++
			fgs[i] = e.jsonColors.ValueColor
		case isOpenTag(c) || isCloseTag(c):
			fgs[i] = e.jsonColors.TagColor
			if isOpenTag(c) {
				objectStart = true
			}
		case c == ':' && !inCommas:
			fgs[i] = e.jsonColors.ColonColor
			hasColon = true
		case c == '"':
			inCommas = !inCommas
			if inCommas {
				start = i
			}
			if !inCommas {
				finish = i
				css = append(css, cs{
					start:  start,
					finish: finish,
				})
			}
		}
	}
	if hasColon && len(css) > 0 {
		cs := css[0]
		for i := cs.start; i <= cs.finish; i++ {
			if objectStart {
				fgs[i] = e.jsonColors.KeyObjectColor
			} else {
				fgs[i] = e.jsonColors.KeyColor
			}
		}
	}
	if p := strings.Index(s, "["+collapseSign+"]"); p >= 0 {
		for i := 0; i < len(collapseSign)+2; i++ {
			fgs[i+p] = e.jsonColors.CollapseSignColor
		}
	}
	if p := strings.Index(s, "{"+collapseSign+"}"); p >= 0 {
		for i := 0; i < len(collapseSign)+2; i++ {
			fgs[i+p] = e.jsonColors.CollapseSignColor
		}
	}

	return fgs
}

func (e *JSONView) row() int {
	return e.marker + e.top
}

func isOpenTag(c byte) bool {
	for i := 0; i < len(openTags); i++ {
		if openTags[i] == c {
			return true
		}
	}
	return false
}
func isCloseTag(c byte) bool {
	for i := 0; i < len(closeTags); i++ {
		if closeTags[i] == c {
			return true
		}
	}
	return false
}

func toBytes(lines []string) []byte {
	data := bytes.Buffer{}
	for i, l := range lines {
		data.WriteString(l)
		if i < len(lines)-1 {
			data.WriteString(lineEnd)
		}
	}
	return data.Bytes()
}

func toLines(data []byte) []string {
	return strings.Split(string(data), lineEnd)
}

func collect(pieces ...[]byte) []byte {
	b := bytes.Buffer{}
	for _, p := range pieces {
		b.Write(p)
	}
	return b.Bytes()
}

type position struct {
	row   int
	col   int
	index int
}
type Value struct {
	start    position
	end      position
	children []*Value
	parent   *Value
	hidden   bool
	extract  []byte
}

func (e *Value) Collapsed() bool {
	return e.extract != nil
}

func (e *Value) GetByRow(row int) *Value {
	if e.hidden {
		return nil
	}
	if e.start.row == row {
		return e
	}
	for _, child := range e.children {
		f := child.GetByRow(row)
		if f != nil {
			return f
		}
	}
	return nil
}

func (e *Value) Inflate(lines []string) []byte {
	current := e

	rows := len(toLines(current.extract)) - 1
	pos := len(current.extract) - len(collapseSign)
	current.adjust(-rows, -pos, false)

	data := toBytes(lines)
	newData := collect(data[:current.start.index+1], current.extract, data[current.start.index+1+len(collapseSign):])
	current.extract = nil
	return newData
}

func (e *Value) Collapse(lines []string) []byte {
	current := e

	data := toBytes(lines)
	current.extract = data[current.start.index+1 : current.end.index]
	newData := collect(data[:current.start.index+1], []byte(collapseSign), data[current.end.index:])

	rows := len(toLines(current.extract)) - 1
	pos := len(current.extract) - len(collapseSign)
	current.adjust(rows, pos, true)

	return newData

}
func (e *Value) adjust(rows, pos int, hidden bool) {
	current := e

	current.end.row -= rows
	current.end.index -= pos
	for _, child := range current.children {
		child.adjustRows(rows, pos, hidden)
	}
	parent := current.parent
	lowerTheMe := false
	for parent != nil {
		for _, child := range parent.children {
			if child == current {
				lowerTheMe = true
				continue
			}
			if lowerTheMe {
				child.adjustRows(rows, pos, child.hidden)
			}
		}
		parent.end.row -= rows
		parent.end.index -= pos
		current = parent
		parent = current.parent
		lowerTheMe = false
	}
}

func (e *Value) adjustRows(rows, pos int, hidden bool) {
	e.start.row -= rows
	e.end.row -= rows
	e.start.index -= pos
	e.end.index -= pos
	e.hidden = hidden
	for _, child := range e.children {
		child.adjustRows(rows, pos, hidden)
	}
}
func createTree(data []byte) *Value {
	col := 0
	row := 0
	inCommas := false
	root := &Value{
		start: position{
			row:   0,
			col:   0,
			index: 0,
		},
	}
	current := root

	for i := 1; i < len(data); i++ {
		c := data[i]
		col++
		switch {
		case c == '\\': // scip the next symbol
			col++
			i++
		case isOpenTag(c) && !inCommas:
			child := &Value{
				start: position{
					row:   row,
					col:   col,
					index: i,
				},
				parent: current,
			}
			current.children = append(current.children, child)
			current = child
		case isCloseTag(c) && !inCommas:
			current.end = position{
				row:   row,
				col:   col,
				index: i,
			}
			current = current.parent
		case c == lineEnd[0]:
			col = 0
			row++
		case c == '"':
			inCommas = !inCommas
		}
	}
	return root
}
