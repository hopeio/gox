package geom

import (
	"math"

	"golang.org/x/exp/constraints"
)

const tolerance = 1e-9

type LineSegment struct {
	Start Point
	End   Point
}

// NewLineSegment creates and returns a new instance.
func NewLineSegment(start, end Point) *LineSegment {
	return &LineSegment{start, end}
}

// Vector returns the result.
func (l *LineSegment) Vector() Vector {
	return Vector{l.End.X - l.Start.X, l.End.Y - l.Start.Y}
}

// ToSlopeInterceptFormLine converts the value.
func (l *LineSegment) ToSlopeInterceptFormLine() *SlopeInterceptFormLine {
	var line SlopeInterceptFormLine
	if l.Start.X == l.End.X {
		// Vertical line
		line.Slope = math.Inf(1)   // Positive infinity to indicate vertical line
		line.Intercept = l.Start.X // The x-coordinate of the vertical line
		return &line
	}

	// Calculate slope (m)
	line.Slope = (l.End.Y - l.Start.Y) / (l.End.X - l.Start.X)

	// Calculate y-intercept (b)
	line.Intercept = l.Start.Y - line.Slope*l.Start.X
	return &line
}

// ContainsPoint reports whether the condition holds.
func (l *LineSegment) ContainsPoint(p Point) bool {
	// Ensure the point is within the bounding box of the line segment
	if math.Min(l.Start.X, l.End.X) <= p.X && p.X <= math.Max(l.Start.X, l.End.X) &&
		math.Min(l.Start.Y, l.End.Y) <= p.Y && p.Y <= math.Max(l.Start.Y, l.End.Y) {

		// Calculate the area of the triangle formed by the three points
		area := 0.5 * math.Abs(l.Start.X*l.End.Y+p.X*l.Start.Y+l.End.X*p.Y-p.X*l.End.Y-l.Start.X*p.Y-l.End.X*l.Start.Y)

		// If the area is effectively zero, the point is on the line
		return math.Abs(area) < tolerance
	}

	return false
}

// IntersectionLineSegment performs the operation.
func (l *LineSegment) IntersectionLineSegment(l2 *LineSegment) (Point, bool) {
	// Compute vectors
	dx1, dy1 := l.End.X-l.Start.X, l.End.Y-l.Start.Y
	dx2, dy2 := l2.End.X-l2.Start.X, l2.End.Y-l2.Start.Y
	dx3, dy3 := l.Start.X-l2.Start.X, l.Start.Y-l2.Start.Y

	// Compute the determinant to test parallelism
	denom := dx1*dy2 - dy1*dx2
	if math.Abs(denom) < tolerance {
		// The two lines are parallel or colinear
		return Point{}, false
	}

	// Compute parameters t and u
	t := (dx3*dy2 - dy3*dx2) / denom
	u := (dx3*dy1 - dy3*dx1) / denom

	// Check whether t and u are in [0, 1]
	if t >= 0 && t <= 1 && u >= 0 && u <= 1 {
		// Compute intersection coordinates
		intersectionX := l.Start.X + t*dx1
		intersectionY := l.Start.Y + t*dy1
		return Point{intersectionX, intersectionY}, true
	}

	// Intersection is outside both segments
	return Point{}, false
}

// IntersectionRay performs the operation.
func (l *LineSegment) IntersectionRay(r *Ray) (Point, bool) {
	// Extract segment and ray parameters
	x1, y1 := l.Start.X, l.Start.Y
	x2, y2 := l.End.X, l.End.Y
	x3, y3, theta := r.Point.X, r.Point.Y, r.Angle/180*math.Pi

	// Compute direction vectors of the segment and ray
	dx1, dy1 := x2-x1, y2-y1                     // segment direction
	dx2, dy2 := math.Cos(theta), math.Sin(theta) // ray direction

	// Compute the determinant to test parallelism
	denom := dx1*dy2 - dy1*dx2
	if math.Abs(denom) < tolerance {
		// Segment and ray are parallel or colinear
		return Point{}, false
	}

	// Compute parameters t and s
	t := ((x3-x1)*dy2 - (y3-y1)*dx2) / denom
	s := ((x3-x1)*dy1 - (y3-y1)*dx1) / denom

	// Check whether t and s are in range
	if t >= 0 && t <= 1 && s >= 0 {
		// Compute intersection coordinates
		intersectionX := x1 + t*dx1
		intersectionY := y1 + t*dy1
		return Point{intersectionX, intersectionY}, true
	}

	// Intersection is outside the segment/ray ranges
	return Point{}, false
}

