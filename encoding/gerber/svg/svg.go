// Package svg parses Gerber to SVG.
package svg

import (
	_ "embed"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/hopeio/gox/encoding/gerber"
	jsonx "github.com/hopeio/gox/encoding/json"
	"github.com/hopeio/gox/math/geom"
	psvg "github.com/hopeio/gox/media/image/svg"
)

// An ElementType is a SVG element type.
type ElementType string

const (
	ElementTypeCircle    ElementType = "Circle"
	ElementTypeRectangle ElementType = "Rect"
	ElementTypePath      ElementType = "Path"
	ElementTypeLine      ElementType = "Line"
	ElementTypeArc       ElementType = "CircularArc"
)

// A Circle is a circle.
type Circle struct {
	Type ElementType
	*gerber.Circle
	Fill string
	Attr map[string]string
}

func (e Circle) Bounds() *geom.Bounds {
	radius := e.Diameter / 2
	return geom.NewBounds(e.Centre.X-radius, e.Centre.Y-radius, e.Centre.X+radius, e.Centre.Y+radius)
}

// MarshalJSON implements json.Marshaler.
func (e Circle) MarshalJSON() ([]byte, error) {
	e.Type = ElementTypeCircle
	return jsonx.Marshal(e)
}

func (e Circle) SetAttr(k, v string) Circle {
	if e.Attr == nil {
		e.Attr = make(map[string]string)
	}
	e.Attr[k] = v
	return e
}

// A Rectangle is a rectangle.
type Rectangle struct {
	Type     ElementType
	Aperture string
	*gerber.Rectangle
	RadiusX float64
	RadiusY float64
	Fill    string
	Attr    map[string]string
}

func (e Rectangle) Bounds() *geom.Bounds {
	return geom.NewBounds(e.Center.X, e.Center.Y, e.Center.X+e.Width, e.Center.Y+e.Height)
}

// MarshalJSON implements json.Marshaler.
func (e Rectangle) MarshalJSON() ([]byte, error) {
	e.Type = ElementTypeRectangle
	return jsonx.Marshal(e)
}

func (e Rectangle) SetAttr(k, v string) Rectangle {
	if e.Attr == nil {
		e.Attr = make(map[string]string)
	}
	e.Attr[k] = v
	return e
}

// A PathLine is a line in a SVG path.
type PathLine struct {
	Type ElementType
	X    float64
	Y    float64
}

// MarshalJSON implements json.Marshaler.
func (e PathLine) MarshalJSON() ([]byte, error) {
	e.Type = ElementTypeLine
	return jsonx.Marshal(e)
}

// A PathArc is an arc in a SVG path.
type PathArc struct {
	Type     ElementType
	XRadius  float64
	YRadius  float64
	LargeArc int
	Sweep    int
	X        float64
	Y        float64
}

// MarshalJSON implements json.Marshaler.
func (e PathArc) MarshalJSON() ([]byte, error) {
	e.Type = ElementTypeArc
	return jsonx.Marshal(e)
}

// A Path is a SVG path.
type Path struct {
	Type     ElementType
	Line     int
	X        float64
	Y        float64
	Commands []interface{}
	Fill     string
	Attr     map[string]string
}

func (e Path) Bounds() (*geom.Bounds, error) {
	bounds := geom.Bounds{Min: geom.Point{math.MaxFloat64, math.MaxFloat64}, Max: geom.Point{-math.MaxFloat64, -math.MaxFloat64}}
	updateMinMax := func(x, y float64) {
		bounds.Min.X = min(bounds.Min.X, x)
		bounds.Max.X = max(bounds.Max.X, x)
		bounds.Min.Y = min(bounds.Min.Y, y)
		bounds.Max.Y = max(bounds.Max.Y, y)
	}

	updateMinMax(e.X, e.Y)
	for _, cmd := range e.Commands {
		switch c := cmd.(type) {
		case PathLine:
			updateMinMax(c.X, c.Y)
		case PathArc:
			updateMinMax(c.X, c.Y)
		default:
			return nil, fmt.Errorf("%#v", c)
		}
	}

	return &bounds, nil
}

// MarshalJSON implements json.Marshaler.
func (e Path) MarshalJSON() ([]byte, error) {
	e.Type = ElementTypePath
	return jsonx.Marshal(e)
}

func (e Path) SetAttr(k, v string) Path {
	if e.Attr == nil {
		e.Attr = make(map[string]string)
	}
	e.Attr[k] = v
	return e
}

// A Line is a SVG line.
type Line struct {
	Type ElementType
	*gerber.Line
	Stroke string
	Attr   map[string]string
}

