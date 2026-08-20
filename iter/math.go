/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package iter

import (
	"iter"
	"math"
	"slices"

	"github.com/hopeio/gox/cmp"
	"github.com/hopeio/gox/types"
	constraintsx "github.com/hopeio/gox/types/constraints"
	"golang.org/x/exp/constraints"
)

// SumComparable returns the result.
func SumComparable[T constraints.Ordered](seq iter.Seq[T]) T {
	var result T
	for v := range seq {
		result += v
	}
	return result
}

// Returns the sum of all the elements in the iterator.
func Sum[T constraintsx.Number](it iter.Seq[T]) T {
	return Fold(it, 0, func(a, b T) T {
		return a + b
	})
}

// Returns the sum of all the elements in the iterator.
func SumCount[T constraintsx.Number](it iter.Seq[T]) (T, int) {
	var count int
	return Fold(it, 0, func(a, b T) T {
		count++
		return a + b
	}), count
}

// Returns the product of all the elements in the iterator.
func Product[T constraintsx.Number](it iter.Seq[T]) T {
	return Fold(it, 1, func(a, b T) T {
		return a * b
	})
}

// Returns the average of all the elements in the iterator.
func Average[T constraintsx.Number](it iter.Seq[T]) T {
	return Fold(Enumerate(it), *new(T), func(result T, item types.Pair[int, T]) T {
		return result + (T(item.Second)-result)/T(item.First+1)
	})
}

// Return the maximum value of all elements of the iterator.
func Max[T constraints.Ordered](it iter.Seq[T]) (T, bool) {
	return Reduce(it, func(a T, b T) T {
		if a > b {
			return a
		} else {
			return b
		}
	})
}

// Return the maximum value of all elements of the iterator.
// greater reports whether a should be ordered before b (i.e. a < b).
func MaxBy[T any](it iter.Seq[T], greater cmp.LessFunc[T]) (T, bool) {
	return Reduce(it, func(a T, b T) T {
		if greater(a, b) {
			return b
		}
		return a
	})
}

// Return the minimum value of all elements of the iterator.
func Min[T constraints.Ordered](it iter.Seq[T]) (T, bool) {
	return Reduce(it, func(a T, b T) T {
		if a < b {
			return a
		} else {
			return b
		}
	})
}

// Calculate the Mean of a slice of floats
func Mean[T constraintsx.Number](seq iter.Seq[T]) float64 {
	var sum float64
	var count int
	for value := range seq {
		sum += float64(value)
		count++
	}
	return sum / float64(count)
}

// Return the minimum value of all elements of the iterator by the less function.
// less reports whether a should be ordered before b (i.e. a < b).
func MinBy[T any](it iter.Seq[T], less cmp.LessFunc[T]) (T, bool) {
	return Reduce(it, func(a T, b T) T {
		if less(a, b) {
			return a
		}
		return b
	})
}

// Return the maximum value of all elements of the iterator. Numbers are compared directly.
func MaxByComparable[T constraints.Ordered](it iter.Seq[T]) (T, bool) {
	return Max(it)
}

// Return the minimum value of all elements of the iterator. Numbers are compared directly.
func MinByComparable[T constraints.Ordered](it iter.Seq[T]) (T, bool) {
	return Min(it)
}

// Calculate the Median of a sequence. Materializes the elements.
func Median[T constraintsx.Number](seq iter.Seq[T]) (T, bool) {
	var data []T
	for v := range seq {
		data = append(data, v)
	}
	n := len(data)
	if n == 0 {
		return *new(T), false
	}
	slices.Sort(data)
	if n%2 == 1 {
		return data[n/2], true
	}
	return (data[n/2-1] + data[n/2]) / 2, true
}

// Calculate the Population standard deviation of a sequence.
func StdDev[T constraintsx.Number](seq iter.Seq[T]) (float64, bool) {
	var sum float64
	var count int
	buf := make([]float64, 0, 16)
	for v := range seq {
		f := float64(v)
		sum += f
		count++
		buf = append(buf, f)
	}
	if count == 0 {
		return 0, false
	}
	mean := sum / float64(count)
	var variance float64
	for _, f := range buf {
		d := f - mean
		variance += d * d
	}
	return math.Sqrt(variance / float64(count)), true
}
