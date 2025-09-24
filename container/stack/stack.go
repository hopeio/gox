/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package stack

// Stack is a FILO stack
type Stack[T any] []T

// New returns a new stack
func New[T any]() Stack[T] {
	return make([]T, 0)
}

// Push pushes a value to the stack
func (s *Stack[T]) Push(v T) {
	*s = append(*s, v)
}

// Pop pops the top value out of the stack
func (s *Stack[T]) Pop() (T, bool) {
	h := *s
	if len(h) == 0 {
		return *new(T), false
	}
	v := h[len(h)]
	*s = h[:len(h)-1]
	return v, true
}