func (e Line) Bounds() *geom.Bounds {
	return geom.NewBounds(e.Start.X, e.Start.Y, e.End.X, e.End.Y)
}

// MarshalJSON implements json.Marshaler.
func (e Line) MarshalJSON() ([]byte, error) {
	e.Type = ElementTypeLine
	return jsonx.Marshal(e)
}

func (e Line) SetAttr(k, v string) Line {
	if e.Attr == nil {
		e.Attr = make(map[string]string)
	}
	e.Attr[k] = v
	return e
}

// An Arc is a SVG Arc.
type Arc struct {
	Type ElementType
	*gerber.Arc
	RadiusX float64
	RadiusY float64

	LargeArc int
	Sweep    int

	Stroke string
	Attr   map[string]string
}

func (e Arc) Bounds() *geom.Bounds {
	return e.Arc.Bounds()
}

// MarshalJSON implements json.Marshaler.
func (e Arc) MarshalJSON() ([]byte, error) {
	e.Type = ElementTypeArc
	return jsonx.Marshal(e)
}

func (e Arc) SetAttr(k, v string) Arc {
	if e.Attr == nil {
		e.Attr = make(map[string]string)
	}
	e.Attr[k] = v
	return e
}

// A Processor is a performer of Gerber graphic operations.
type Processor struct {
	// Data contains SVG elements.
	Data []interface{}

	// Viewbox of Gerber image.
	*gerber.ViewBox

	// Color for Gerber polarities, defaults to black and white.
	PolarityDark  string
	PolarityClear string

	// Optional scaling factor of coordinates when writing SVG image.
	Scale float64

	// Optional width and height of output SVG image.
	Width  string
	Height string

	// Whether to output javascript for interactive panning and zooming in SVG.
	PanZoom bool
	err     error
}

// NewProcessor creates a Processor.
func NewProcessor() *Processor {
	p := &Processor{}
	p.Data = make([]interface{}, 0)
	p.Scale = 1
	p.PolarityDark = "white"
	p.PolarityClear = "black"
	p.PanZoom = false
	return p
}

func (p *Processor) fill(polarity bool) string {
	if polarity {
		return p.PolarityDark
	}
	return p.PolarityClear
}

func (p *Processor) Circle(circle *gerber.Circle) {
	p.Data = append(p.Data, Circle{Circle: circle, Fill: p.fill(circle.Polarity)})
}

func (p *Processor) Rectangle(rectangle *gerber.Rectangle) {
	rectangle.Center.X -= rectangle.Width / 2
	rectangle.Center.Y += rectangle.Height / 2
	p.Data = append(p.Data, Rectangle{Aperture: "R", Rectangle: rectangle, Fill: p.fill(rectangle.Polarity)})
}

func (p *Processor) Obround(obround *gerber.Obround) {
	r := min(obround.Width, obround.Height) / 2
	obround.Center.X -= obround.Width / 2
	obround.Center.Y += obround.Height / 2
	p.Data = append(p.Data, Rectangle{Aperture: "O", Rectangle: &gerber.Rectangle{obround.Line, obround.Polarity, obround.Rectangle}, RadiusX: r, RadiusY: r, Fill: p.fill(obround.Polarity)})
}

func (p *Processor) Contour(contour *gerber.Contour) {
	if len(contour.Segments) == 1 {
		s := contour.Segments[0]
		if s.Interpolation == gerber.InterpolationClockwise || s.Interpolation == gerber.InterpolationCCW {
			if s.X == contour.X && s.Y == contour.Y {
				vx, vy := s.X-s.CenterX, s.Y-s.CenterY
				r := math.Round(math.Sqrt(vx*vx + vy*vy))
				c := Circle{Circle: &gerber.Circle{Line: contour.Line, Circle: geom.Circle{geom.Pt(s.X, s.Y), r * 2}},
					Fill: p.fill(contour.Polarity)}
				p.Data = append(p.Data, c)
				return
			}
		}
	}

	svgPath := Path{Line: contour.Line, X: contour.X, Y: contour.Y, Fill: p.fill(contour.Polarity)}
	for i, s := range contour.Segments {
		switch s.Interpolation {
		case gerber.InterpolationLinear:
			svgPath.Commands = append(svgPath.Commands, PathLine{X: s.X, Y: s.Y})
		case gerber.InterpolationClockwise:
			fallthrough
		case gerber.InterpolationCCW:
			arc, err := calcArc(contour, i)
			if err != nil {
				p.err = err
				return
			}
			svgPath.Commands = append(svgPath.Commands, arc)
		default:
			p.err = fmt.Errorf("%d %+v", i, s)
			return
		}
	}
	p.Data = append(p.Data, svgPath)
	return
}

