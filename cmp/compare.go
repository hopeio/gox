/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package cmp

import (
	"unsafe"

	"golang.org/x/exp/constraints"
)

// Less ...
func Less[T constraints.Ordered](a T, b T) bool {
	return a < b
}

// Greater ...
func Greater[T constraints.Ordered](a T, b T) bool {
	return a > b
}

// Equal ...
func Equal[T comparable](a T, b T) bool {
	return a == b
}

// Compare ...
func Compare[T constraints.Ordered](x, y T) int {
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}

type GTValue[T constraints.Ordered] struct {
	Value T
}

// Compare ...
func (a GTValue[T]) Compare(b GTValue[T]) bool {
	return a.Value > b.Value
}

type LTValue[T constraints.Ordered] struct {
	Value T
}

// Compare ...
func (a LTValue[T]) Compare(b GTValue[T]) bool {
	return a.Value < b.Value
}

// SignedFlip ...
func SignedFlip[T constraints.Signed](i T) T {
	if i < 0 && i == T(-1<<(unsafe.Sizeof(i)-1)) {
		return 1<<unsafe.Sizeof(i) - 1
	}
	return -i
}

// UnSignedFlip ...
func UnSignedFlip[T constraints.Unsigned](i T) T {
	return 1<<unsafe.Sizeof(i) - 1 - i
}

// FloatFlip ...
func FloatFlip[T constraints.Float](i T) T {
	if isNaN(i) {
		return i
	}
	return -i
}

// isNaN reports whether the condition holds.
func isNaN[T constraints.Ordered](x T) bool {
	return x != x
}
