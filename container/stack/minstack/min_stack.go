/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package minstack

import (
	"github.com/hopeio/gox/cmp"
	"github.com/hopeio/gox/container/list"
)

// MinStack defines a type.
type MinStack[T any] struct {
	store *list.List[node[T]]
	less  cmp.LessFunc[T]
}

type node[T any] struct {
	value T
	min   T
}

// NewMinStack creates a new instance.
func NewMinStack[T any](less cmp.LessFunc[T]) MinStack[T] {
	return MinStack[T]{store: list.New[node[T]](), less: less}
}

// Push updates or inserts a value.
func (ms *MinStack[T]) Push(x T) {
	if ms.store.Head() != nil && ms.less(ms.store.Head().Value.min, x) {
		ms.store.PushFront(node[T]{value: x, min: ms.store.Head().Value.min})
	} else {
		ms.store.PushFront(node[T]{value: x, min: x})
	}
}

// Pop removes or resets state.
func (ms *MinStack[T]) Pop() (T, bool) {
	node, ok := ms.store.Pop()
	if !ok {
		return *new(T), false
	}
	return node.value, true
}

// Top converts the value; returns the zero value on an empty stack instead of panicking.
func (ms *MinStack[T]) Top() T {
	var zero T
	if h := ms.store.Head(); h != nil {
		return h.Value.value
	}
	return zero
}

// GetMin returns the value; returns the zero value on an empty stack instead of panicking.
func (ms *MinStack[T]) GetMin() T {
	var zero T
	if h := ms.store.Head(); h != nil {
		return h.Value.min
	}
	return zero
}
