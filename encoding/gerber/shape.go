package gerber

import (
	"image"

	"github.com/hopeio/gox/math/geom"
	imagex "github.com/hopeio/gox/media/image"
)

// A Segment is a stroked line.
type Segment struct {
	Interpolation Interpolation
	X             float64
	Y             float64
	CenterX       float64
	CenterY       float64
}

// A Contour is a closed sequence of connected linear or circular segments.
type Contour struct {
	Line     int
	X        float64
	Y        float64
	Segments []Segment
	Polarity bool
}

func (e *Contour) Bounds() *geom.Bounds {
	bounds := image.Rectangle{Min: image.Point{X: int(e.X), Y: int(e.Y)}, Max: image.Point{int(e.X), int(e.Y)}}
	lastPoint := image.Point{X: int(e.X), Y: int(e.Y)}
	for _, s := range e.Segments {
		if s.Interpolation == InterpolationLinear {
			bounds = imagex.RectUnionPoint(bounds, image.Point{X: int(s.X), Y: int(s.Y)})
		} else {
			// 粗暴解决
			bounds = bounds.Union(image.Rect(lastPoint.X, lastPoint.Y, int(s.X), int(s.Y)).Add(image.Pt(int(s.X), int(s.Y))))
		}

	}
	return geom.BoundsFromImageRect(bounds)
}

type Rectangle struct {
	Line     int
	Polarity bool
	geom.Rectangle
}

func (e *Rectangle) Bounds() *geom.Bounds {
	return e.Rectangle.Bounds()
}

type Obround struct {
	Line     int
	Polarity bool
	geom.Rectangle
}

func (e *Obround) Bounds() *geom.Bounds {
	return e.Rectangle.Bounds()
}

type Circle struct {
	Line     int
	Polarity bool
	geom.Circle
}

func (e *Circle) Bounds() *geom.Bounds {
	return e.Circle.Bounds()
}

type Arc struct {
	Line int
	geom.CircularArc2
	StrokeWidth float64
	Interpolation
}

func (e *Arc) Bounds() *geom.Bounds {
	return e.CircularArc2.ToCircularArc().Bounds()
}

type Line struct {
	LineNo int
	geom.LineSegment
	StrokeWidth float64
	Cap         LineCap
}

func (e *Line) Bounds() *geom.Bounds {
	if e.Cap == LineCapButt {
		vector := geom.NewVector(e.Start, e.End)
		return geom.NewRect(geom.Pt((e.Start.X+e.End.X)/2, (e.Start.Y+e.End.Y)/2), vector.Length(), e.StrokeWidth, vector.Angle()).Bounds()
	}
	//TODO
	return nil
}

type ViewBox = geom.Bounds
