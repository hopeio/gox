/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package color

import "image/color"

// ColorToGray ...
func ColorToGray(c color.Color) color.Gray {
	if g, ok := c.(color.Gray); ok {
		return g
	}
	r, g, b, _ := c.RGBA()

	// These coefficients (the fractions 0.299, 0.587 and 0.114) are the same
	// as those given by the JFIF specification and used by func RGBToYCbCr in
	// ycbcr.go.
	//
	// Note that 19595 + 38470 + 7471 equals 65536.
	//
	// The 24 is 16 + 8. The 16 is the same as used in RGBToYCbCr. The 8 is
	// because the return value is 8 bit color, not 16 bit color.
	y := (19595*r + 38470*g + 7471*b + 1<<15) >> 24

	return color.Gray{Y: uint8(y)}
}

// RGBAToGray ...
func RGBAToGray(c color.RGBA) color.Gray {
	y := (19595*uint32(c.R)<<8 + 38470*uint32(c.G)<<8 + 7471*uint32(c.B)<<8 + 1<<15) >> 24
	return color.Gray{Y: uint8(y)}
}

// RGBToGray ...
func RGBToGray(c RGB) color.Gray {
	y := (19595*uint32(c.R)<<8 + 38470*uint32(c.G)<<8 + 7471*uint32(c.B)<<8 + 1<<15) >> 24
	return color.Gray{Y: uint8(y)}
}

// ColorToGray2 ...
func ColorToGray2(c color.Color) color.Gray {
	r, g, b, _ := c.RGBA()
	return color.Gray{Y: uint8(0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8))}
}

// ColorToRGBu8 ...
func ColorToRGBu8(c color.Color) RGB {
	r32, g32, b32, _ := c.RGBA()
	return RGB{uint8(r32), uint8(g32), uint8(b32)}
}
