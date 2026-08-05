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
	"image/draw"
	"math"
)

// MergeUniformBoundsImagesByOverlap returns the result.
func MergeUniformBoundsImagesByOverlap(imgIdxs [][]int, getImage func(int) image.Image, imgWidth, imgHeight int,
	horizontalOverlaps, verticalOverlaps []int) image.Image {
	var resultWidth, resultHeight int
	for i := range imgIdxs[0] {
		resultWidth += imgWidth
		if i < len(horizontalOverlaps) {
			resultWidth -= horizontalOverlaps[i]
		}
	}
	for i := range imgIdxs {
		resultHeight += imgHeight
		if i < len(verticalOverlaps) {
			resultHeight -= verticalOverlaps[i]
		}
	}

	// Create a new RGBA image for the merged result
	result := image.NewRGBA(image.Rect(0, 0, resultWidth, resultHeight))
	slideWin := image.Rect(0, 0, imgWidth, imgHeight)
	var img image.Image
	// Copy img1 into the result image
	for i, rowimgs := range imgIdxs {
		for j, imgIdx := range rowimgs {
			img = getImage(imgIdx)
			draw.Draw(result, slideWin, img, image.Point{}, draw.Src)
			if j < len(horizontalOverlaps) {
				slideWin.Min.X += slideWin.Dx() - horizontalOverlaps[j]
				slideWin.Max.X += slideWin.Dx() + slideWin.Min.X
			}
		}
		if i < len(verticalOverlaps) {
			slideWin.Min.Y += slideWin.Dy() - verticalOverlaps[i]
			slideWin.Max.Y += slideWin.Dy() + slideWin.Min.Y
			slideWin.Min.X = 0
			slideWin.Max.X = slideWin.Dx()
		}
	}

	return result
}

// MergeUniformBoundsImagesByOverlapReuseMemory returns the result.
func MergeUniformBoundsImagesByOverlapReuseMemory(imgIdxs [][]int, getImage func(int) image.Image, imgWidth, imgHeight int,
	horizontalOverlaps, verticalOverlaps []int, result *image.RGBA) {
	if result == nil {
		panic("result is nil")
	}
	if len(result.Pix) == 0 {
		var resultWidth, resultHeight int
		for i := range imgIdxs[0] {
			resultWidth += imgWidth
			if i < len(horizontalOverlaps) {
				resultWidth -= horizontalOverlaps[i]
			}
		}
		for i := range imgIdxs {
			resultHeight += imgHeight
			if i < len(verticalOverlaps) {
				resultHeight -= verticalOverlaps[i]
			}
		}

		// Create a new RGBA image for the merged result
		result = image.NewRGBA(image.Rect(0, 0, resultWidth, resultHeight))
	}
	slideWin := image.Rect(0, 0, imgWidth, imgHeight)
	var img image.Image
	// Copy img1 into the result image
	for i, rowimgs := range imgIdxs {
		for j, imgIdx := range rowimgs {
			img = getImage(imgIdx)
			draw.Draw(result, slideWin, img, image.Point{}, draw.Src)
			if j < len(horizontalOverlaps) {
				slideWin.Min.X += slideWin.Dx() - horizontalOverlaps[j]
				slideWin.Max.X += slideWin.Dx() + slideWin.Min.X
			}
		}
		if i < len(verticalOverlaps) {
			slideWin.Min.Y += slideWin.Dy() - verticalOverlaps[i]
			slideWin.Max.Y += slideWin.Dy() + slideWin.Min.Y
			slideWin.Min.X = 0
			slideWin.Max.X = slideWin.Dx()
		}
	}
}

type MergeImage struct {
	Pixes                      [][]image.Image
	stride                     int
	effectiveRow, effectiveCol []int
	cacheRow, cacheCol         int
	Rect                       image.Rectangle
}

// ColorModel returns the result.
func (m *MergeImage) ColorModel() color.Model {
	return m.Pixes[0][0].ColorModel()
}

// Bounds returns the result.
func (m *MergeImage) Bounds() image.Rectangle {
	return m.Rect
}

// ImgOffset returns the result.
func (m *MergeImage) ImgOffset(row, col int) image.Image {
	if m.effectiveRow[m.cacheRow] == row {
		m.cacheRow += 1
	} else {
		if m.effectiveRow[m.cacheRow] < row {
			m.cacheRow = findImgIdx(m.effectiveRow, m.cacheRow+1, len(m.effectiveRow), row)
		} else if m.cacheRow-1 >= 0 && m.effectiveRow[m.cacheRow-1] > row {
			m.cacheRow = findImgIdx(m.effectiveRow, 0, m.cacheRow, row)
		}
	}
	if m.effectiveCol[m.cacheCol] == col {
		m.cacheCol += 1
	} else {
		if m.effectiveCol[m.cacheCol] < col {
			m.cacheCol = findImgIdx(m.effectiveCol, m.cacheCol+1, len(m.effectiveCol), col)
		} else if m.cacheCol-1 >= 0 && m.effectiveCol[m.cacheCol-1] > col {
			m.cacheCol = findImgIdx(m.effectiveCol, 0, m.cacheCol, col)
		}
	}
	return m.Pixes[m.cacheCol][m.cacheRow]
}

// findImgIdx returns the result.
func findImgIdx(idx []int, start, end, x int) int {
	for i := start; i < end; i++ {
		if idx[i] > x && (i-1 < 0 || idx[i-1] <= x) {
			return i
		}
	}
	return len(idx) - 1
}

// At returns the result.
func (m *MergeImage) At(x, y int) color.Color {
	if !(image.Point{X: x, Y: y}.In(m.Rect)) {
		return colori.RGB{}
	}
	pix := m.ImgOffset(x, y)
	if m.cacheRow > 0 {
		x -= m.effectiveRow[m.cacheRow-1]
	}
	if m.cacheCol > 0 {
		y -= m.effectiveCol[m.cacheCol-1]
	}
	return pix.At(x, y)
}

// NewMergeImage creates and returns a new instance.
func NewMergeImage(imgs [][]image.Image, width, height int, horizontalOverlaps, verticalOverlaps []int) *MergeImage {
	effectiveRow := make([]int, len(imgs[0]))
	effectiveCol := make([]int, len(imgs))
	var resultWidth, resultHeight int
	for i := range imgs[0] {
		resultWidth += width
		if i < len(horizontalOverlaps) {
			resultWidth -= horizontalOverlaps[i]
		}
		effectiveRow[i] = resultWidth
	}
	for i := range imgs {
		resultHeight += height
		if i < len(verticalOverlaps) {
			resultHeight -= verticalOverlaps[i]
		}
		effectiveCol[i] = resultHeight
	}
	return &MergeImage{
		Pixes:        imgs,
		stride:       width * 3,
		effectiveRow: effectiveRow,
		effectiveCol: effectiveCol,
		Rect:         image.Rect(0, 0, resultWidth, resultHeight),
	}
}

// CalculateContrast returns the result.
func CalculateContrast(img image.Image) float64 {
	var sum int
	// Iterate every pixel
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			sum += int(gray.Y)
		}
	}

	mean := float64(sum) / float64(bounds.Dx()*bounds.Dy())
	var varianceSum float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			v := float64(color.GrayModel.Convert(img.At(x, y)).(color.Gray).Y)
			varianceSum += (v - mean) * (v - mean)
		}
	}

	// Contrast is the standard deviation of luminance
	return math.Sqrt(varianceSum / float64(bounds.Dx()*bounds.Dy()))
}
