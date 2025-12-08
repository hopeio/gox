/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package math

import (
	constraintsi "github.com/hopeio/gox/types/constraints"
)

func StandardDeviation[S ~[]T, T constraintsi.Number](data S, isSample bool) float64 {
	n := float64(len(data))
	var sum float64
	for v := range data {
		sum += float64(v)
	}
	mean := sum / n

	var varianceSum float64
	for _, v := range data {
		varianceSum += (float64(v) - mean) * (float64(v) - mean)
	}
	if isSample {
		return varianceSum / (n - 1)
	}
	return varianceSum / n
}

func Variance[S ~[]T, T constraintsi.Number](data S, isSample bool) float64 {
	n := float64(len(data))
	if n == 0 {
		return 0
	}

	// 计算均值
	var sum float64
	for _, v := range data {
		sum += float64(v)
	}
	mean := sum / n

	// 计算平方和
	var varianceSum float64
	for _, v := range data {
		varianceSum += (float64(v) - mean) * (float64(v) - mean)
	}

	// 总体或样本方差
	if isSample {
		return varianceSum / (n - 1)
	}
	return varianceSum / n
}

func Max[T constraintsi.Ordered](data ...T) T {
	max := data[0]
	n := len(data)
	for i := 1; i < n; i++ {
		if data[i] > max {
			max = data[i]
		}
	}
	return max
}

func Min[T constraintsi.Ordered](data ...T) T {
	min := data[0]
	n := len(data)
	for i := 1; i < n; i++ {
		if data[i] < min {
			min = data[i]
		}
	}
	return min
}

func MinAndMax[T constraintsi.Ordered](data ...T) (T, T) {
	min, max := data[0], data[0]
	n := len(data)
	for i := 1; i < n; i++ {
		if data[i] > max {
			max = data[i]
		}
		if data[i] < min {
			min = data[i]
		}
	}
	return min, max
}
