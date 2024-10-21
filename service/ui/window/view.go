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

var _ Drawable = &View{}

type View struct {
	color       Color
	transformer AreaTransformer
	parent      Drawable
	children    []Drawable

	visible bool
	focus   bool
	enabled bool
}

func (e *View) OnResize() {
}

func (e *View) GetColor() Color {
	return e.color
}
func (e *View) SetColor(color Color) {
	e.color = color
}

func (e *View) Children() []Drawable {
	return e.children
}

func (e *View) SetParent(parent Drawable) {
	e.parent = parent
}

func (e *View) Add(child Drawable) {
	child.SetParent(e)
	e.children = append(e.children, child)
}

func (e *View) GetArea() Area {

	return e.transformer(e.parent.GetArea())
}

func (e *View) InitView(transformer AreaTransformer) {
	e.visible = true
	e.enabled = true
	e.focus = false
	e.transformer = transformer
}

func (e *View) SetTransformer(transformer AreaTransformer) {
	e.transformer = transformer
}

func (e *View) SetVisible() {
	e.visible = true
	for _, el := range e.Children() {
		el.SetVisible()
	}
}
func (e *View) SetHidden() {
	e.visible = false
	for _, el := range e.Children() {
		el.SetHidden()
	}
}

func (e *View) IsVisible() bool {
	return e.visible
}

func (e *View) Draw(_ *console.Console) {

}

func (e *View) ReceiveFocus() {
	e.focus = true
}
func (e *View) LostFocus() {
	e.focus = false
}

func (e *View) HasFocus() bool {
	return e.focus
}

func (e *View) Disabled() {
	e.enabled = false
	for _, el := range e.Children() {
		el.Disabled()
	}
}

func (e *View) Enabled() {
	e.enabled = true
	for _, el := range e.Children() {
		el.Enabled()
	}
}
func (e *View) IsEnabled() bool {
	return e.enabled
}

func (e *View) OnEvent(_ tb.Event) {

}
