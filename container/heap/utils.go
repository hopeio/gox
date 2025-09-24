/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package heap

import (
	"github.com/hopeio/gox/cmp"
)

func Init[T cmp.Comparable[T]](heap []T) {
	// heapify
	n := len(heap)
	for i := n/2 - 1; i >= 0; i-- {
		Down(heap, i, n)
	}
}

func Down[T cmp.Comparable[T]](heap []T, i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && heap[j2].Compare(heap[j1]) > 0 {
			j = j2 // = 2*i + 2  // right child
		}
		if heap[j].Compare(heap[i]) <= 0 {
			break
		}
		heap[i], heap[j] = heap[j], heap[i]
		i = j
	}
	return i > i0
}

func Up[T cmp.Comparable[T]](heap []T, j int) {

	for {
		i := (j - 1) / 2 // parent
		if i == j || heap[j].Compare(heap[i]) <= 0 {
			break
		}
		heap[i], heap[j] = heap[j], heap[i]
		j = i
	}
}

func Fix[T cmp.Comparable[T]](heap []T, i int) {
	if !Down(heap, i, len(heap)) {
		Up(heap, i)
	}
}

// 与Down一致，不同的写法
func AdjustDown[T cmp.Comparable[T]](heap []T, i int) {
	child := leftChild(i)
	for child < len(heap) {
		if child+1 < len(heap) && heap[child+1].Compare(heap[child]) > 0 {
			child++
		}
		if heap[child].Compare(heap[i]) <= 0 {
			break
		}
		heap[i], heap[child] = heap[child], heap[i]
		i = child
		child = leftChild(i)
	}
}

// 与Up一致，不同的写法
func AdjustUp[T cmp.Comparable[T]](heap []T, i int) {
	p := parent(i)
	for p >= 0 && heap[i].Compare(heap[p]) > 0 {
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
