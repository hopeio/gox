/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package heap

import (
	"github.com/hopeio/gox/cmp"
)

func Init[T any](heap []T, cmp cmp.CompareFunc[T]) {
	// heapify
	n := len(heap)
	for i := n/2 - 1; i >= 0; i-- {
		Down(heap, i, n, cmp)
	}
}

func Down[T any](heap []T, i0, n int, cmp cmp.CompareFunc[T]) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && cmp(heap[j2], heap[j1]) < 0 {
			j = j2 // = 2*i + 2  // right child
		}
		if cmp(heap[j], heap[i]) >= 0 {
			break
		}
		heap[i], heap[j] = heap[j], heap[i]
		i = j
	}
	return i > i0
}

func Up[T any](heap []T, j int, cmp cmp.CompareFunc[T]) {

	for {
		i := (j - 1) / 2 // parent
		if i == j || cmp(heap[j], heap[i]) >= 0 {
			break
		}
		heap[i], heap[j] = heap[j], heap[i]
		j = i
	}
}

func Fix[T any](heap []T, i int, cmp cmp.CompareFunc[T]) {
	if !Down(heap, i, len(heap), cmp) {
		Up(heap, i, cmp)
	}
}

// 与Down一致，不同的写法
func AdjustDown[T any](heap []T, i int, cmp cmp.CompareFunc[T]) {
	child := leftChild(i)
	for child < len(heap) {
		if child+1 < len(heap) && cmp(heap[child+1], heap[child]) < 0 {
			child++
		}
		if cmp(heap[child], heap[i]) >= 0 {
			break
		}
		heap[i], heap[child] = heap[child], heap[i]
		i = child
		child = leftChild(i)
	}
}

// 与Up一致，不同的写法
func AdjustUp[T any](heap []T, i int, cmp cmp.CompareFunc[T]) {
	p := parent(i)
	for p >= 0 && cmp(heap[i], heap[p]) < 0 {
		heap[i], heap[p] = heap[p], heap[i]
		i = p
		p = parent(i)
	}
}

func parent(i int) int {
	return (i - 1) / 2
}
func leftChild(i int) int {
	return i*2 + 1
}
