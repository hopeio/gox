package geom

import "math"

type Ellipse struct {
	Centre    Point
	MajorAxis float64 // Length of the major axis
	MinorAxis float64 // Length of the minor axis
	Angle     float64 // Angle angle in degrees from the x-axis
}

// Area returns the result.
func (e *Ellipse) Area() float64 {
	majorHalf := e.MajorAxis / 2
	minorHalf := e.MinorAxis / 2
	return math.Pi * majorHalf * minorHalf
}

// Perimeter returns the result.
func (e *Ellipse) Perimeter() float64 {
	majorHalf := e.MajorAxis / 2
	minorHalf := e.MinorAxis / 2
	h := math.Pow(majorHalf-minorHalf, 2) / math.Pow(majorHalf+minorHalf, 2)
	return math.Pi * (majorHalf + minorHalf) * (1 + (3*h)/(10+math.Sqrt(4-3*h)))
}

// GetHalfAxes returns the half lengths of the major and minor axes.
func (e *Ellipse) GetHalfAxes() (float64, float64) {
	return e.MajorAxis / 2, e.MinorAxis / 2
}
