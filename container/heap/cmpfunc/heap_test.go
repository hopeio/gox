/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package heap

import (
	"testing"
)

func intCmp(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func TestNew(t *testing.T) {
	h := New[int](10, intCmp)
	if h == nil {
		t.Fatal("New() returned nil")
	}
}

func TestHeap_PushPop(t *testing.T) {
	h := New[int](10, intCmp)
	h.Push(5)
	h.Push(3)
	h.Push(7)
	h.Push(1)

	val, ok := h.Pop()
	if !ok || val != 1 {
		t.Errorf("Pop() = %v, %v, want 1, true", val, ok)
	}

	val, ok = h.Pop()
	if !ok || val != 3 {
		t.Errorf("Pop() = %v, %v, want 3, true", val, ok)
	}

	val, ok = h.Pop()
	if !ok || val != 5 {
		t.Errorf("Pop() = %v, %v, want 5, true", val, ok)
	}

	val, ok = h.Pop()
	if !ok || val != 7 {
		t.Errorf("Pop() = %v, %v, want 7, true", val, ok)
	}

	_, ok = h.Pop()
	if ok {
		t.Error("Pop() on empty heap should return false")
	}
}

func TestNewFromArray(t *testing.T) {
	arr := []int{5, 3, 7, 1}
	h := NewFromArray(arr, intCmp)

	val, ok := h.Pop()
	if !ok || val != 1 {
		t.Errorf("Pop() = %v, %v, want 1, true", val, ok)
	}
}

func TestHeap_First(t *testing.T) {
	h := New[int](10, intCmp)
	_, ok := h.First()
	if ok {
		t.Error("First() on empty heap should return false")
	}

	h.Push(5)
	h.Push(3)
	first, ok := h.First()
	if !ok || first != 3 {
		t.Errorf("First() = %v, %v, want 3, true", first, ok)
	}
}

func TestHeap_Remove(t *testing.T) {
	h := New[int](10, intCmp)
	h.Push(5)
	h.Push(3)
	h.Push(7)

	_, ok := h.Remove(0)
	if !ok {
		t.Error("Remove() should return true")
	}
}

func TestHeap_Put(t *testing.T) {
	h := New[int](3, intCmp)
	h.Push(5)
	h.Push(3)
	h.Push(7)

	// Capacity reached, Put should replace min if val > min
	h.Put(4)
	first, _ := h.First()
	// After Put(4) with capacity full, 3 should be replaced by 4
	if first != 4 {
		t.Errorf("First() after Put = %v, want 4", first)
	}

	// Put smaller value should be ignored
	h.Put(2)
	first, _ = h.First()
	if first != 4 {
		t.Errorf("First() after Put(smaller) = %v, want 4", first)
	}
}

func TestDown(t *testing.T) {
	heap := []int{7, 3, 5, 1}
	swapped := Down(heap, 0, len(heap), intCmp)
	if !swapped {
		t.Error("Down() should return true when swaps occurred")
	}
}

func TestUp(t *testing.T) {
	heap := []int{1, 3, 5, 0}
	Up(heap, 3, intCmp)
	if heap[0] != 0 {
		t.Errorf("After Up, heap[0] = %v, want 0", heap[0])
	}
}

func TestFix(t *testing.T) {
	heap := []int{1, 3, 5, 7}
	heap[0] = 10
	Fix(heap, 0, intCmp)
	// After fix, 10 should move down
	if heap[0] == 10 {
		t.Error("Fix should have moved 10 down")
	}
}

func TestInit(t *testing.T) {
	heap := []int{7, 3, 5, 1}
	Init(heap, intCmp)
	// After init, the minimum should be at root
	if heap[0] != 1 {
		t.Errorf("After Init, heap[0] = %v, want 1", heap[0])
	}
}

func TestAdjustDown(t *testing.T) {
	heap := []int{7, 1, 5, 3}
	AdjustDown(heap, 0, intCmp)
	if heap[0] != 1 {
		t.Errorf("After AdjustDown, heap[0] = %v, want 1", heap[0])
	}
}

func TestAdjustUp(t *testing.T) {
	heap := []int{1, 3, 5, 0}
	AdjustUp(heap, 3, intCmp)
	if heap[0] != 0 {
		t.Errorf("After AdjustUp, heap[0] = %v, want 0", heap[0])
	}
}
