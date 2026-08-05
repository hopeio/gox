/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package math

import (
	"math"
	"strconv"
)

// DecimalPlaces returns the result.
func DecimalPlaces(value float64, prec int) float64 {
	multiplier := math.Pow(10, float64(prec))
	return float64(int(value*multiplier)) / multiplier
}

// DecimalPlacesRound returns the result.
func DecimalPlacesRound(value float64, rank int) float64 {
	multiplier := math.Pow(10, float64(rank))
	return math.Round(value*multiplier) / multiplier
}

// FormatFloat formats or converts the value.
func FormatFloat(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 64)
}
