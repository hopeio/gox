/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package qrcode

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
)

const (
	black = 30
	white = 47
)

// ConsolePrint ...
func ConsolePrint(qrCode image.Image) {
	// 将二维码转换为控制台颜色代码
	pixels := convertQRCodeToConsolePixels(qrCode)

	// 在控制台显示二维码
	buffer := bytes.Buffer{}
	for _, row := range pixels {
		for _, pixel := range row {
			if pixel == white {
				buffer.WriteString("\033[47m   \033[0m")
			} else {
				buffer.WriteString("\033[40m   \033[0m")
			}
		}
		buffer.WriteString("\n")
	}
	fmt.Println(buffer.String())

}

// convertQRCodeToConsolePixels ...
func convertQRCodeToConsolePixels(qr image.Image) [][]int {
	bounds := qr.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	pixels := make([][]int, height)
	for y := 0; y < height; y++ {
		pixels[y] = make([]int, width)
		for x := 0; x < width; x++ {
			if qr.At(x, y) == color.Black {
				pixels[y][x] = black
			} else {
				pixels[y][x] = white
			}
		}
	}
	return pixels
}
