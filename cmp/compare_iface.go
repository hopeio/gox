/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package cmp

import (
	"cmp"
)

type Comparable[T any] interface {
	Compare(T) int
}

// Deprecated: use cmp.Comparable
type CompareLess[T any] interface {
	Less(T) bool
}

type ComparableIdx interface {
	Compare(i, j int) int
}

// Deprecated: use cmp.Comparable
type CompareIdxLess interface {
	Less(i, j int) bool
}

type CompareEqual[T any] interface {
	Equal(T) bool
}

type EqualKey[T comparable] interface {
	EqualKey() T
}

type CompareKey[T cmp.Ordered] interface {
	CompareKey() T
}
