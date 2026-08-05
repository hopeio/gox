/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package color

import "image/color"

type RGBu16 struct {
	R, G, B uint16
}

// RGBA performs the operation.
func (c RGBu16) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = 0xffff
	return
}

// NewRGB64 creates and returns a new instance.
func NewRGB64(r, g, b uint16) RGBu16 {
	return RGBu16{R: r, G: g, B: b}
}

type RGB struct {
	R, G, B uint8
}

// RGBA performs the operation.
func (c RGB) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = 0xffff
	return
}

// NewRGB creates and returns a new instance.
func NewRGB(r, g, b uint8) RGB {
	return RGB{R: r, G: g, B: b}
}

// rgbModel returns the result.
func rgbModel(c color.Color) color.Color {
	if _, ok := c.(RGB); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	return RGB{uint8(r), uint8(g), uint8(b)}
}

var (
	RGBModel = color.ModelFunc(rgbModel)
)

// ColorRGBAu8 performs the operation.
func ColorRGBAu8(c color.Color) (r, g, b, a uint8) {
	r32, g32, b32, a32 := c.RGBA()
	return uint8(r32), uint8(g32), uint8(b32), uint8(a32)
}
