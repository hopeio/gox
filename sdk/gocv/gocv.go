/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gocv

import (
	"image"
	"image/color"

	imagex "github.com/hopeio/gox/media/image"
	"gocv.io/x/gocv"
)

// SearchCircle performs the operation.
func SearchCircle(path string, radius int) (circles []imagex.Circle, err error) {
	gimg := gocv.IMRead(path, gocv.IMReadGrayScale)
	// Define Gaussian kernel size and sigma
	blurred := gocv.NewMat()
	defer blurred.Close()
	gocv.GaussianBlur(gimg, &blurred, image.Pt(9, 9), 0, 0, gocv.BorderDefault)
	circleMap := gocv.NewMat()
	defer circleMap.Close()
	gocv.HoughCirclesWithParams(blurred, &circleMap, gocv.HoughGradient, 1, float64(radius), 30,
		30, radius, radius)
	if !circleMap.Empty() {
		for i := range circleMap.Cols() {
			v := circleMap.GetVecfAt(0, i)
			x := int(v[0])
			y := int(v[1])
			r := int(v[2])
			circles = append(circles, imagex.Circle{Center: image.Pt(x, y), Radius: r})
		}
	}
	return
}

// MergeUniformBoundsImagesByOverlap returns the result.
func MergeUniformBoundsImagesByOverlap(imgIdxs [][]int, getImage func(int) ([]byte, error), imgWidth, imgHeight int,
	horizontalOverlaps, verticalOverlaps []int, dst string) error {
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

	data, err := getImage(0)
	if err != nil {
		return err
	}
	img0, err := gocv.IMDecode(data, gocv.IMReadAnyColor|gocv.IMReadAnyDepth)
	if err != nil {
		return err
	}
	result := gocv.NewMatWithSize(resultHeight, resultWidth, img0.Type())

	var bounds = image.Rect(0, 0, imgWidth, imgHeight)
	var img gocv.Mat
	// Copy img1 into the result image
	for i, rowimgs := range imgIdxs {
		for j, imgIdx := range rowimgs {
			if imgIdx != 0 {
				data, err = getImage(imgIdx)
				if err != nil {
					return err
				}
			} else {
				img = img0
			}
			img, err = gocv.IMDecode(data, gocv.IMReadAnyColor|gocv.IMReadAnyDepth)
			if err != nil {
				return err
			}
			rect := result.Region(bounds)
			img.CopyTo(&rect)
			img.Close()
			if j < len(horizontalOverlaps) {
				bounds.Min.X += bounds.Dx() - horizontalOverlaps[j]
				bounds.Max.X = bounds.Dx() + bounds.Min.X
			}
		}
		if i < len(verticalOverlaps) {
			bounds.Min.Y += bounds.Dy() - verticalOverlaps[i]
			bounds.Max.Y = bounds.Dy() + bounds.Min.Y
			bounds.Min.X = 0
			bounds.Max.X = bounds.Dx()
		}
	}
	gocv.IMWrite(dst, result)
	return nil
}

// Sharpness performs the operation.
func Sharpness(imgPath string, rect image.Rectangle) (float64, error) {
	img := gocv.IMRead(imgPath, gocv.IMReadGrayScale|gocv.IMReadAnyDepth)

	img = img.Region(rect)
	laplacian := gocv.NewMat()
	defer laplacian.Close()
	// Compute the Laplacian operator's standard deviation
	gocv.Laplacian(img, &laplacian, gocv.MatTypeCV64F, 1, 1, 0, gocv.BorderDefault)
	// Compute the standard deviation
	mean, stddev := gocv.NewMat(), gocv.NewMat()
	defer mean.Close()
	defer stddev.Close()
	gocv.MeanStdDev(laplacian, &mean, &stddev)
	return stddev.GetDoubleAt(0, 0), nil
}

