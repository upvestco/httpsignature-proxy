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

	"github.com/nsf/termbox-go"
	"golang.design/x/clipboard"
)

type Console struct {
	fg termbox.Attribute
	bg termbox.Attribute
}

func (c *Console) Size() (int, int) {
	return termbox.Size()
}

func Create(fg, bg termbox.Attribute) *Console {

	err := clipboard.Init()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	if err := termbox.Init(); err != nil {
		fmt.Println(err)
		panic(err)
	}
	termbox.SetInputMode(termbox.InputMouse)

	c := &Console{
		fg: fg,
		bg: bg,
	}
	c.Clear()
	return c
}

func (c *Console) Background() termbox.Attribute {
	return c.bg
}
func (c *Console) Foreground() termbox.Attribute {
	return c.fg
}

func (c *Console) Update() {
	_ = termbox.Flush()
}

func (c *Console) Clear() {
	_ = termbox.Clear(c.fg, c.bg)
}

func (c *Console) Close() {
	termbox.Close()
}

func (c *Console) PrintStringWithAttributes(x, y int, s string, bg, fg termbox.Attribute) {
	for i, r := range s {
		c.SetCharWithAttributes(x+i, y, r, fg, bg)
	}
}

func (c *Console) PrintRepeatWithAttributes(x, y int, s rune, n int, fg, bg termbox.Attribute) {
	for i := 0; i < n; i++ {
		c.SetCharWithAttributes(x+i, y, s, fg, bg)
	}
}

func (c *Console) SetCharWithAttributes(x, y int, r rune, fg, bg termbox.Attribute) {
	termbox.SetCell(x, y, r, fg, bg)
}

func (c *Console) PrintString(x, y int, s string) {
	for i, r := range s {
		termbox.SetChar(x+i, y, r)
	}
}

func (c *Console) GetChar(x, y int) rune {
	w, h := c.Size()
	if x >= w || y >= h {
		return 0
	}
	return termbox.GetCell(x, y).Ch
}
