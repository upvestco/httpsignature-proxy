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

func CreateCards(area AreaTransformer) *Cards {
	e := &Cards{
		cards: map[string]Drawable{},
	}
	e.InitView(area)
	return e
}

type Cards struct {
	View
	cards   map[string]Drawable
	current Drawable
}

func (e *Cards) Children() []Drawable {
	if e.current != nil {
		return []Drawable{e.current}
	}
	return nil
}

func (e *Cards) Insert(id string, c Drawable) {
	e.updateStatuses(c)
	c.Disabled()
	c.SetParent(e)
	e.cards[id] = c
	if e.current == nil {
		e.current = c
	}
}
func (e *Cards) Remove(id string) {
	delete(e.cards, id)
}

func (e *Cards) updateStatuses(c Drawable) {
	if e.IsVisible() {
		c.SetVisible()
	} else {
		c.SetHidden()
	}
	if e.HasFocus() {
		c.ReceiveFocus()
	} else {
		c.LostFocus()
	}
}

func (e *Cards) BringUp(id string) {
	active := e.cards[id]
	if active != nil {
		e.current.SetHidden()
		e.current.LostFocus()
		e.current.Disabled()

		e.updateStatuses(active)
		active.Enabled()
		e.current = active
	}
}

func (e *Cards) OnEvent(event tb.Event) {
	if !e.View.IsEnabled() {
		return
	}
	if e.current != nil {
		e.current.OnEvent(event)
	}
}

func (e *Cards) ReceiveFocus() {
	e.View.ReceiveFocus()
	if e.current != nil {
		e.current.ReceiveFocus()
	}
}

func (e *Cards) LostFocus() {
	e.View.LostFocus()
	if e.current != nil {
		e.current.LostFocus()
	}
}

func (e *Cards) Draw(c *console.Console) {
	if e.current != nil {
		if e.current.IsVisible() {
			e.current.Draw(c)
		}
	}
}

func (e *Cards) SetVisible() {
	e.View.SetVisible()
	if e.current != nil {
		e.current.SetVisible()
	}
}

func (e *Cards) SetHidden() {
	e.View.SetHidden()
	if e.current != nil {
		e.current.SetHidden()
	}
}

func (e *Cards) Disabled() {
	e.View.Disabled()
	if e.current != nil {
		e.current.Disabled()
	}
}

func (e *Cards) Enabled() {
	e.View.Enabled()
	if e.current != nil {
		e.current.Enabled()
	}
}
