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

type Area struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
}

type Point struct {
	X int
	Y int
}

func NewPoint(x, y func() int) func() Point {
	return func() Point {
		return Point{
			X: x(),
			Y: y(),
		}
	}
}

func Rectangle(topLeft func() Point, rightBottom func() Point) Area {
	tl := topLeft()
	rb := rightBottom()
	return NewArea(tl.X, tl.Y, rb.X, rb.Y)
}

func HLine(x1, x2, y func() int) Area {
	return NewArea(x1(), y(), x2(), y())
}
func VLine(y1, y2, x func() int) Area {
	return NewArea(x(), y1(), x(), y2())
}

func (a Area) TopLeft() func() Point {
	return func() Point {
		return Point{
			X: a.X1,
			Y: a.Y1,
		}
	}
}
func (a Area) BottomRight() func() Point {
	return func() Point {
		return Point{
			X: a.X2,
			Y: a.Y2,
		}
	}
}

func (a Area) Left(x int) func() int {
	return func() int {
		return a.X1 + x
	}
}
func (a Area) OnLeft() func() int {
	return func() int {
		return a.X1
	}
}

func (a Area) Right(x int) func() int {
	return func() int {
		return a.X2 - x
	}
}
func (a Area) OnRight() func() int {
	return func() int {
		return a.X2
	}
}

func (a Area) Top(y int) func() int {
	return func() int {
		return a.Y1 + y
	}
}
func (a Area) OnTop() func() int {
	return func() int {
		return a.Y1
	}
}

func (a Area) OnBottom() func() int {
	return func() int {
		return a.Y2
	}
}

func (a Area) Bottom(y int) func() int {
	return func() int {
		return a.Y2 - y
	}
}

func (a Area) Add(x1, y1, x2, y2 int) Area {
	return NewArea(a.X1+x1, a.Y1+y1, a.X2+x2, a.Y2+y2)
}

func (a Area) Valid() bool {
	return a.X1 >= 0 &&
		a.Y1 >= 0 &&
		a.X2 >= 0 &&
		a.Y2 >= 0 &&
		a.Width() > 0 && a.Height() > 0
}

func ZeroArea() Area {
	return NewArea(0, 0, 0, 0)
}
func WholeArea() func(parent Area) Area {
	return func(parent Area) Area {
		return parent.Copy()
	}
}

func (a Area) Copy() Area {
	return NewArea(a.X1, a.Y1, a.X2, a.Y2)

}
func NewArea(x1, y1, x2, y2 int) Area {
	return Area{
		X1: x1,
		Y1: y1,
		X2: x2,
		Y2: y2,
	}
}

func (a Area) Width() int {
	return a.X2 - a.X1 + 1
}
func (a Area) Height() int {
	return a.Y2 - a.Y1 + 1
}
func (a Area) Inside(x, y int) bool {
	return x >= a.X1 && x <= a.X2 &&
		y >= a.Y1 && y <= a.Y2
}
func ShrinkAreaTransformer(dx, dy int) func(parent Area) Area {
	return func(a Area) Area {
		return a.Add(dx, dy, -dx, -dy)
	}
}
