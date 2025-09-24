/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package heap

import (
	"github.com/hopeio/gox/cmp"
	"github.com/hopeio/gox/container/heap"
	"sync"
)

type MutexHeap[T cmp.Comparable[T]] struct {
	mu   sync.RWMutex
	data []T
	zero T
}

func New[T cmp.Comparable[T]](l int) MutexHeap[T] {
	return MutexHeap[T]{
		data: make([]T, 0, l),
	}
}

func NewFromArray[T cmp.Comparable[T]](arr []T) MutexHeap[T] {
	return MutexHeap[T]{
		data: arr,
	}
}
func (h *MutexHeap[T]) First() (T, bool) {
	h.mu.RLock()
	if len(h.data) == 0 {
		h.mu.RUnlock()
		return h.zero, false
	}
	first := h.data[0]
	h.mu.RUnlock()
	return first, true
}

func (h *MutexHeap[T]) Init() {
	h.mu.Lock()
	heap.Init(h.data)
	h.mu.Unlock()
}

func (h *MutexHeap[T]) Push(x T) {
	h.mu.Lock()
	h.data = append(h.data, x)
	heap.Up(h.data, len(h.data)-1)
	h.mu.Unlock()
}

func (h *MutexHeap[T]) Pop() (T, bool) {
	h.mu.Lock()
	if len(h.data) == 0 {
		h.mu.Unlock()
		return h.zero, false
	}
	n := len(h.data) - 1
	item := h.data[0]
	h.data[0], h.data[n] = h.data[n], h.data[0]
	heap.Down(h.data, 0, n)
	h.data = h.data[:n]
	h.mu.Unlock()
	return item, true
}

func (h MutexHeap[T]) Last() (T, bool) {
	h.mu.Lock()
	if len(h.data) == 0 {
		h.mu.Unlock()
		return h.zero, false
	}
	last := h.data[len(h.data)-1]
	h.mu.Unlock()
	return last, false
}

func (h *MutexHeap[T]) Remove(i int) (T, bool) {
	h.mu.Lock()
	if len(h.data) == 0 {
		h.mu.Unlock()
		return h.zero, false
	}
	n := len(h.data) - 1
	item := h.data[i]
	if n != i {
		h.data[i], h.data[n] = h.data[n], h.data[i]
		if !heap.Down(h.data, i, n) {
			heap.Up(h.data, i)
		}
	}
	h.data = h.data[:n]
	h.mu.Unlock()
	return item, true
}
