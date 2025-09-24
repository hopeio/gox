/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package heap

import (
	"github.com/hopeio/gox/cmp"
)

type Heap[T cmp.Comparable[T]] []T

func New[T cmp.Comparable[T]](l int) Heap[T] {
	return make([]T, 0, l)
}

func NewFromArray[T cmp.Comparable[T]](arr []T) Heap[T] {
	heap := Heap[T](arr)
	for i := 1; i < len(arr); i++ {
		Up(heap, i)
	}
	return heap
}

func (heap Heap[T]) Init() {
	Init(heap)
}

func (heap *Heap[T]) Push(x T) {
	h := *heap
	h = append(h, x)
	Up(h, len(h)-1)
	*heap = h
}

func (heap *Heap[T]) Pop() (T, bool) {
	h := *heap
	if len(h) == 0 {
		return *new(T), false
	}
	n := len(h) - 1
	item := h[0]
	h[0], h[n] = h[n], h[0]
	Down(h, 0, n)
	*heap = h[:n]
	return item, true
}

func (heap Heap[T]) First() (T, bool) {
	if len(heap) == 0 {
		return *new(T), false
	}
	return heap[0], true
}

func (heap Heap[T]) Last() (T, bool) {
	if len(heap) == 0 {
		return *new(T), false
	}
	return heap[len(heap)-1], false
}

func (heap *Heap[T]) Remove(i int) (T, bool) {
	h := *heap
	if len(h) == 0 {
		return *new(T), false
	}
	n := len(h) - 1
	item := h[i]
	if n != i {
		h[i], h[n] = h[n], h[i]
		if !Down(h, i, n) {
			Up(h, i)
		}
	}
	*heap = h[:n]
	return item, true
}
