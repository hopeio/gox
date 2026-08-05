package geom

import (
	"golang.org/x/exp/constraints"
	"math"
)

type Circle struct {
	Centre   Point
	Diameter float64
}

// Bounds ...
func (c *Circle) Bounds() *Bounds {
	return NewBounds(c.Centre.X, c.Centre.Y, c.Diameter, c.Diameter)
}

type CircleInt[T constraints.Integer] struct {
	Center   PointInt[T]
	Diameter T
}

// ToFloat64 ...
func (e *CircleInt[T]) ToFloat64(factor float64) *Circle {
	if factor == 0 {
		factor = 1
	}
	return &Circle{
		Centre:   Point{float64(e.Center.X) / factor, float64(e.Center.Y) / factor},
		Diameter: float64(e.Diameter) / factor,
	}
}

// Overlap ...
func (c *Circle) Overlap(c2 Circle) bool {
	// 计算两个圆心之间的距离
	distance := math.Sqrt(math.Pow(c2.Centre.X-c.Centre.X, 2) + math.Pow(c2.Centre.Y-c.Centre.Y, 2))

	// 判断距离是否小于等于两个圆的半径之和
	return distance <= (c.Diameter/2 + c2.Diameter/2)
}