func calcArcParams(vs, ve [2]float64, sweep int) (float64, int, error) {
	radiusS := math.Sqrt(math.Pow(float64(vs[0]), 2) + math.Pow(float64(vs[1]), 2))
	radiusE := math.Sqrt(math.Pow(float64(ve[0]), 2) + math.Pow(float64(ve[1]), 2))
	diff := math.Abs(radiusS - radiusE)
	diffRatio := math.Abs(radiusS/radiusE - 1)
	if diff > 3 && diffRatio > 1e-2 {
		return math.NaN(), -1, fmt.Errorf("%f %f %f %f", radiusS, radiusE, diff, diffRatio)
	}

	var largeArc int
	cross := vs[0]*ve[1] - ve[0]*vs[1]
	if (cross > 0) != (sweep == 0) {
		largeArc = 1
	}

	return radiusS, largeArc, nil
}

func calcArc(contour *gerber.Contour, idx int) (PathArc, error) {
	var xs, ys float64
	if idx == 0 {
		xs, ys = contour.X, contour.Y
	} else {
		prev := contour.Segments[idx-1]
		xs, ys = prev.X, prev.Y
	}

	s := contour.Segments[idx]
	arc := PathArc{X: s.X, Y: s.Y}
	switch s.Interpolation {
	case gerber.InterpolationClockwise:
		arc.Sweep = 1
	case gerber.InterpolationCCW:
		arc.Sweep = 0
	default:
		return PathArc{}, fmt.Errorf("%d", s.Interpolation)
	}

	vs := [2]float64{xs - s.CenterX, ys - s.CenterY}
	ve := [2]float64{s.X - s.CenterX, s.Y - s.CenterY}
	if ve == vs {
		return PathArc{}, fmt.Errorf("degenerate arc")
	}

	radius, largeArc, err := calcArcParams(vs, ve, arc.Sweep)
	if err != nil {
		return PathArc{}, fmt.Errorf("%f %f %v %w", xs, ys, s, err)
	}
	arc.XRadius, arc.YRadius = math.Round(radius), math.Round(radius)
	arc.LargeArc = largeArc

	return arc, nil
}

func (p *Processor) Line(gline *gerber.Line) {
	line := Line{Line: gline, Stroke: p.PolarityDark}
	p.Data = append(p.Data, line)
}

func (p *Processor) Arc(garc *gerber.Arc) {
	if garc.End.X == garc.Start.X && garc.End.Y == garc.Start.Y {
		p.err = fmt.Errorf("degenerate arc")
		return
	}

	arc := Arc{Arc: garc, Stroke: p.PolarityDark}
	switch garc.Interpolation {
	case gerber.InterpolationClockwise:
		arc.Sweep = 1
	case gerber.InterpolationCCW:
		arc.Sweep = 0
	}

	vs := [2]float64{garc.Start.X - garc.Center.X, garc.Start.Y - garc.Center.Y}
	ve := [2]float64{garc.End.X - garc.Center.X, garc.End.Y - garc.Center.Y}

	radius, largeArc, err := calcArcParams(vs, ve, arc.Sweep)
	if err != nil {
		p.err = err
		return
	}
	arc.RadiusX, arc.RadiusY = math.Round(radius), math.Round(radius)
	arc.LargeArc = largeArc

	p.Data = append(p.Data, arc)
	return
}

func (p *Processor) SetViewBox(box *gerber.ViewBox) {
	p.ViewBox = box
}

