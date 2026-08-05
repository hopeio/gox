package geom

import (
	"golang.org/x/exp/constraints"
	"math"
	"sort"
)

// counter clockwise if clockwise ,Start.X,Start.Y <->  endX, endY
type CircularArc struct {
	Circle     Circle
	StartAngle float64
	EndAngle   float64
}

// NewCircularArc creates and returns a new instance.
func NewCircularArc(circle Circle, startAngle, endAngle float64) *CircularArc {
	return &CircularArc{
		Circle:     circle,
		StartAngle: startAngle,
		EndAngle:   endAngle,
	}
}

type CircularArc2 struct {
	Center Point
	Start  Point
	End    Point
}

// NewCircularArc2 creates and returns a new instance.
func NewCircularArc2(center, start, end Point) *CircularArc2 {
	return &CircularArc2{
		Center: center,
		Start:  start,
		End:    end,
	}
}

// ToCircularArc ...
func (a *CircularArc2) ToCircularArc() *CircularArc {
	return CircularArcFromPoints(a.Center, a.Start, a.End)
}

// CircularArcFromPoints ...
func CircularArcFromPoints(center, start, end Point) *CircularArc {
	r := math.Hypot(start.X-center.X, start.Y-center.Y)
	thetaStart := math.Atan2(start.Y-center.Y, start.X-center.X)
	thetaEnd := math.Atan2(end.Y-center.Y, end.X-center.X)
	if thetaStart < 0 {
		thetaStart += 2 * math.Pi
	}
	if thetaEnd < 0 {
		thetaEnd += 2 * math.Pi
	}
	if thetaEnd < thetaStart {
		thetaEnd += 2 * math.Pi
	}
	return &CircularArc{
		Circle: Circle{
			Centre:   center,
			Diameter: r * 2,
		},
		StartAngle: thetaStart / math.Pi * 180,
		EndAngle:   thetaEnd / math.Pi * 180,
	}
}

// Bounds ...
func (a *CircularArc) Bounds() *Bounds {
	r := a.Circle.Diameter / 2
	startAngle, endAngle := a.StartAngle*math.Pi/180, a.EndAngle*math.Pi/180
	angles := []float64{startAngle, endAngle}
	start := float64(int(startAngle/math.Pi)) * math.Pi
	for _, sa := range []float64{start + math.Pi/2, start + math.Pi, start + 3*math.Pi/2, start + math.Pi*2} {
		if startAngle <= sa && sa <= endAngle {
			angles = append(angles, sa)
		}
	}

	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64
	for _, theta := range angles {
		x := a.Circle.Centre.X + r*math.Cos(theta)
		y := a.Circle.Centre.Y + r*math.Sin(theta)
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	return NewBounds(minX, minY, maxX, maxY)
}

// Sample ...
func (a *CircularArc) Sample(samples int) []Point {
	r := a.Circle.Diameter / 2
	startAngle, endAngle := a.StartAngle*math.Pi/180, a.EndAngle*math.Pi/180
	thetaDiff := endAngle - startAngle
	points := make([]Point, 0, samples)
	// Sample points along the arc
	for i := range samples {
		theta := startAngle + thetaDiff*float64(i)/float64(samples)
		x := a.Circle.Centre.X + r*math.Cos(theta)
		y := a.Circle.Centre.Y + r*math.Sin(theta) // Flip y for image coordinate system
		points = append(points, Point{x, y})
	}

	return points
}

// Compute the convex hull of a set of points using Andrew's monotone chain algorithm
func convexHull(points []Point) []Point {
	sort.Slice(points, func(i, j int) bool {
		if points[i].X == points[j].X {
			return points[i].Y < points[j].Y
		}
		return points[i].X < points[j].X
	})

	var lower, upper []Point
	for _, p := range points {
		for len(lower) >= 2 && cross(lower[len(lower)-2], lower[len(lower)-1], p) <= 0 {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, p)
	}

	for i := len(points) - 1; i >= 0; i-- {
		p := points[i]
		for len(upper) >= 2 && cross(upper[len(upper)-2], upper[len(upper)-1], p) <= 0 {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, p)
	}

	return append(lower[:len(lower)-1], upper[:len(upper)-1]...)
}

// Cross product of vectors OA and OB
func cross(o, a, b Point) float64 {
	return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X)
}

// Rotating calipers to find the minimum bounding rectangle
// untested
func minimumBoundingRectangle(hull []Point) *Rectangle {
	minArea := math.MaxFloat64
	var centerX, centerY, length, width, angle float64
	for i := 0; i < len(hull); i++ {
		// Edge vector
		edge := Point{hull[(i+1)%len(hull)].X - hull[i].X, hull[(i+1)%len(hull)].Y - hull[i].Y}
		edgeLength := math.Hypot(edge.X, edge.Y)
		edge.X /= edgeLength
		edge.Y /= edgeLength

		// Perpendicular vector
		perp := Point{-edge.Y, edge.X}

		// Project points onto edge and perpendicular
		var minProj, maxProj float64
		var minPerp, maxPerp float64
		for j, p := range hull {
			proj := edge.X*(p.X-hull[i].X) + edge.Y*(p.Y-hull[i].Y)
			perpProj := perp.X*(p.X-hull[i].X) + perp.Y*(p.Y-hull[i].Y)
			if j == 0 || proj < minProj {
				minProj = proj
			}
			if j == 0 || proj > maxProj {
				maxProj = proj
			}
			if j == 0 || perpProj < minPerp {
				minPerp = perpProj
			}
			if j == 0 || perpProj > maxPerp {
				maxPerp = perpProj
			}
		}

		// Compute area of rectangle
		widthTemp := maxPerp - minPerp
		lengthTemp := maxProj - minProj
		area := widthTemp * lengthTemp
		if area < minArea {
			minArea = area
			centerX = hull[i].X + (minProj+maxProj)/2*edge.X
			centerY = hull[i].Y + (minProj+maxProj)/2*edge.Y
			length = lengthTemp
			width = widthTemp
			angle = math.Atan2(edge.Y, edge.X)
		}
	}

	return NewRect(Point{centerX, centerY}, length, width, angle)
}

// untested
func (a *CircularArc) minimumBoundingRectangle() *Rectangle {
	points := a.Sample(100)
	hull := convexHull(points)
	return minimumBoundingRectangle(hull)
}

type CircularArcInt[T constraints.Integer] struct {
	Circle     CircleInt[T]
	StartAngle float64
	EndAngle   float64
}

// ToFloat64 ...
func (e *CircularArcInt[T]) ToFloat64(factor float64) *CircularArc {
	if factor == 0 {
		factor = 1
	}
	return &CircularArc{
		Circle:     *e.Circle.ToFloat64(factor),
		StartAngle: e.StartAngle,
		EndAngle:   e.EndAngle,
	}
}

// ArcIntFromFloat64 ...
func ArcIntFromFloat64[T constraints.Integer](e *CircularArc, factor float64) *CircularArcInt[T] {
	if factor == 0 {
		factor = 1
	}
	return &CircularArcInt[T]{
		Circle: CircleInt[T]{
			Center:   PointInt[T]{T(math.Round(e.Circle.Centre.X * factor)), T(math.Round(e.Circle.Centre.Y * factor))},
			Diameter: T(math.Round(e.Circle.Diameter * factor)),
		},
		StartAngle: e.StartAngle,
		EndAngle:   e.EndAngle,
	}
}
