/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package heap

import (
	"github.com/hopeio/gox/cmp"
)

type Heap[T any] struct {
	arr  []T
	cmp  cmp.CompareFunc[T]
	zero T
}

func New[T any](l int, cmp cmp.CompareFunc[T]) *Heap[T] {
	return &Heap[T]{
		arr: make([]T, 0, l),
		cmp: cmp,
	}
}

func NewFromArray[T any](arr []T, cmp cmp.CompareFunc[T]) *Heap[T] {
	heap := &Heap[T]{
		arr: arr,
		cmp: cmp,
	}
	for i := 1; i < len(arr); i++ {
		heap.up(i)
	}
	return heap
}

func (h *Heap[T]) Init() {
	// heapify
	n := len(h.arr)
	for i := n/2 - 1; i >= 0; i-- {
		h.down(i, n)
	}
}

// 当达到堆预设大小时会增加堆的大小
func (h *Heap[T]) Push(x T) {
	h.arr = append(h.arr, x)
	h.up(len(h.arr) - 1)
}

// 不会改变预设堆的大小
func (h *Heap[T]) Put(val T) {
	if len(h.arr) < cap(h.arr) {
		h.arr = append(h.arr, val)
		for i := 1; i < len(h.arr); i++ {
			h.up(i)
		}
		return
	}
	if h.cmp(val, h.arr[0]) < 0 {
		return
	}
	h.arr[0] = val
	h.down(0, len(h.arr))
}

func (h *Heap[T]) Pop() (T, bool) {
	if len(h.arr) == 0 {
		return h.zero, false
	}
	n := len(h.arr) - 1
	item := h.arr[0]
	h.arr[0], h.arr[n] = h.arr[n], h.arr[0]
	h.down(0, n)
	h.arr = h.arr[:n]
	return item, true
}

func (h *Heap[T]) Remove(i int) (T, bool) {
	if len(h.arr) == 0 {
		return h.zero, false
	}
	n := len(h.arr) - 1
	item := h.arr[i]
	if n != i {
		h.arr[i], h.arr[n] = h.arr[n], h.arr[i]
		if !h.down(i, n) {
			h.up(i)
		}
	}
	h.arr = h.arr[:n]
	return item, true
}

func (h *Heap[T]) down(i0, n int) bool {
	return Down(h.arr, i0, n, h.cmp)
}

func (h *Heap[T]) up(j int) {
	Up(h.arr, j, h.cmp)
}

func (h *Heap[T]) fix(i int) {
	Fix(h.arr, i, h.cmp)
}

func (h *Heap[T]) First() (T, bool) {
	if len(h.arr) == 0 {
		return *new(T), false
	}
	return h.arr[0], true
}

func (h Heap[T]) Last() (T, bool) {
	if len(h.arr) == 0 {
		return *new(T), false
	}
	return h.arr[len(h.arr)-1], false
}
