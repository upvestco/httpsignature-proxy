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

package console

import (
	"fmt"

	tb "github.com/nsf/termbox-go"
)

type Console struct {
	fg tb.Attribute
	bg tb.Attribute
}

func (c *Console) Size() (int, int) {
	return tb.Size()
}

func Create(fg, bg tb.Attribute) *Console {

	if err := tb.Init(); err != nil {
		fmt.Println(err)
		panic(err)
	}
	tb.SetInputMode(tb.InputMouse)

	c := &Console{
		fg: fg,
		bg: bg,
	}
	c.Clear()
	return c
}

func (c *Console) Background() tb.Attribute {
	return c.bg
}
func (c *Console) Foreground() tb.Attribute {
	return c.fg
}

func (c *Console) Update() {
	_ = tb.Flush()
}

func (c *Console) Clear() {
	_ = tb.Clear(c.fg, c.bg)
}

func (c *Console) Close() {
	tb.Close()
}

func (c *Console) PrintStringWithAttributes(x, y int, s string, bg, fg tb.Attribute) {
	for i, r := range s {
		c.SetCharWithAttributes(x+i, y, r, fg, bg)
	}
}

func (c *Console) PrintRepeatWithAttributes(x, y int, s rune, n int, fg, bg tb.Attribute) {
	for i := 0; i < n; i++ {
		c.SetCharWithAttributes(x+i, y, s, fg, bg)
	}
}

func (c *Console) SetCharWithAttributes(x, y int, r rune, fg, bg tb.Attribute) {
	tb.SetCell(x, y, r, fg, bg)
}

func (c *Console) PrintString(x, y int, s string) {
	for i, r := range s {
		tb.SetChar(x+i, y, r)
	}
}

func (c *Console) GetChar(x, y int) rune {
	w, h := c.Size()
	if x >= w || y >= h {
		return 0
	}
	return tb.GetCell(x, y).Ch
}
