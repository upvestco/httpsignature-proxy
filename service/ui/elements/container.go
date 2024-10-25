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
	"github.com/upvestco/httpsignature-proxy/service/ui/console"
	"github.com/upvestco/httpsignature-proxy/service/ui/window"
)

func NewContainer(areaTransformer window.AreaTransformer) *Container {
	e := &Container{}
	e.InitView(areaTransformer)

	return e
}

type Container struct {
	window.View
}

func (e *Container) Draw(c *console.Console) {
	area := e.GetArea()
	for i := 0; i < area.Height(); i++ {
		c.PrintRepeatWithAttributes(area.X1, area.Y1+i, ' ', area.Width(), e.GetColor().FG, e.GetColor().BG)
	}
}