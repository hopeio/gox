/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package style

import (
	"bufio"
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
)

// 字符
const rbgcFormat = "\x1b[38;2;%d;%d;%dm%c"
const rbgcWithResetFormat = rbgcFormat + reset
const rbgcBgFormat = "\x1b[48;2;%d;%d;%dm%c"
const rbgcBgWithResetFormat = rbgcBgFormat + reset

// 字符串
const rbgsFormat = "\x1b[38;2;%d;%d;%dm%s"
const rbgsWithResetFormat = rbgsFormat + reset
const rbgsBgFormat = "\x1b[48;2;%d;%d;%dm%s"
const rbgsBgWithResetFormat = rbgsBgFormat + reset

type colorRGB struct {
	r, g, b uint16
}

// Format ...
func (c colorRGB) Format(s string) string {
	return fmt.Sprintf(rbgsWithResetFormat, c.r, c.g, c.b, s)
}

// NewRGBColor creates and returns a new instance.
func NewRGBColor(r, g, b byte) colorRGB {
	return colorRGB{r: uint16(r), g: uint16(g), b: uint16(b)}
}

// RGB ...
func RGB(s string, r, g, b byte) string {
	return fmt.Sprintf(rbgsWithResetFormat, r, g, b, s)
}

// BgRGB ...
func BgRGB(s string, r, g, b byte) string {
	return fmt.Sprintf(rbgsBgWithResetFormat, r, g, b, s)
}

// Gradient ...
func Gradient(text string, begin, end colorRGB) string {
	var colorText []string
	for i, r := range text {
		var ratio = float64(i) / float64(len(text)-1)
		var red = byte(math.Round(float64(begin.r) + float64(end.r-begin.r)*ratio))
		var green = byte(math.Round(float64(begin.g) + float64(end.g-begin.g)*ratio))
		var blue = byte(math.Round(float64(begin.b) + float64(end.b-begin.b)*ratio))
		colorText = append(colorText, fmt.Sprintf(rbgcFormat, red, green, blue, r))
	}
	colorText = append(colorText, reset)
	return strings.Join(colorText, "")
}

// GradientRandom ...
func GradientRandom(text string) string {
	var begin = colorRGB{r: uint16(rand.N(byte(255))), g: uint16(rand.N(byte(255))), b: uint16(rand.N(byte(255)))}
	var end = colorRGB{r: uint16(rand.N(byte(255))), g: uint16(rand.N(byte(255))), b: uint16(rand.N(byte(255)))}
	return Gradient(text, begin, end)
}

// GradientMultiLine ...
func GradientMultiLine(text string, begin, end colorRGB) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	var colorText []string
	for scanner.Scan() {
		colorText = append(colorText, Gradient(scanner.Text(), begin, end))
	}
	return strings.Join(colorText, "\n")
}

// GradientMultiLineRandom ...
func GradientMultiLineRandom(text string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	var colorText []string
	for scanner.Scan() {
		var begin = colorRGB{r: uint16(rand.N(byte(255))), g: uint16(rand.N(byte(255))), b: uint16(rand.N(byte(255)))}
		var end = colorRGB{r: uint16(rand.N(byte(255))), g: uint16(rand.N(byte(255))), b: uint16(rand.N(byte(255)))}
		colorText = append(colorText, Gradient(scanner.Text(), begin, end))
	}
	return strings.Join(colorText, "\n")
}

var (
	RainbowRGB       = [...]colorRGB{RainbowRedRGB, RainbowOrangeRGB, RainbowYellowRGB, RainbowGreenRGB, RainbowCyanRGB, RainbowBlueRGB, RainbowPurpleRGB}
	RainbowRedRGB    = NewRGBColor(255, 0, 0)
	RainbowOrangeRGB = NewRGBColor(255, 165, 0)
	RainbowYellowRGB = NewRGBColor(255, 255, 0)
	RainbowGreenRGB  = NewRGBColor(0, 255, 0)
	RainbowCyanRGB   = NewRGBColor(0, 255, 255)
	RainbowBlueRGB   = NewRGBColor(0, 0, 255)
	RainbowPurpleRGB = NewRGBColor(128, 0, 128)
)

// Rainbow ...
func Rainbow(text string) string {
	var colorText []string
	var n int
	for _, r := range text {
		color := RainbowRGB[n]
		colorText = append(colorText, fmt.Sprintf(rbgcFormat, color.r, color.g, color.b, r))
		n++
		if n == 7 {
			n = 0
		}
	}
	colorText = append(colorText, reset)
	return strings.Join(colorText, "")
}

// RainbowMultiLine ...
func RainbowMultiLine(text string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	var colorText []string
	var n int
	for scanner.Scan() {
		color := RainbowRGB[n]
		colorText = append(colorText, color.Format(scanner.Text()))
		n++
		if n == 7 {
			n = 0
		}
	}
	colorText = append(colorText, reset)
	return strings.Join(colorText, "\n")
}

// TODO
func rainbowGradient(text string) {

}

// rgbToAnsi256 ...
func rgbToAnsi256(r, g, b byte) byte {
	// We use the extended greyscale palette here, with the exception of
	// black and white. normal palette only has 4 greyscale shades.
	if r>>4 == g>>4 && g>>4 == b>>4 {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return byte(math.Round(float64(((r-8)/247)*24)) + 232)
	}
	var ansi = 16 +
		36*math.Round(float64((r/255)*5)) +
		6*math.Round(float64((g/255)*5)) +
		math.Round(float64((b/255)*5))
	return byte(ansi)
}
