/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package svg

import (
	"fmt"
	stringsx "github.com/hopeio/gox/strings"
	"strconv"
)

type Path struct {
	X        float64
	Y        float64
	Commands []fmt.Stringer
	Stroke   string
	Fill     string
	Attr     map[string]string
}

func (r *Path) String() string {
	return fmt.Sprintf(`<path d="M%f %f %s" fill="%s" stroke="%s" %s`, r.X, r.Y, stringsx.Join(r.Commands, " "), r.Fill, r.Stroke, FormatAttr(r.Attr))
}

type PathArcA struct {
	XRadius  float64 // x轴半径
	YRadius  float64 // y轴半径
	Angle    float64
	LargeArc float64
	Sweep    float64
	X        float64
	Y        float64
}

func (r *PathArcA) String() string {
	return fmt.Sprintf(`A%f %f %f %f %f %f %f`, r.XRadius, r.YRadius, r.Angle, r.LargeArc, r.Sweep, r.X, r.Y)
}

type PathArcC struct {
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
	X  float64
	Y  float64
}

func (r *PathArcC) String() string {
	return fmt.Sprintf(`C%f %f %f %f %f %f`, r.X1, r.Y1, r.X2, r.Y2, r.X, r.Y)
}

type PathArcS struct {
	X1 float64
	Y1 float64
	X  float64
	Y  float64
}

func (r *PathArcS) String() string {
	return fmt.Sprintf(`S%f %f %f %f`, r.X1, r.Y1, r.X, r.Y)
}

type PathArcQ struct {
	X1 float64
	Y1 float64
	X  float64
	Y  float64
}

func (r *PathArcQ) String() string {
	return fmt.Sprintf(`Q%f %f %f %f`, r.X1, r.Y1, r.X, r.Y)
}

type PathArcT struct {
	X float64
	Y float64
}

func (r *PathArcT) String() string {
	return fmt.Sprintf(`T%f %f`, r.X, r.Y)
}

type PathL struct {
	X float64
	Y float64
}

func (r *PathL) String() string {
	return fmt.Sprintf(`L%f %f`, r.X, r.Y)
}

type PathZ struct {
}

func (r *PathZ) String() string {
	return fmt.Sprintf(`Z`)
}

type Circle struct {
	X           float64
	Y           float64
	Radius      float64
	StrokeWidth float64
	Stroke      string
	Fill        string
	Attr        map[string]string
}

func (r *Circle) String() string {
	return fmt.Sprintf(`<circle cx="%f" cy="%f" stroke="%s" stroke-width="%f" %s`, r.X, r.Y, r.Stroke, r.StrokeWidth, FormatAttr(r.Attr))
}

type Ellipse struct {
	X           float64
	Y           float64
	XRadius     float64
	YRadius     float64
	StrokeWidth float64
	Stroke      string
	Fill        string
	Attr        map[string]string
}

func (r *Ellipse) String() string {
	return fmt.Sprintf(`<ellipse cx="%f" cy="%f" rx="%f" ry="%f" stroke="%s" stroke-width="%f" %s`, r.X, r.Y, r.XRadius, r.YRadius, r.Stroke, r.StrokeWidth, FormatAttr(r.Attr))
}

type Line struct {
	StartX      float64
	StartY      float64
	EndX        float64
	EndY        float64
	StrokeWidth float64
	Stroke      string
	Attr        map[string]string
}

func (r *Line) String() string {
	return fmt.Sprintf(`<line x1="%f" x2="%f" y1="%f" y2="%f" stroke="%s" stroke-width="%f" %s`, r.StartX, r.EndX, r.StartY, r.EndY, r.Stroke, r.StrokeWidth, FormatAttr(r.Attr))
}

type Polyline struct {
	Points      []float64
	StrokeWidth float64
	Stroke      string
	Fill        string
	Attr        map[string]string
}

func (r *Polyline) String() string {
	return fmt.Sprintf(`<polyline points="%s"  stroke="%s" fill="%s" stroke-width="%f" %s`, stringsx.JoinValueFunc(r.Points, func(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }, " "), r.Stroke, r.Fill, r.StrokeWidth, FormatAttr(r.Attr))
}

type Polygon struct {
	Points      []float64
	StrokeWidth float64
	Stroke      string
	Fill        string
	Attr        map[string]string
}

func (r *Polygon) String() string {
	return fmt.Sprintf(`<polygon points="%s"  stroke="%s" fill="%s" stroke-width="%f" %s`, stringsx.JoinValueFunc(r.Points, func(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }, " "), r.Stroke, r.Fill, r.StrokeWidth, FormatAttr(r.Attr))
}

type Rectangle struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
	RX     float64
	RY     float64
	Angle  float64
	Fill   string
	Attr   map[string]string
}

func (r *Rectangle) String() string {
	if r.Angle != 0 {
		return fmt.Sprintf(`<rect x="%f" y="%f" width="%f" height="%f" rx="%f" ry="%f" fill="%s" transform="rotate(%f,%f,%f)" %s />`,
			r.X, r.Y, r.Width, r.Height, r.RX, r.RY, r.Fill, r.Angle, r.X+(r.Width/2), r.Y+(r.Height/2), FormatAttr(r.Attr))
	}
	return fmt.Sprintf(`<rect x="%f" y="%f" width="%f" height="%f" rx="%f" ry="%f" fill="%s" %s />`,
		r.X, r.Y, r.Width, r.Height, r.RX, r.RY, r.Fill, FormatAttr(r.Attr))
}

func FormatAttr(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	s := ""
	for k, v := range m {
		s += fmt.Sprintf(` %s="%s"`, k, v)
	}
	return s
}
