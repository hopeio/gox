package geom

import (
	"golang.org/x/exp/constraints"
	"math"
	"math/rand/v2"
)

// Point represents a point
type Point struct {
	X float64
	Y float64
}

// Pt returns the result.
func Pt(x, y float64) Point {
	return Point{X: x, Y: y}
}

// RandomPoint returns the result.
func RandomPoint(min, max Point) Point {
	return Point{
		X: math.Floor(min.X + math.Floor(rand.Float64()*(max.X-min.X))),
		Y: math.Floor(min.Y + math.Floor(rand.Float64()*(max.Y-min.Y))),
	}
}

// Vector returns the result.
func (p Point) Vector(point Point) Vector {
	return Vector{point.X - p.X, point.Y - p.Y}
}

// Rotate returns the result.
func (p Point) Rotate(center Point, angleDeg float64) Point {
	angleRad := angleDeg * math.Pi / 180.0
	// Calculate cosine and sine of the angle
	cosA := math.Cos(angleRad)
	sinA := math.Sin(angleRad)
	// Compute rotated coordinates
	newX := center.X + (p.X-center.X)*cosA - (p.Y-center.Y)*sinA
	newY := center.Y + (p.X-center.X)*sinA + (p.Y-center.Y)*cosA

	return Point{newX, newY}
}

// Length returns the result.
func (p Point) Length(p2 Point) float64 {
	dx := p2.X - p.X
	dy := p2.Y - p.Y
	return math.Hypot(dx, dy)
}

type Point3D struct {
	X float64
	Y float64
	Z float64
}

// PointInt represents an integer point
type PointInt[T constraints.Integer] struct {
	X T
	Y T
}

// ToFloat64 converts the value.
func (l *PointInt[T]) ToFloat64(factor float64) *Point {
	return &Point{
		X: float64(l.X) / factor,
		Y: float64(l.Y) / factor,
	}
}

type Point3DInt[T constraints.Integer] struct {
	X T
	Y T
	Z T
}

// ToFloat64 converts the value.
func (l *Point3DInt[T]) ToFloat64(factor float64) *Point3D {
	return &Point3D{
		X: float64(l.X) / factor,
		Y: float64(l.Y) / factor,
		Z: float64(l.Z) / factor,
	}
}

// Mirror returns the result.
func (p Point) Mirror(line *GeneralFormLine) Point {
	denominator := line.A*line.A + line.B*line.B
	if denominator == 0 {
		panic("Invalid line equation: the line cannot be vertical and horizontal at the same time.")
	}
	factor := (line.A*p.X + line.B*p.Y + line.C) / denominator
	return Point{
		X: p.X - 2*line.A*factor,
		Y: p.Y - 2*line.B*factor,
	}
}
