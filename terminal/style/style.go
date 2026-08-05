/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package style

import (
	"fmt"
	stringsx "github.com/hopeio/gox/strings"
	"strconv"
)

const baseFormat = "\x1b[%sm"
const styleFormat = "\x1b[%sm%s"
const styleWithResetFormat = styleFormat + reset
const reset = "\x1b[0m"
const color256Format = "\x1b[38;5;%dm%s"
const color256WithResetFormat = color256Format + reset
const bgColor256Format = "\x1b[48;5;%dm%s"
const bgColor256WithResetFormat = bgColor256Format + reset

// colorize returns the result.
func colorize(colorCode Style, s string) string {
	return fmt.Sprintf(styleFormat+reset, colorCode.String(), s)
}

// Styles returns the result.
func Styles(s string, styles ...Style) string {
	if len(styles) == 0 {
		return s
	}
	return fmt.Sprintf(styleWithResetFormat, stringsx.Join(styles, ";"), s)
}

type Style int

// String returns the string representation.
func (d Style) String() string {
	return strconv.Itoa(int(d))
}

// Decoration
const (
	DcReset Style = iota
	DcBold
	DcFaint
	DcItalic
	DcUnderline
	DcFlashSlow
	DcFlashRapid
	DcReverse
	DcConcealed
	DcCrossedOut
)
const (
	DcResetBold = 22 + iota
	DcResetItalic
	DcResetUnderline
	DcResetFlashing
	_
	DcResetReverse
	DcResetConcealed
	DcResetCrossedOut
)

// Color
const (
	ColorBlack Style = 30 + iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorGray
)

// HighLightColor
const (
	HLColorBlack Style = 90 + iota
	HLColorRed
	HLColorGreen
	HLColorYellow
	HLColorBlue
	HLColorMagenta
	HLColorCyan
	HLColorGray
)

// BackGround
const (
	BgColorBlack Style = 40 + iota
	BgColorRed
	BgColorGreen
	BgColorYellow
	BgColorBlue
	BgColorMagenta
	BgColorCyan
	BgColorGray
)

// HighLightBackGround
const (
	HLBgColorBlack Style = 100 + iota
	HLBgColorRed
	HLBgColorGreen
	HLBgColorYellow
	HLBgColorBlue
	HLBgColorMagenta
	HLBgColorCyan
	HLBgColorGray
)

// Blue returns the result.
func Blue(s string) string {
	return colorize(ColorBlue, s)
}

// Cyan returns the result.
func Cyan(s string) string {
	return colorize(ColorCyan, s)
}

// Magenta returns the result.
func Magenta(s string) string {
	return colorize(ColorMagenta, s)
}

// Gray returns the result.
func Gray(s string) string {
	return colorize(ColorGray, s)
}

// Red returns the result.
func Red(s string) string {
	return colorize(ColorRed, s)
}

// BgRed returns the result.
func BgRed(s string) string {
	return colorize(BgColorRed, s)
}

// Green returns the result.
func Green(s string) string {
	return colorize(ColorGreen, s)
}

// Yellow returns the result.
func Yellow(s string) string {
	return colorize(ColorYellow, s)
}

// Custom returns the result.
func Custom(s string, begin, end any) string {
	return fmt.Sprintf("\x1b[%vm%s\x1b[%vm", begin, s, end)
}

// Color256 returns the result.
func Color256(s string, c byte) string {
	return fmt.Sprintf(color256WithResetFormat, c, s)
}

// BgColor256 returns the result.
func BgColor256(s string, c byte) string {
	return fmt.Sprintf(bgColor256WithResetFormat, c, s)
}