// IntersectionStraightLine performs the operation.
func (l *LineSegment) IntersectionStraightLine(sl *StraightLine) (Point, bool) {
	// Extract segment and line parameters
	x1, y1 := l.Start.X, l.Start.Y
	x2, y2 := l.End.X, l.End.Y
	x3, y3, theta := sl.Point.X, sl.Point.Y, sl.Angle/180*math.Pi

	// Compute direction vectors of the segment and line
	dx1, dy1 := x2-x1, y2-y1                     // segment direction
	dx2, dy2 := math.Cos(theta), math.Sin(theta) // line direction

	// Compute the determinant to test parallelism
	denom := dx1*dy2 - dy1*dx2
	if math.Abs(denom) < 1e-9 {
		// Segment and line are parallel or colinear
		return Point{}, false
	}

	// Compute parameters t and s
	t := ((x3-x1)*dy2 - (y3-y1)*dx2) / denom

	// Check whether t is in range
	if t >= 0 && t <= 1 {
		// Compute intersection coordinates
		intersectionX := x1 + t*dx1
		intersectionY := y1 + t*dy1
		return Point{intersectionX, intersectionY}, true
	}

	// Intersection is not on the segment
	return Point{}, false
}

type LineSegmentInt[T constraints.Integer] struct {
	Start PointInt[T]
	End   PointInt[T]
}

// ToFloat64 converts the value.
func (l *LineSegmentInt[T]) ToFloat64(factor float64) *LineSegment {
	return &LineSegment{
		Start: Point{float64(l.Start.X) / factor, float64(l.Start.Y) / factor},
		End:   Point{float64(l.End.X) / factor, float64(l.End.Y) / factor},
	}
}

// LineIntFromFloat64 returns the result.
func LineIntFromFloat64[T constraints.Integer](e *LineSegment, factor float64) *LineSegmentInt[T] {
	return &LineSegmentInt[T]{
		Start: PointInt[T]{T(math.Round(e.Start.X * factor)), T(math.Round(e.Start.Y * factor))},
		End:   PointInt[T]{T(math.Round(e.End.X * factor)), T(math.Round(e.End.Y * factor))},
	}
}

// y=mx+b
type SlopeInterceptFormLine struct {
	Slope     float64
	Intercept float64
}

// IsVertical reports whether the condition holds.
func (l *SlopeInterceptFormLine) IsVertical() bool {
	return math.IsInf(l.Slope, 0)
}

// ToGeneralFormLine converts the value.
func (l *SlopeInterceptFormLine) ToGeneralFormLine() *GeneralFormLine {
	if l.IsVertical() {
		// For vertical lines: x = k, convert to Ax + By + C = 0 where A = 1, B = 0, C = -k
		k := l.Intercept
		return &GeneralFormLine{A: 1, B: 0, C: -k}
	}

	return &GeneralFormLine{A: l.Slope, B: -1, C: l.Intercept}
}

// ToStraightLine converts the value.
func (l *SlopeInterceptFormLine) ToStraightLine() *StraightLine {
	return &StraightLine{
		Point: Point{l.Intercept, 0},
		Angle: math.Atan(l.Slope) * 180 / math.Pi,
	}
}

// NewSlopeInterceptLine creates and returns a new instance.
func NewSlopeInterceptLine(m, b float64) *SlopeInterceptFormLine {
	return &SlopeInterceptFormLine{m, b}
}

// ax + by + c = 0
type GeneralFormLine struct {
	A float64
	B float64
	C float64
}

// ToSlopeInterceptLine converts the value.
func (l *GeneralFormLine) ToSlopeInterceptLine() *SlopeInterceptFormLine {
	if l.B == 0 {
		return &SlopeInterceptFormLine{math.Inf(1), -l.C}
	}
	return &SlopeInterceptFormLine{
		Slope:     -l.A / l.B,
		Intercept: l.C / l.B,
	}
}

// ToStraightLine converts the value.
func (l *GeneralFormLine) ToStraightLine() *StraightLine {
	return &StraightLine{
		Point: Point{l.C / l.B, 0},
		Angle: math.Atan(l.A/l.B) * 180 / math.Pi,
	}
}

// NewGeneralFormLine creates and returns a new instance.
func NewGeneralFormLine(a, b, c float64) *GeneralFormLine {
	return &GeneralFormLine{a, b, c}
}

type StraightLine struct {
	Point
	Angle float64
}

// NewStraightLine creates and returns a new instance.
func NewStraightLine(p Point, angle float64) *StraightLine {
	return &StraightLine{p, angle}
}

// ToGeneralFormLine converts the value.
func (l *StraightLine) ToGeneralFormLine() *GeneralFormLine {
	angleInRadians := l.Angle / 180 * math.Pi
	return &GeneralFormLine{math.Cos(angleInRadians), math.Sin(angleInRadians), -l.X*math.Cos(angleInRadians) - l.Y*math.Sin(angleInRadians)}
}

// ToSlopeInterceptLine converts the value.
func (l *StraightLine) ToSlopeInterceptLine() *SlopeInterceptFormLine {
	return l.ToGeneralFormLine().ToSlopeInterceptLine()
}

// ContainsPoint reports whether the condition holds.
func (l *StraightLine) ContainsPoint(p Point) bool {
	// Convert angle from degrees to radians
	angleInRadians := l.Angle * math.Pi / 180

	// Handler vertical line case (angle is 90 or 270 degrees)
	if math.Mod(math.Abs(l.Angle-90), 180) < tolerance || math.Mod(math.Abs(l.Angle-270), 360) < tolerance {
		return math.Abs(p.X-l.X) < tolerance
	}

	// Calculate slope m and intercept b
	m := math.Tan(angleInRadians)
	b := l.Y - m*l.X

	// Check if the point satisfies the line equation within a small tolerance

	return math.Abs(p.Y-(m*p.X+b)) < tolerance
}

