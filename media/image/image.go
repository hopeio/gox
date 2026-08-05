/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package image

import (
	colori "github.com/hopeio/gox/media/image/color"
	"image"
	"image/color"
	"math"
)

// CalculateOverlap returns the result.
func CalculateOverlap(img1, img2 image.Image, col bool, minOverlap, maxOverlap int) int {
	bounds1, bounds2 := img1.Bounds(), img2.Bounds()
	dx1, dy1 := bounds1.Dx(), bounds1.Dy()
	dx2, dy2 := bounds2.Dx(), bounds2.Dy()
	dx, dy := min(dx1, dx2), min(dy1, dy2)
	minX1, minY1 := max(bounds1.Min.X, bounds1.Max.X-dx), bounds1.Min.Y
	minX2, minY2 := bounds2.Min.X, bounds2.Min.Y
	maxX1, maxY1 := minX1+dx, minY1+dy
	maxY2 := minY2 + dy
	maxOverlap = min(maxOverlap, dy)
	if col {
		minX1, minY1 = max(bounds1.Min.Y, bounds1.Max.Y-dy), bounds1.Min.X
		minX2, minY2 = bounds2.Min.Y, bounds2.Min.X
		maxX1, maxY1 = minY1+dy, minX1+dx
		maxY2 = minX2 + dx
		dy = dx
	}
	data1 := make([]uint8, maxOverlap*dy)
	// Iterate pixels of the source image and convert to grayscale
	var i int
	var c color.Color
	for x := maxX1 - maxOverlap; x < maxX1; x++ {
		for y := minY1; y < maxY1; y++ {
			if col {
				c = img1.At(y, x)
			} else {
				c = img1.At(x, y)
			}
			// Compute grayscale with a weighted average
			gray := colori.ColorToGray(c)
			data1[i] = gray.Y
			i++
		}
	}
	data2 := make([]uint8, maxOverlap*dy)
	var j int
	for x := minX2; x < maxOverlap; x++ {
		for y := minY2; y < maxY2; y++ {
			if col {
				c = img2.At(y, x)
			} else {
				c = img2.At(x, y)
			}
			// Compute grayscale with a weighted average
			gray := colori.ColorToGray(c)
			data2[j] = gray.Y
			j++
		}
	}
	n := len(data1)
	minMean := math.MaxFloat64
	y := bounds2.Dy()
	var overlap int
	for o := minOverlap; o <= maxOverlap; o++ {
		var sum int
		m := o * y
		subimg1 := data1[n-m:]
		subimg2 := data2[:m]
		for i := range m {
			v := int(subimg1[i]) - int(subimg2[i])
			sum += v * v
		}
		mse := float64(sum) / float64(m)
		if mse < minMean {
			minMean = mse
			overlap = o
		}
	}

	return overlap
}

// CalculateOverlapReuseMemory returns the result.
func CalculateOverlapReuseMemory(img1, img2 image.Image, col bool, minOverlap, maxOverlap int, gary1, gary2 []uint8) int {
	bounds1, bounds2 := img1.Bounds(), img2.Bounds()
	dx1, dy1 := bounds1.Dx(), bounds1.Dy()
	dx2, dy2 := bounds2.Dx(), bounds2.Dy()
	dx, dy := min(dx1, dx2), min(dy1, dy2)
	minX1, minY1 := max(bounds1.Min.X, bounds1.Max.X-dx), bounds1.Min.Y
	minX2, minY2 := bounds2.Min.X, bounds2.Min.Y
	maxX1, maxY1 := minX1+dx, minY1+dy
	maxY2 := minY2 + dy
	maxOverlap = min(maxOverlap, dy)
	if col {
		minX1, minY1 = max(bounds1.Min.Y, bounds1.Max.Y-dy), bounds1.Min.X
		minX2, minY2 = bounds2.Min.Y, bounds2.Min.X
		maxX1, maxY1 = minY1+dy, minX1+dx
		maxY2 = minX2 + dx
		dy = dx
	}
	if len(gary1) == 0 {
		gary1 = make([]uint8, maxOverlap*dy)
	}
	// Iterate pixels of the source image and convert to grayscale
	var i int
	var c color.Color
	for x := maxX1 - maxOverlap; x < maxX1; x++ {
		for y := minY1; y < maxY1; y++ {
			if col {
				c = img1.At(y, x)
			} else {
				c = img1.At(x, y)
			}
			r, g, b, _ := c.RGBA()
			// Compute grayscale with a weighted average
			gary1[i] = uint8((19595*r + 38470*g + 7471*b + 1<<15) >> 24)
			i++
		}
	}
	if len(gary2) == 0 {
		gary2 = make([]uint8, maxOverlap*dy)
	}
	var j int
	for x := minX2; x < maxOverlap; x++ {
		for y := minY2; y < maxY2; y++ {
			if col {
				c = img2.At(y, x)
			} else {
				c = img2.At(x, y)
			}
			r, g, b, _ := c.RGBA()
			// Compute grayscale with a weighted average
			gary2[j] = uint8((19595*r + 38470*g + 7471*b + 1<<15) >> 24)
			j++
		}
	}
	n := len(gary1)
	minMean := math.MaxFloat64
	y := bounds2.Dy()
	var overlap int
	for o := minOverlap; o <= maxOverlap; o++ {
		var sum int
		m := o * y
		subimg1 := gary1[n-m:]
		subimg2 := gary2[:m]
		for i := range m {
			v := int(subimg1[i]) - int(subimg2[i])
			sum += v * v
		}
		mse := float64(sum) / float64(m)
		if mse < minMean {
			minMean = mse
			overlap = o
		}
	}

	return overlap
}

// RectRotateByCenter returns the result.
func RectRotateByCenter(x, y, l, w int, angle float64) []image.Point {
	rad := angle / 180 * math.Pi
	lSinYAxis := int(float64(l) / 2 * math.Sin(rad))
	lCosXAxis := int(float64(l) / 2 * math.Cos(rad))
	wSinXAxis := int(float64(w) / 2 * math.Sin(rad))
	wCosYAxis := int(float64(w) / 2 * math.Cos(rad))
	return []image.Point{
		{X: x - lCosXAxis - wSinXAxis, Y: y + lSinYAxis - wCosYAxis},
		{X: x + lCosXAxis - wSinXAxis, Y: y - lSinYAxis - wCosYAxis},
		{X: x + lCosXAxis + wSinXAxis, Y: y - lSinYAxis + wCosYAxis},
		{X: x - lCosXAxis + wSinXAxis, Y: y + lSinYAxis + wCosYAxis},
	}
}

// ToGary converts the value.
func ToGary(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gary := image.NewGray(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			gary.Set(x, y, color.Gray{Y: uint8((19595*r + 38470*g + 7471*b + 1<<15) >> 24)})
		}
	}
	return gary
}

// ToGaryReuseMemory converts the value.
func ToGaryReuseMemory(img image.Image, gary *image.Gray) {
	bounds := img.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			gary.Set(x, y, color.Gray{Y: uint8((19595*r + 38470*g + 7471*b + 1<<15) >> 24)})
		}
	}
}
