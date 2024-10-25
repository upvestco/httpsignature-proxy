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
	"github.com/upvestco/httpsignature-proxy/service/ui/console"

	tb "github.com/nsf/termbox-go"
)

type FrameStyle struct {
	Top       []rune
	Middle    []rune
	Intersect []rune
	Bottom    []rune
}

// NormalFrameStyle
// uses https://en.wikipedia.org/wiki/Box-drawing_characters
var NormalFrameStyle = FrameStyle{
	Top:       []rune{'┌', '┬', '┐'},
	Intersect: []rune{'├', '┼', '┤'},
	Bottom:    []rune{'└', '┴', '┘'},
	Middle:    []rune{'│', '─'},
}

var DoubleFrameStyle = FrameStyle{
	Top:       []rune{'╔', '╦', '╗'},
	Intersect: []rune{'╠', '╬', '╣'},
	Bottom:    []rune{'╚', '╩', '╝'},
	Middle:    []rune{'║', '═'},
}

var BoldFrameStyle = FrameStyle{
	Top:       []rune{'┏', '┳', '┓'},
	Intersect: []rune{'┣', '╋', '┫'},
	Bottom:    []rune{'┗', '┻', '┛'},
	Middle:    []rune{'┃', '━'},
}

type AreaTransformer func(Area) Area

type Drawable interface {
	Draw(c *console.Console)
	OnEvent(event tb.Event)
	OnResize()
	GetArea() Area

	Add(child Drawable) // do not use these 3 function directly.
	SetParent(Drawable)
	Children() []Drawable

	SetVisible() // show/hide element
	SetHidden()
	IsVisible() bool

	ReceiveFocus() // tell element that it's got focus - keyboard events goes to this element
	LostFocus()
	HasFocus() bool

	Disabled() // tell element that it's disabled and no longer receives events from mause or keyboard
	Enabled()
	IsEnabled() bool
}

type ButtonStyle struct {
	TextColor     *Color
	FrameColor    *Color
	FrameStyle    *FrameStyle
	TextAttribute tb.Attribute
}
