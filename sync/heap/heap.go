/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package heap

import (
	"sync"

	"github.com/hopeio/gox/cmp"
	"github.com/hopeio/gox/container/heap"
)

type MutexHeap[T cmp.Comparable[T]] struct {
	mu   sync.RWMutex
	data []T
	zero T
}

// New creates a new instance.
// 返回指针：MutexHeap 含锁，按值返回会引导调用方拷贝锁。
func New[T cmp.Comparable[T]](l int) *MutexHeap[T] {
	return &MutexHeap[T]{
		data: make([]T, 0, l),
	}
}

// NewFromArray creates a heap that takes ownership of arr and heapifies it.
func NewFromArray[T cmp.Comparable[T]](arr []T) *MutexHeap[T] {
	h := &MutexHeap[T]{
		data: arr,
	}
	heap.Init(h.data)
	return h
}

// First performs the operation.
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

// Init performs the operation.
func (h *MutexHeap[T]) Init() {
	h.mu.Lock()
	heap.Init(h.data)
	h.mu.Unlock()
}

// Push updates or inserts a value.
func (h *MutexHeap[T]) Push(x T) {
	h.mu.Lock()
	h.data = append(h.data, x)
	heap.Up(h.data, len(h.data)-1)
	h.mu.Unlock()
}

// Pop removes or resets state.
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

// Last performs the operation.
func (h *MutexHeap[T]) Last() (T, bool) {
	h.mu.Lock()
	if len(h.data) == 0 {
		h.mu.Unlock()
		return h.zero, false
	}
	last := h.data[len(h.data)-1]
	h.mu.Unlock()
	return last, true
}

// Remove removes or resets state.
func (h *MutexHeap[T]) Remove(i int) (T, bool) {
	h.mu.Lock()
	if i < 0 || i >= len(h.data) {
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
