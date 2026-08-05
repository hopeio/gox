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

// Circle ...
func (s *StoreProcessor) Circle(circle *Circle) {
	s.Circles = append(s.Circles, circle)
}

// Rectangle ...
func (s *StoreProcessor) Rectangle(rectangle *Rectangle) {
	s.Rects = append(s.Rects, rectangle)
}

// Obround ...
func (s *StoreProcessor) Obround(obround *Obround) {
	s.Obrounds = append(s.Obrounds, obround)
}

// Contour ...
func (s *StoreProcessor) Contour(contour *Contour) {
	s.Contours = append(s.Contours, contour)
}

// Line ...
func (s *StoreProcessor) Line(line *Line) {
	s.Lines = append(s.Lines, line)
}

// Arc ...
func (s *StoreProcessor) Arc(arc *Arc) {
	s.Arcs = append(s.Arcs, arc)
}

// SetViewBox ...
func (s *StoreProcessor) SetViewBox(box *ViewBox) {
	s.ViewBox = box
}

var _ Processor = (*StoreProcessor)(nil)
