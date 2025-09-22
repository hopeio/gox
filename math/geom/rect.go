/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package geom

import (
	mathx "github.com/hopeio/gox/math"
	"golang.org/x/exp/constraints"
	"image"
	"math"
)

type Rectangle struct {
	Center Point
	Width  float64
	Height float64
	Angle  float64
}

func NewRect(center Point, width, height float64, angleDeg float64) *Rectangle {
	return &Rectangle{
		Center: center,
		Width:  width,
		Height: height,
		Angle:  angleDeg,
	}
}

func RectNoRotate(x0, y0, x1, y1 float64) *Rectangle {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	return &Rectangle{
		Center: Point{(x0 + x1) / 2, (y0 + y1) / 2},
		Width:  x1 - x0,
		Height: y1 - y0,
	}
}

func RectNoRotate2(ltp, rdp Point) *Rectangle {
	return &Rectangle{
		Center: Point{(ltp.X + rdp.X) / 2, (ltp.Y + rdp.Y) / 2},
		Width:  rdp.X - ltp.X,
		Height: rdp.Y - ltp.Y,
	}
}

func RectFromImageRect(r image.Rectangle) *Rectangle {
	return RectNoRotate(float64(r.Min.X), float64(r.Min.Y), float64(r.Max.X), float64(r.Max.Y))
}

func RectFromCorners(points [4]Point) *Rectangle {
	var center Point
	for _, p := range points {
		center.X += p.X
		center.Y += p.Y
	}
	center.X /= 4
	center.Y /= 4
	width := points[0].Length(points[0])
	height := points[1].Length(points[2])
	dx := points[1].X - points[0].X
	dy := points[1].Y - points[0].Y
	angle := math.Atan2(dy, dx) * 180 / math.Pi // 转为度数
	return &Rectangle{
		Center: center,
		Width:  width,
		Height: height,
		Angle:  angle,
	}
}

func (rect *Rectangle) Bounds() *Bounds {
	if rect.Angle == 0 {
		return NewBounds(rect.Center.X-rect.Width/2, rect.Center.Y-rect.Height/2, rect.Center.X+rect.Width/2, rect.Center.Y+rect.Height/2)
	}
	corners := rect.Corners()
	minx, maxx := mathx.MinAndMax(corners[0].X, corners[1].X, corners[2].X, corners[3].X)
	miny, maxy := mathx.MinAndMax(corners[0].Y, corners[1].Y, corners[2].Y, corners[3].Y)
	return NewBounds(minx, miny, maxx, maxy)
}

func (rect *Rectangle) Corners() [4]Point {
	if rect.Angle == 0 {
		return [4]Point{{rect.Center.X - rect.Width/2, rect.Center.Y - rect.Height/2},
			{rect.Center.X + rect.Width/2, rect.Center.Y - rect.Height/2},
			{rect.Center.X + rect.Width/2, rect.Center.Y + rect.Height/2},
			{rect.Center.X - rect.Width/2, rect.Center.Y + rect.Height/2}}
	}
	angleRad := rect.Angle * math.Pi / 180.0
	// Calculate cosine and sine of the angle
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)
	halfW, halfH := rect.Width/2, rect.Height/2
	// 计算矩形四个角的坐标 (A左下-B右下-C右上-D左上)
	dx := rect.Center.X + halfW*cosA - halfH*sinA
	dy := rect.Center.Y + halfW*sinA + halfH*cosA
	ax := rect.Center.X - halfW*cosA - halfH*sinA
	ay := rect.Center.Y - halfW*sinA + halfH*cosA
	bx := rect.Center.X - halfW*cosA + halfH*sinA
	by := rect.Center.Y - halfW*sinA - halfH*cosA
	cx := rect.Center.X + halfW*cosA + halfH*sinA
	cy := rect.Center.Y + halfW*sinA - halfH*cosA
	return [4]Point{{ax, ay}, {bx, by}, {cx, cy}, {dx, dy}}
}

// 图片就是第四象限,角度90+θ
func (rect *Rectangle) ContainsPoint(p Point) bool {

	// 射线法判断点是否在矩形内
	inside := false
	intersections := 0
	corners := rect.Corners()

	for i := 0; i < len(corners); i++ {
		x1, y1 := corners[i].X, corners[i].Y
		x2, y2 := corners[(i+1)%len(corners)].X, corners[(i+1)%len(corners)].Y

		if y1 == y2 { // 水平边
			continue
		}
		if p.Y < min(y1, y2) || p.Y > max(y1, y2) { // 在边的外部
			continue
		}

		xIntersect := x1 + (p.Y-y1)*(x2-x1)/(y2-y1)
		if p.X < xIntersect {
			intersections++
		}
	}

	if intersections%2 == 1 {
		inside = true
	}

	return inside
}

type RectangleInt[T constraints.Integer] struct {
	Center PointInt[T]
	Width  T
	Height T
	Angle  float64
}

func (rect *RectangleInt[T]) ToFloat64(factor float64) *Rectangle {
	if factor == 0 {
		factor = 1
	}
	return &Rectangle{
		Center: Point{float64(rect.Center.X) / factor, float64(rect.Center.Y) / factor},
		Width:  float64(rect.Width) / factor,
		Height: float64(rect.Height) / factor,
		Angle:  rect.Angle,
	}
}

func NewRectInt[T constraints.Integer](center PointInt[T], width, height T, angle float64) *RectangleInt[T] {
	return &RectangleInt[T]{center, width, height, angle}
}

func RectIntFromFloat64[T constraints.Integer](e *Rectangle, factor float64) *RectangleInt[T] {
	if factor == 0 {
		factor = 1
	}
	return &RectangleInt[T]{
		Center: PointInt[T]{
			X: T(math.Round(e.Center.X * factor)),
			Y: T(math.Round(e.Center.Y * factor)),
		},
		Width:  T(math.Round(e.Width * factor)),
		Height: T(math.Round(e.Angle * factor)),
		Angle:  e.Angle,
	}
}

type Bounds struct {
	Min Point
	Max Point
}

func (b *Bounds) ToRect() *Rectangle {
	return RectNoRotate((b.Min.X+b.Max.X)/2, (b.Min.Y+b.Max.Y)/2, b.Max.X-b.Min.X, b.Max.Y-b.Min.Y)
}

func NewBounds(x0, y0, x1, y1 float64) *Bounds {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	return &Bounds{
		Min: Point{X: x0, Y: y0},
		Max: Point{x1, y1},
	}
}

func BoundsFromImageRect(r image.Rectangle) *Bounds {
	return NewBounds(float64(r.Min.X), float64(r.Min.Y), float64(r.Max.X), float64(r.Max.Y))
}
