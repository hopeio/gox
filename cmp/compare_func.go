/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package cmp

type LessFunc[T any] func(T, T) bool

type EqualFunc[T any] func(T, T) bool

type CompareFunc[T any] func(T, T) int

func (c CompareFunc[T]) LessFunc() LessFunc[T] {
	return func(a, b T) bool {
		return c(a, b) < 0
	}
}
