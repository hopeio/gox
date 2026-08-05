package geom

import (
	"golang.org/x/exp/constraints"
	"math"
)

type Vector struct {
	X float64
	Y float64
}

// NewVector creates and returns a new instance.
func NewVector(p1, p2 Point) Vector {
	return Vector{p2.X - p1.X, p2.Y - p1.Y}
}

// Angle ...
func (v Vector) Angle() float64 {
	angleRadians := math.Atan2(v.Y, v.X)
	return angleRadians * (180.0 / math.Pi)
}

// Length ...
func (v Vector) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// AngleWith ...
func (v Vector) AngleWith(v2 Vector) float64 {
	dotProduct := v.X*v2.X + v.Y*v2.Y
	magnitudeV1 := math.Sqrt(v.X*v.X + v.Y*v.Y)
	magnitudeV2 := math.Sqrt(v2.X*v2.X + v2.Y*v2.Y)
	angleInRadians := math.Acos(dotProduct / (magnitudeV1 * magnitudeV2))
	return angleInRadians * (180.0 / math.Pi)
}

// CrossProduct ...
func (v Vector) CrossProduct(v2 Vector) float64 {
	return v.X*v2.Y - v.Y*v2.X
}

type VectorInt[T constraints.Integer] struct {
	X T
	Y T
}

// ToFloat64 ...
func (l *VectorInt[T]) ToFloat64(factor float64) *Vector {
	return &Vector{
		X: float64(l.X) / factor,
		Y: float64(l.Y) / factor,
	}
}
