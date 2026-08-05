/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package raw

import (
	"fmt"
	colori "github.com/hopeio/gox/media/image/color"
	"image"
	"image/color"
)

type BGR struct {
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int // 3 * r.Dx()
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// ColorModel ...
func (raw *BGR) ColorModel() color.Model {
	return colori.RGBModel
}

// Bounds ...
func (raw *BGR) Bounds() image.Rectangle {
	return raw.Rect
}

// PixOffset ...
func (raw *BGR) PixOffset(x, y int) int {
	return (y-raw.Rect.Min.Y)*raw.Stride + (x-raw.Rect.Min.X)*3
}

// At ...
func (raw *BGR) At(x, y int) color.Color {
	if !(image.Point{X: x, Y: y}.In(raw.Rect)) {
		return colori.RGB{}
	}
	i := raw.PixOffset(x, y)
	b, g, r := raw.Pix[i], raw.Pix[i+1], raw.Pix[i+2]
	return colori.RGB{R: r, G: g, B: b}
}

// Set ...
func (raw *BGR) Set(x, y int, c color.Color) {
	if !(image.Point{X: x, Y: y}.In(raw.Rect)) {
		return
	}
	i := raw.PixOffset(x, y)
	r, g, b, _ := c.RGBA()
	raw.Pix[i], raw.Pix[i+1], raw.Pix[i+2] = uint8(b), uint8(g), uint8(r)
}

// NewBGR creates and returns a new instance.
func NewBGR(rawValues []byte, width, height int) (*BGR, error) {
	if len(rawValues) != width*height*3 {
		return nil, fmt.Errorf("invalid image raw data")
	}
	return &BGR{
		Pix:    rawValues,
		Stride: width * 3,
		Rect:   image.Rect(0, 0, width, height),
	}, nil
}
