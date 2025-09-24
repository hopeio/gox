/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package heap

type Interface[T any] interface {
	~[]T
	Less(i, j int) bool
}

type Heap[E any, I Interface[E]] []E

func New[E any, I Interface[E]](capacity int) Heap[E, I] {
	return make([]E, 0, capacity)
}

func NewFromArray[E any, I Interface[E]](arr I) Heap[E, I] {
	heap := Heap[E, I](arr)
	for i := 1; i < len(arr); i++ {
		heap.up(i)
	}
	return heap
}

func (heap Heap[E, I]) Init() {
	// heapify
	n := len(heap)
	for i := n/2 - 1; i >= 0; i-- {
		heap.down(i, n)
	}
}

func (heap *Heap[E, I]) Push(x E) {
	h := *heap
	*heap = append(h, x)
	heap.up(len(h))
}

func (heap *Heap[E, I]) Put(val E) {
	h := *heap
	if len(h) < cap(h) {
		*heap = append(h, val)
		for i := 1; i < len(h); i++ {
			heap.up(i)
		}
		return
	}
	temp := I{h[0], val}
	if temp.Less(1, 0) {
		return
	}
	h[0] = val
	heap.down(0, len(h))
}

func (heap *Heap[E, I]) Pop() (E, bool) {
	h := *heap
	if len(h) == 0 {
		return *new(E), false
	}
	n := len(h) - 1
	item := h[0]
	h[0], h[n] = h[n], h[0]
	heap.down(0, n)
	*heap = h[:n]
	return item, true
}

func (heap Heap[E, I]) First() (E, bool) {
	if len(heap) == 0 {
		return *new(E), false
	}
	return heap[0], true
}

func (heap Heap[E, I]) Last() (E, bool) {
	if len(heap) == 0 {
		return *new(E), false
	}
	return heap[len(heap)-1], true
}

func (heap *Heap[E, I]) Remove(i int) (E, bool) {
	h := *heap
	if len(h) <= i {
		return *new(E), false
	}
	n := len(h) - 1
	item := h[i]
	if n != i {
		h[i], h[n] = h[n], h[i]
		if !heap.down(i, n) {
			heap.up(i)
		}
	}
	*heap = h[:n]
	return item, true
}

func (heap Heap[E, I]) up(j int) {

	for {
		i := (j - 1) / 2 // parent
		if i == j || !I(heap).Less(j, i) {
			break
		}
		heap[i], heap[j] = heap[j], heap[i]
		j = i
	}
}

func (heap Heap[E, I]) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && I(heap).Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !I(heap).Less(j, i) {
			break
		}
		heap[i], heap[j] = heap[j], heap[i]
		i = j
	}
	return i > i0
}

func (heap Heap[E, I]) Size() int {
	return len(heap)
}