// IntersectStraightLine reports whether the condition holds.
func (l *StraightLine) IntersectStraightLine(l2 *StraightLine) bool {
	theta1, theta2 := l.Angle/180*math.Pi, l2.Angle/180*math.Pi
	// Compute the line direction vector
	dx1, dy1 := math.Cos(theta1), math.Sin(theta1)
	dx2, dy2 := math.Cos(theta2), math.Sin(theta2)

	// Solve for equation coefficients
	denom := dx1*dy2 - dy1*dx2
	if math.Abs(denom) < tolerance {
		// The two lines are parallel or coincident
		return false
	}
	return true
}

// IntersectionStraightLine performs the operation.
func (l *StraightLine) IntersectionStraightLine(l2 *StraightLine) (Point, bool) {
	// Extract line parameters
	x1, y1, theta1 := l.Point.X, l.Point.Y, l.Angle/180*math.Pi
	x2, y2, theta2 := l2.Point.X, l2.Point.Y, l2.Angle/180*math.Pi

	// Compute the line direction vector
	dx1, dy1 := math.Cos(theta1), math.Sin(theta1)
	dx2, dy2 := math.Cos(theta2), math.Sin(theta2)

	// Solve for equation coefficients
	denom := dx1*dy2 - dy1*dx2
	if math.Abs(denom) < tolerance {
		// The two lines are parallel or coincident
		return Point{}, false
	}

	// Compute the intersection
	t := ((x2-x1)*dy2 - (y2-y1)*dx2) / denom
	intersectionX := x1 + t*dx1
	intersectionY := y1 + t*dy1

	return Point{intersectionX, intersectionY}, true
}

type Ray struct {
	Point
	Angle float64
}

// NewRay creates and returns a new instance.
func NewRay(p Point, angle float64) *Ray {
	return &Ray{p, angle}
}

// IntersectRay reports whether the condition holds.
func (l *Ray) IntersectRay(l2 *Ray) bool {
	panic("todo")
}

// IntersectionRay performs the operation.
func (r *Ray) IntersectionRay(r2 *Ray) (Point, bool) {
	// Extract ray parameters
	x1, y1, theta1 := r.Point.X, r.Point.Y, r.Angle/180*math.Pi
	x2, y2, theta2 := r2.Point.X, r2.Point.Y, r2.Angle/180*math.Pi

	// Compute the ray direction vector
	dx1, dy1 := math.Cos(theta1), math.Sin(theta1)
	dx2, dy2 := math.Cos(theta2), math.Sin(theta2)

	// Compute the determinant to test parallelism
	denom := dx1*dy2 - dy1*dx2
	if math.Abs(denom) < tolerance {
		// Rays are parallel or colinear
		return Point{}, false
	}

	// Compute intersection (treat as infinite lines)
	t1 := ((x2-x1)*dy2 - (y2-y1)*dx2) / denom
	intersectionX := x1 + t1*dx1
	intersectionY := y1 + t1*dy1
	intersection := Point{X: intersectionX, Y: intersectionY}

	// Check whether the intersection lies on both ray forward halves
	if r.IsOnForwardRange(intersection) && r2.IsOnForwardRange(intersection) {
		return intersection, true
	}

	return Point{}, false
}

// IsOnForwardRange reports whether the condition holds.
func (r *Ray) IsOnForwardRange(p Point) bool {
	dx, dy := p.X-r.Point.X, p.Y-r.Point.Y
	angle := r.Angle / 180 * math.Pi
	dirX, dirY := math.Cos(angle), math.Sin(angle)
	// Test whether the point lies along the ray direction
	dotProduct := dx*dirX + dy*dirY
	return dotProduct > 0 // Point is consistent with the ray direction
}

// IntersectionStraightLine performs the operation.
func (r *Ray) IntersectionStraightLine(l *StraightLine) (Point, bool) {
	// Extract ray and line parameters
	x1, y1, theta := r.Point.X, r.Point.Y, r.Angle/180*math.Pi
	x2, y2, phi := l.Point.X, l.Point.Y, l.Angle/180*math.Pi

	// Compute direction vectors of the ray and line
	dx1, dy1 := math.Cos(theta), math.Sin(theta) // ray direction
	dx2, dy2 := math.Cos(phi), math.Sin(phi)     // line direction

	// Compute the determinant to test parallelism
	denom := dx1*dy2 - dy1*dx2
	if math.Abs(denom) < tolerance {
		// Ray and line are parallel or colinear
		return Point{}, false
	}

	// Compute parameters t and s
	t := ((x2-x1)*dy2 - (y2-y1)*dx2) / denom

	// Check whether t satisfies the ray range (t >= 0)
	if t >= 0 {
		intersectionX := x1 + t*dx1
		intersectionY := y1 + t*dy1
		return Point{intersectionX, intersectionY}, true
	}

	return Point{}, false
}
