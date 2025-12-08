/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package slices

import (
	"math"
	"slices"

	"github.com/hopeio/gox/types/constraints"
)

// Calculate the Median of a slice of floats
func Median[S ~[]T, T constraints.Number](data S) T {
	slices.Sort(data)
	n := len(data)
	if n%2 == 0 {
		return (data[n/2-1] + data[n/2]) / 2
	}
	return data[n/2]
}

// Calculate the Mean of a slice of floats
func Mean[S ~[]T, T constraints.Number](data S) float64 {
	var sum float64
	for _, value := range data {
		sum += float64(value)
	}
	return sum / float64(len(data))
}

// Remove outliers using the MAD method and calculate the Mean of the remaining data
func RemoveOutliersMean[S ~[]T, T constraints.Number](data S) float64 {
	if len(data) == 0 {
		return 0
	}

	med := Median(data)

	// Calculate absolute deviations from the Median
	absDevs := make([]float64, len(data))
	for i, value := range data {
		absDevs[i] = math.Abs(float64(value) - float64(med))
	}

	// Calculate the Median of the absolute deviations
	mad := Median(absDevs)

	// Define a threshold using the MAD; here we use 3 times the MAD
	threshold := 3.0 * mad

	// Filter out outliers
	filteredData := make([]float64, 0)
	for i, value := range data {
		if absDevs[i] <= threshold {
			filteredData = append(filteredData, float64(value))
		}
	}

	// Calculate the Mean of the remaining data
	return Mean(filteredData)
}
