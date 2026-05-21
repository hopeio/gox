/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package geom

import (
	"math"
)

func AngleRadian(angleDegrees float64) float64 {
	return angleDegrees * math.Pi / 180.0
}
func AngleDegrees(angleRadian float64) float64 {
	return angleRadian / math.Pi * 180.0
}

func NormalizeAngleRadian(theta float64) float64 {
	if theta == 0 {
		return 0
	}
	pi2 := math.Pi * 2
	if theta > 0 {
		for theta >= pi2 {
			theta -= pi2
		}
	} else {
		theta += pi2
		for theta < 0 {
			theta += pi2
		}
	}
	return theta
}

func NormalizeAngleDegrees(theta float64) float64 {
	if theta == 0 {
		return 0
	}

	if theta > 0 {
		for theta > 360 {
			theta -= 360
		}
	} else {
		theta += 360
		for theta < 0 {
			theta += 360
		}
	}
	return theta
}