// Write writes Gerber graphics operations as SVG.
func (p *Processor) Write(w io.Writer) error {
	svg := "<svg "
	if p.Width != "" && p.Height != "" {
		svg += fmt.Sprintf(`width="%s" height="%s" `, p.Width, p.Height)
	}
	svg += fmt.Sprintf(`viewBox="%s %s %s %s" style="background-color: %s;" xmlns="http://www.w3.org/2000/svg">`+"\n", p.x(p.Min.X), p.m(p.Min.Y), p.m(p.Max.X), p.m(p.Max.Y), p.PolarityClear)
	if _, err := w.Write([]byte(svg)); err != nil {
		return err
	}

	if p.PanZoom {
		if _, err := w.Write([]byte(`<script xlink:href="svgpan.js"/><g id="viewport" transform="translate(0, 0)">` + "\n")); err != nil {
			return err
		}
	}

	for _, datum := range p.Data {
		bounds, err := Bounds(datum)
		if err != nil {
			return err
		}
		if bounds.Min.X > p.Max.X || p.Min.X > bounds.Max.X || bounds.Min.Y > p.Max.Y || p.Min.Y > bounds.Max.Y {
			continue
		}

		var b []byte
		switch d := datum.(type) {
		case Circle:
			b = []byte(fmt.Sprintf(`<circle cx="%s" cy="%s" r="%s" fill="%s" %s/>`, p.x(d.Centre.X), p.y(d.Centre.Y),
				p.m(d.Diameter/2),
				d.Fill, psvg.FormatAttr(d.Attr)))
		case Rectangle:
			w, h := p.m(d.Width), p.m(d.Height)
			b = []byte(fmt.Sprintf(`<rect x="%s" y="%s" width="%s" height="%s" rx="%s" ry="%s" fill="%s" transform
="rotate(%.1f, %s, %s)" %s/>`, p.x(d.Center.X), p.y(d.Center.Y), w, h, p.m(d.RadiusX), p.m(d.RadiusY), d.Fill, p.a(d.Angle), p.x(d.Center.X+d.Width/2), p.y(d.Center.Y-d.Height/2), psvg.FormatAttr(d.Attr)))
		case Path:
			var err error
			b, err = p.pathBytes(d)
			if err != nil {
				return err
			}
		case Line:
			b = []byte(fmt.Sprintf(`<line x1="%s" y1="%s" x2="%s" y2="%s" stroke-width="%s" stroke-linecap="%s" stroke="%s"%s/>`, p.x(d.Start.X), p.y(d.Start.Y), p.x(d.End.X), p.y(d.End.Y), p.m(d.StrokeWidth), d.Cap, d.Stroke, psvg.FormatAttr(d.Attr)))
		case Arc:
			b = []byte(fmt.Sprintf(`<path d="M %s %s A %s %s 0 %d %d %s %s" stroke-width="%s" stroke="%s" stroke
-linecap="round"%s/>`, p.x(d.Start.X), p.y(d.Start.Y), p.m(d.RadiusX), p.m(d.RadiusY), d.LargeArc, d.Sweep, p.x(d.End.X), p.y(d.End.Y), p.m(d.StrokeWidth), d.Stroke, psvg.FormatAttr(d.Attr)))
		default:
			return fmt.Errorf("%+v", d)
		}
		if _, err := w.Write(append(b, '\n')); err != nil {
			return err
		}
	}

	if p.PanZoom {
		if _, err := w.Write([]byte(`</g>`)); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte(`</svg>`)); err != nil {
		return err
	}
	return nil
}

func Bounds(element interface{}) (*geom.Bounds, error) {
	switch e := element.(type) {
	case Circle:
		return e.Bounds(), nil
	case Rectangle:
		return e.Bounds(), nil
	case Path:
		return e.Bounds()
	case Line:
		return e.Bounds(), nil
	case Arc:
		return e.Bounds(), nil
	default:
		return nil, fmt.Errorf("%#v", e)
	}
}

func (p *Processor) pathBytes(svgp Path) ([]byte, error) {
	cmds := []string{fmt.Sprintf("M %s %s", p.x(svgp.X), p.y(svgp.Y))}
	for _, cmd := range svgp.Commands {
		var s string
		switch c := cmd.(type) {
		case PathLine:
			s = fmt.Sprintf("L %s %s", p.x(c.X), p.y(c.Y))
		case PathArc:
			s = fmt.Sprintf("A %s %s 0 %d %d %s %s", p.m(c.XRadius), p.m(c.YRadius), c.LargeArc, c.Sweep, p.x(c.X), p.y(c.Y))
		default:
			return nil, fmt.Errorf("%+v", c)
		}
		cmds = append(cmds, s)
	}
	b := fmt.Sprintf(`<path d="%s" fill="%s" line="%d"%s/>`, strings.Join(cmds, " "), svgp.Fill, svgp.Line, psvg.FormatAttr(svgp.Attr))
	return []byte(b), nil
}

func (p *Processor) x(x float64) string {
	return strconv.FormatFloat(x*p.Scale, 'f', -1, 64)
}

func (p *Processor) y(y float64) string {
	return strconv.FormatFloat((p.Max.Y-y)*p.Scale, 'f', -1, 64)
}

func (p *Processor) m(f float64) string {
	return strconv.FormatFloat(f*p.Scale, 'f', -1, 64)
}

func (p *Processor) a(a float64) float64 {
	return 360 - a
}

// SVG parses Gerber input into SVG.
func SVG(r io.Reader) (*Processor, error) {
	processor := NewProcessor()
	parser := gerber.NewParser(processor)
	if err := parser.Parse(r); err != nil {
		return nil, err
	}

	return processor, nil
}