// AffineMatByPoints returns the result.
func AffineMatByPoints(p1, p2, p3, q1, q2, q3 gocv.Point2f) gocv.Mat {
	src := gocv.NewMatWithSize(3, 1, gocv.MatTypeCV32FC2)
	defer src.Close()
	dst := gocv.NewMatWithSize(3, 1, gocv.MatTypeCV32FC2)
	defer dst.Close()
	src.SetFloatAt(0, 0, p1.X)
	src.SetFloatAt(0, 1, p1.Y)
	dst.SetFloatAt(0, 0, q1.X)
	dst.SetFloatAt(0, 1, q1.Y)
	src.SetFloatAt(1, 0, p2.X)
	src.SetFloatAt(1, 1, p2.Y)
	dst.SetFloatAt(1, 0, q2.X)
	dst.SetFloatAt(1, 1, q2.Y)
	src.SetFloatAt(2, 0, p3.X)
	src.SetFloatAt(2, 1, p3.Y)
	dst.SetFloatAt(2, 0, q3.X)
	dst.SetFloatAt(2, 1, q3.Y)
	srcVec, dstVec := gocv.NewPoint2fVectorFromMat(src), gocv.NewPoint2fVectorFromMat(dst)
	defer srcVec.Close()
	defer dstVec.Close()
	return gocv.GetAffineTransform2f(srcVec, dstVec)
}

// AffineMat returns the result.
func AffineMat(src []gocv.Point2f, dst []gocv.Point2f) gocv.Mat {
	pvsrc := gocv.NewPoint2fVectorFromPoints(src)
	defer pvsrc.Close()

	pvdst := gocv.NewPoint2fVectorFromPoints(dst)
	defer pvdst.Close()
	return gocv.GetAffineTransform2f(pvsrc, pvdst)
}

// AffineTransform returns the result.
func AffineTransform(affineMat gocv.Mat, points []gocv.Point2f) []gocv.Point2f {
	n := len(points)
	mat := gocv.NewMatWithSize(n, 1, gocv.MatTypeCV32FC2)
	defer mat.Close()
	for i, p := range points {
		mat.SetFloatAt(i, 0, p.X)
		mat.SetFloatAt(i, 1, p.Y)
	}
	oMat := gocv.NewMat()
	defer oMat.Close()
	gocv.Transform(mat, &oMat, affineMat)
	ret := make([]gocv.Point2f, n)
	for i := 0; i < n; i++ {
		ret[i].X, ret[i].Y = oMat.GetFloatAt(i, 0), oMat.GetFloatAt(i, 1)
	}
	return ret
}

// CropRotated returns the result.
func CropRotated(img gocv.Mat, centerX, centerY, length, width float64, angle float64) gocv.Mat {
	points := imagex.RectRotateByCenter(int(centerX), int(centerY), int(length), int(width), angle)
	srcPoints := gocv.NewPointVectorFromPoints(points)
	dstPoints := gocv.NewPointVectorFromPoints([]image.Point{
		{X: 0, Y: 0},
		{X: int(length), Y: 0},
		{X: int(length), Y: int(width)},
		{X: 0, Y: int(width)},
	})
	// count perspective transform matrix
	transformMat := gocv.GetPerspectiveTransform(srcPoints, dstPoints)
	srcPoints.Close()
	dstPoints.Close()
	// wrap perspective
	dst := gocv.NewMatWithSize(int(length), int(width), img.Type())
	gocv.WarpPerspective(img, &dst, transformMat, image.Point{
		X: int(length),
		Y: int(width),
	})
	transformMat.Close()
	return dst
}

// CountNonZeroInPointsVector returns the result.
func CountNonZeroInPointsVector(img gocv.Mat, pointsVector gocv.PointsVector) int {
	mask := gocv.Zeros(img.Rows(), img.Cols(), img.Type())
	gocv.FillPoly(&mask, pointsVector, color.RGBA{255, 255, 255, 255})
	maskedImage := gocv.NewMat()
	defer maskedImage.Close()
	gocv.BitwiseAnd(img, mask, &maskedImage)
	nonZeroCount := gocv.CountNonZero(maskedImage)
	return nonZeroCount
}

// PointVectorToPointsVector returns the result.
func PointVectorToPointsVector(pointVector gocv.PointVector) gocv.PointsVector {
	return gocv.NewPointsVectorFromPoints([][]image.Point{pointVector.ToPoints()})
}

// PointsVectorFromPoints returns the result.
func PointsVectorFromPoints(points []image.Point) gocv.PointsVector {
	return gocv.NewPointsVectorFromPoints([][]image.Point{points})
}
