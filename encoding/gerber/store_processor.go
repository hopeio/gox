package gerber

type StoreProcessor struct {
	Circles  []*Circle
	Rects    []*Rectangle
	Obrounds []*Obround
	Lines    []*Line
	Contours []*Contour
	Arcs     []*Arc
	ViewBox  *ViewBox
}

// Circle performs the operation.
func (s *StoreProcessor) Circle(circle *Circle) {
	s.Circles = append(s.Circles, circle)
}

// Rectangle performs the operation.
func (s *StoreProcessor) Rectangle(rectangle *Rectangle) {
	s.Rects = append(s.Rects, rectangle)
}

// Obround performs the operation.
func (s *StoreProcessor) Obround(obround *Obround) {
	s.Obrounds = append(s.Obrounds, obround)
}

// Contour performs the operation.
func (s *StoreProcessor) Contour(contour *Contour) {
	s.Contours = append(s.Contours, contour)
}

// Line performs the operation.
func (s *StoreProcessor) Line(line *Line) {
	s.Lines = append(s.Lines, line)
}

// Arc performs the operation.
func (s *StoreProcessor) Arc(arc *Arc) {
	s.Arcs = append(s.Arcs, arc)
}

// SetViewBox updates or inserts a value.
func (s *StoreProcessor) SetViewBox(box *ViewBox) {
	s.ViewBox = box
}

var _ Processor = (*StoreProcessor)(nil)
