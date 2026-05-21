/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package heap

import (
	"testing"
)

// comparableInt implements cmp.Comparable[comparableInt]
// Note: Compare returns >0 when a > b, making this a max-heap by default
type comparableInt int

func (a comparableInt) Compare(b comparableInt) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func TestNewMutexHeap(t *testing.T) {
	h := New[comparableInt](10)
	_ = h
}

func TestNewFromArray(t *testing.T) {
	arr := []comparableInt{5, 3, 7, 1}
	h := NewFromArray(arr)
	h.Init()

	// container/heap is a max-heap: largest element at top
	first, ok := h.First()
	if !ok || first != 7 {
		t.Errorf("First() = %v, %v, want 7, true", first, ok)
	}
}

func TestMutexHeap_PushPop(t *testing.T) {
	h := New[comparableInt](10)
	h.Push(5)
	h.Push(3)
	h.Push(7)
	h.Push(1)

	// Max-heap: pops in descending order
	val, ok := h.Pop()
	if !ok || val != 7 {
		t.Errorf("Pop() = %v, %v, want 7, true", val, ok)
	}

	val, ok = h.Pop()
	if !ok || val != 5 {
		t.Errorf("Pop() = %v, %v, want 5, true", val, ok)
	}
}

func TestMutexHeap_First(t *testing.T) {
	h := New[comparableInt](10)
	_, ok := h.First()
	if ok {
		t.Error("First() on empty heap should return false")
	}

	h.Push(5)
	h.Push(3)
	// Max-heap: largest at top
	first, ok := h.First()
	if !ok || first != 5 {
		t.Errorf("First() = %v, %v, want 5, true", first, ok)
	}
}

func TestMutexHeap_Remove(t *testing.T) {
	h := New[comparableInt](10)
	h.Push(5)
	h.Push(3)
	h.Push(7)

	_, ok := h.Remove(0)
	if !ok {
		t.Error("Remove() should return true")
	}
}

func TestMutexHeap_Last(t *testing.T) {
	h := New[comparableInt](10)
	_, ok := h.Last()
	if ok {
		t.Error("Last() on empty heap should return false")
	}
}

func TestMutexHeap_ConcurrentPush(t *testing.T) {
	h := New[comparableInt](100)
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(n comparableInt) {
			h.Push(n)
			done <- true
		}(comparableInt(i))
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
