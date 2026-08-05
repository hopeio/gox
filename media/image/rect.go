package image

import (
	"image"
)

// RectUnionPoint returns the result.
func RectUnionPoint(rect image.Rectangle, p image.Point) image.Rectangle {
	if p.X < rect.Min.X {
		rect.Min.X = p.X
	}
	if p.X > rect.Max.X {
		rect.Max.X = p.X
	}
	if p.Y < rect.Min.Y {
		rect.Min.Y = p.Y
	}
	if p.Y > rect.Max.Y {
		rect.Max.Y = p.Y
	}
	return rect
}

// RectClipInBounds performs the operation.
func RectClipInBounds(rect *image.Rectangle, imgWidth, imgHeight int) {
	if rect.Min.X < 0 {
		rect.Min.X = 0
	}
	if rect.Max.X > imgWidth {
		rect.Max.X = imgWidth
	}
	if rect.Min.Y < 0 {
		rect.Min.Y = 0
	}
	if rect.Max.Y > imgHeight {
		rect.Max.Y = imgHeight
	}
}
