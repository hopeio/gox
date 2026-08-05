/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package types

type Ref[T any] struct {
	value *T
}

// RefOf ...
func RefOf[T any](v *T) Ref[T] {
	return Ref[T]{v}
}

// Val ...
func (a Ref[T]) Val() (v T, ok bool) {
	if a.value == nil {
		return
	}
	return *a.value, true
}

// Get ...
func (a Ref[T]) Get() T {
	return *a.value
}

// Set ...
func (a Ref[T]) Set(v T) T {
	var old = *a.value
	*a.value = v
	return old
}

// IsNil reports whether the condition holds.
func (a Ref[T]) IsNil() bool {
	return a.value == nil
}

// IsNotNil reports whether the condition holds.
func (a Ref[T]) IsNotNil() bool {
	return a.value != nil
}
