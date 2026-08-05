package geom

type Polygon []Point

// ContainsPoint reports whether the condition holds.
func (p Polygon) ContainsPoint(point Point) bool {
	inside := false
	j := len(p) - 1 // The last vertex connects to the first one.

	for i := range p {
		// Check if the point is on an edge of the polygon
		if NewLineSegment(p[i], p[j]).ContainsPoint(point) {
			return true
		}

		// Check intersection with horizontal ray extending right from the point
		if (p[i].Y > point.Y) != (p[j].Y > point.Y) && point.X < (p[j].X-p[i].X)*(point.Y-p[i].Y)/(p[j].Y-p[i].Y)+p[i].X {
			inside = !inside
		}
		j = i // Move to next pair of edges
	}

	return inside
}

type RegularPolygon struct {
	Centre Point
	Radius float64
	Sides  int
}
