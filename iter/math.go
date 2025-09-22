/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package iter

import (
	"github.com/hopeio/gox/cmp"
	"github.com/hopeio/gox/types"
	constraintsx "github.com/hopeio/gox/types/constraints"
	"golang.org/x/exp/constraints"
	"iter"
)

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
func MaxBy[T any](it iter.Seq[T], greater cmp.LessFunc[T]) (T, bool) {
	return Reduce(it, func(a T, b T) T {
		if greater(a, b) {
			return a
		} else {
			return b
		}
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
