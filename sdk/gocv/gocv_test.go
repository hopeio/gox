/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */


package gocv

import (
	"fmt"
	"gocv.io/x/gocv"
	"testing"
)

func TestAffineMatrix(t *testing.T) {
	p1, p2, p3, q1, q2, q3 := gocv.Point2f{2000, 7000}, gocv.Point2f{48000, 80000}, gocv.Point2f{2000, 85000}, gocv.Point2f{3558, 17895}, gocv.Point2f{11016, 5997}, gocv.Point2f{3538, 5182}
	affineMat := AffineMat([]gocv.Point2f{p1, p2, p3}, []gocv.Point2f{q1, q2, q3})
	t.Log(AffineTransform(affineMat, []gocv.Point2f{{X: 48000, Y: 13000}}))

	affineMat = AffineMat([]gocv.Point2f{{128.08328, 13.295279}, {123.16628, 24.473278}, {110.23628, 17.256279}}, []gocv.Point2f{{26.525, 10.1625}, {24.475, 21.3}, {9, 14}})
	fmt.Println(affineMat.Type())
	for i := range affineMat.Rows() {
		for j := range affineMat.Cols() {
			fmt.Print(affineMat.GetDoubleAt(i, j), " ")
		}
		fmt.Println()
	}
	t.Log(AffineTransform(affineMat, []gocv.Point2f{{X: 128.08328, Y: 13.295279}}))
	// 定义要变换的点 (例如 [100, 150])
	point := [3]float32{128.08328, 13.295279, 1}

	// 应用仿射变换矩阵到点
	transformedPoint := [2]float32{
		affineMat.GetFloatAt(0, 0)*point[0] + affineMat.GetFloatAt(0, 1)*point[1] + affineMat.GetFloatAt(0, 2),
		affineMat.GetFloatAt(1, 0)*point[0] + affineMat.GetFloatAt(1, 1)*point[1] + affineMat.GetFloatAt(1, 2),
	}

	// 打印变换后的点
	fmt.Printf("Transformed point: [%f, %f]\n", transformedPoint[0], transformedPoint[1])
}
