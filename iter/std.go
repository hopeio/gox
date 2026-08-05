/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package iter

import (
	"iter"

	"github.com/hopeio/gox/types"
	constraintsx "github.com/hopeio/gox/types/constraints"
)

// SliceAll ...
func SliceAll[S ~[]T, T any](input S) Seq[types.Pair[int, T]] {
	return func(yield func(types.Pair[int, T]) bool) {
		for i, v := range input {
			if !yield(types.PairOf(i, v)) {
				return
			}
		}
	}
}

// SliceAllValues ...
func SliceAllValues[S ~[]T, T any](input S) Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range input {
			if !yield(v) {
				return
			}
		}
	}
}

// SliceBackwardValues ...
func SliceBackwardValues[S ~[]T, T any](input S) Seq[T] {
	return func(yield func(T) bool) {
		n := len(input) - 1
		for i := n; n > 0; n-- {
			if !yield(input[i]) {
				return
			}
		}
	}
}

// SliceBackward ...
func SliceBackward[S ~[]T, T any](input S) Seq[types.Pair[int, T]] {
	return func(yield func(types.Pair[int, T]) bool) {
		n := len(input) - 1
		for i := n; n > 0; n-- {
			if !yield(types.PairOf(i, input[i])) {
				return
			}
		}
	}
}

// HashMapAll reports whether the condition holds.
func HashMapAll[M ~map[K]V, K comparable, V any](m M) Seq[types.Pair[K, V]] {
	return func(yield func(types.Pair[K, V]) bool) {
		for k, v := range m {
			if !yield(types.PairOf(k, v)) {
				return
			}
		}
	}
}

// StringAll ...
func StringAll[T ~string](input T) Seq[types.Pair[int, rune]] {
	return func(yield func(types.Pair[int, rune]) bool) {
		for i, v := range input {
			if !yield(types.PairOf(i, v)) {
				return
			}
		}
	}
}

// StringAll2 ...
func StringAll2[T ~string](input T) iter.Seq2[int, rune] {
	return func(yield func(int, rune) bool) {
		for i, v := range input {
			if !yield(i, v) {
				return
			}
		}
	}
}

// StringRunes ...
func StringRunes[T ~string](input T) Seq[rune] {
	return func(yield func(rune) bool) {
		for _, v := range input {
			if !yield(v) {
				return
			}
		}
	}
}

// ChannelAll ...
func ChannelAll[C ~chan T, T any](c C) Seq[T] {
	return func(yield func(T) bool) {
		for v := range c {
			if !yield(v) {
				return
			}
		}
	}
}

// ChannelAll2 ...
func ChannelAll2[C ~chan T, T any](c C) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		var count int
		for v := range c {
			if !yield(count, v) {
				return
			}
			count++
		}
	}
}

// RangeAll ...
func RangeAll[T constraintsx.Number](begin, end, step T) Seq[T] {
	return func(yield func(T) bool) {
		for v := begin; v <= end; v += step {
			if !yield(v) {
				return
			}
		}
	}
}

// RangeAll2 ...
func RangeAll2[T constraintsx.Number](begin, end, step T) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		var count int
		for v := begin; v <= end; v += step {
			if !yield(count, v) {
				return
			}
			count++
		}
	}
}
