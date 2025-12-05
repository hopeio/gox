/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package image

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"runtime"
	"strconv"
	"testing"

	runtimex "github.com/hopeio/gox/runtime"
)

func TestMerge(t *testing.T) {
	dir := `D:\work\`
	suff := "--1.jpg"
	fovs := [][]int{{0, 1, 2}, {5, 4, 3}, {6, 7, 8}, {11, 10, 9}, {12, 13, 14}, {17, 16, 15}, {18, 19, 20}}
	points := [][]int{{238125, 262125}, {245275, 276609}}
	scale := float64(5120) / float64(32000)
	sWidth, sHeight := 5120, 5120
	// 拼图

	var horizontalOverlaps, verticalOverlaps []int
	var imgs = make([][]image.Image, len(fovs))
	if len(fovs) > 0 {
		horizontalOverlaps = make([]int, len(fovs[0])-1)
		verticalOverlaps = make([]int, len(fovs)-1)
		imgs[0] = make([]image.Image, len(fovs[0]))
	}

	predictOverlap := sHeight - int(scale*float64(points[1][1]-points[1][0]))
	var gray1, gray2 []uint8
	for i := 1; i < len(fovs); i++ {
		idx1, idx2 := fovs[i-1][0], fovs[i][0]
		data1, _ := os.Open(dir + strconv.Itoa(idx1) + suff)
		data2, _ := os.Open(dir + strconv.Itoa(idx2) + suff)
		minOverlap := max(1, predictOverlap-100)
		maXOverlap := min(predictOverlap+100, sHeight/2)
		if i == 1 {
			gray1 = make([]uint8, maXOverlap*sHeight)
			gray2 = make([]uint8, maXOverlap*sHeight)
		}

		img1, _ := jpeg.Decode(data1)
		img2, _ := jpeg.Decode(data2)
		verticalOverlaps[i-1] = CalculateOverlapReuseMemory(img1, img2, true, minOverlap, maXOverlap, gray1, gray2)

		data1.Close()
		data2.Close()
		imgs[i] = make([]image.Image, len(fovs[0]))
	}

	row := fovs[0]
	predictOverlap = sWidth - int(scale*float64(points[0][1]-points[0][0]))
	for i := 1; i < len(row); i++ {
		idx1, idx2 := row[i-1], row[i]
		data1, _ := os.Open(dir + strconv.Itoa(idx1) + suff)
		data2, _ := os.Open(dir + strconv.Itoa(idx2) + suff)
		minOverlap := max(1, predictOverlap-100)
		maXOverlap := min(predictOverlap+100, sWidth/2)
		if i == 1 {
			gray1 = make([]uint8, maXOverlap*sWidth)
			gray2 = make([]uint8, maXOverlap*sWidth)
		}
		img1, _ := jpeg.Decode(data1)
		img2, _ := jpeg.Decode(data2)
		horizontalOverlaps[i-1] = CalculateOverlapReuseMemory(img1, img2, false, minOverlap, maXOverlap, gray1, gray2)

		data1.Close()
		data2.Close()
	}
	gray1, gray2 = nil, nil
	runtime.GC()
	t.Logf("overlap %v , %v", horizontalOverlaps, verticalOverlaps)
	runtimex.PrintMemoryUsage(1)

	for i := range fovs {
		for j := range fovs[i] {
			data, _ := os.Open(dir + strconv.Itoa(fovs[i][j]) + suff)
			imgs[i][j], _ = jpeg.Decode(data)
		}
	}

	mi := NewMergeImage(imgs, sWidth, sHeight, horizontalOverlaps, verticalOverlaps)
	runtimex.PrintMemoryUsage(2)
	outFile, err := os.Create(dir + "panel.jpg")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	jpeg.Encode(outFile, mi, nil)
	outFile.Close()

	runtimex.PrintMemoryUsage(3)
}
