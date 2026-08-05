/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package image

import (
	"image"
	"image/color"
)

type AveragePoolImage struct {
	img               image.Image
	Rect              image.Rectangle
	xyReductionFactor int
	reductionFactor   int
}

// ColorModel returns the result.
func (img *AveragePoolImage) ColorModel() color.Model {
	return img.img.ColorModel()
}

// Bounds returns the result.
func (img *AveragePoolImage) Bounds() image.Rectangle {
	return img.Rect
}

// At returns the result.
func (img *AveragePoolImage) At(x, y int) color.Color {
	var r, b, g, a uint32
	reductionFactor := uint32(img.reductionFactor)
	for i := range img.xyReductionFactor {
		for j := range img.xyReductionFactor {
			ox, oy := x*img.xyReductionFactor+i, y*img.xyReductionFactor+j
			if ox >= img.Rect.Max.X || oy >= img.Rect.Max.Y {
				reductionFactor--
				continue
			}
			c := img.img.At(ox, oy)
			cr, cb, cg, ca := c.RGBA()
			r += cr
			g += cg
			b += cb
			a += ca
		}
	}
	return color.RGBA64{uint16(r / reductionFactor), uint16(g / reductionFactor), uint16(b / reductionFactor), uint16(a / reductionFactor)}
}

// NewAveragePoolImage creates and returns a new instance.
func NewAveragePoolImage(img image.Image, reductionFactor int) *AveragePoolImage {
	bounds := img.Bounds()
	bounds.Max.X = bounds.Max.X / reductionFactor
	if bounds.Max.X*reductionFactor != bounds.Dx() {
		bounds.Max.X++
	}
	bounds.Max.Y = bounds.Max.Y / reductionFactor
	if bounds.Max.Y*reductionFactor != bounds.Dy() {
		bounds.Max.Y++
	}
	return &AveragePoolImage{img: img, Rect: image.Rect(0, 0, bounds.Dx(), bounds.Dy()), xyReductionFactor: reductionFactor, reductionFactor: reductionFactor * reductionFactor}
}
