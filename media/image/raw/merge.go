/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package raw

import (
	colori "github.com/hopeio/gox/media/image/color"
	"image"
	"image/color"
)

type MergeBGR struct {
	Pixes                           [][][]uint8
	stride                          int
	effectiveWidth, effectiveHeight []int
	cacheXIdx, cacheYIdx            int
	Rect                            image.Rectangle
}

// ColorModel ...
func (m *MergeBGR) ColorModel() color.Model {
	return colori.RGBModel
}

// Bounds ...
func (m *MergeBGR) Bounds() image.Rectangle {
	return m.Rect
}

// ImgOffset ...
func (m *MergeBGR) ImgOffset(x, y int) []uint8 {
	if m.effectiveWidth[m.cacheXIdx] == x {
		m.cacheXIdx += 1
	} else {
		if m.effectiveWidth[m.cacheXIdx] < x {
			m.cacheXIdx = findImgIdx(m.effectiveWidth, m.cacheXIdx+1, len(m.effectiveWidth), x)
		} else if m.cacheXIdx-1 >= 0 && m.effectiveWidth[m.cacheXIdx-1] > x {
			m.cacheXIdx = findImgIdx(m.effectiveWidth, 0, m.cacheXIdx, x)
		}
	}
	if m.effectiveHeight[m.cacheYIdx] == y {
		m.cacheYIdx += 1
	} else {
		if m.effectiveHeight[m.cacheYIdx] < y {
			m.cacheYIdx = findImgIdx(m.effectiveHeight, m.cacheYIdx+1, len(m.effectiveHeight), y)
		} else if m.cacheYIdx-1 >= 0 && m.effectiveHeight[m.cacheYIdx-1] > y {
			m.cacheYIdx = findImgIdx(m.effectiveHeight, 0, m.cacheYIdx, y)
		}
	}
	return m.Pixes[m.cacheYIdx][m.cacheXIdx]
}

// findImgIdx ...
func findImgIdx(idx []int, start, end, x int) int {
	for i := start; i < end; i++ {
		if idx[i] > x && (i-1 < 0 || idx[i-1] <= x) {
			return i
		}
	}
	return len(idx) - 1
}

// At ...
func (m *MergeBGR) At(x, y int) color.Color {
	if !(image.Point{X: x, Y: y}.In(m.Rect)) {
		return colori.RGB{}
	}
	pix := m.ImgOffset(x, y)
	if m.cacheXIdx > 0 {
		x -= m.effectiveWidth[m.cacheYIdx-1]
	}
	if m.cacheYIdx > 0 {
		y -= m.effectiveHeight[m.cacheYIdx-1]
	}
	i := y*m.stride + x*3
	b, g, cr := pix[i], pix[i+1], pix[i+2]
	return colori.RGB{R: cr, G: g, B: b}
}

// NewMergeBGR creates and returns a new instance.
func NewMergeBGR(rawValues [][][]byte, width, height int, horizontalOverlaps, verticalOverlaps []int) *MergeBGR {
	effectiveWidth := make([]int, len(rawValues[0]))
	effectiveHeight := make([]int, len(rawValues))
	var resultWidth, resultHeight int
	for i := range rawValues[0] {
		resultWidth += width
		if i < len(horizontalOverlaps) {
			resultWidth -= horizontalOverlaps[i]
		}
		effectiveWidth[i] = resultWidth
	}
	for i := range rawValues {
		resultHeight += height
		if i < len(verticalOverlaps) {
			resultHeight -= verticalOverlaps[i]
		}
		effectiveHeight[i] = resultHeight
	}
	return &MergeBGR{
		Pixes:           rawValues,
		stride:          width * 3,
		effectiveWidth:  effectiveWidth,
		effectiveHeight: effectiveHeight,
		Rect:            image.Rect(0, 0, resultWidth, resultHeight),
	}
}
