/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package minstack

import (
	"testing"

	"github.com/hopeio/gox/cmp"
)

func TestMinStack_PushPop(t *testing.T) {
	ms := NewMinStack[int](cmp.Less[int])
	ms.Push(5)
	ms.Push(3)
	ms.Push(7)
	ms.Push(1)

	top := ms.Top()
	if top != 1 {
		t.Errorf("Top() = %v, want 1", top)
	}

	min := ms.GetMin()
	if min != 1 {
		t.Errorf("GetMin() = %v, want 1", min)
	}

	val, ok := ms.Pop()
	if !ok || val != 1 {
		t.Errorf("Pop() = %v, %v, want 1, true", val, ok)
	}

	min = ms.GetMin()
	if min != 3 {
		t.Errorf("GetMin() after pop = %v, want 3", min)
	}
}

func TestMinStack_GetMin(t *testing.T) {
	ms := NewMinStack[int](cmp.Less[int])
	ms.Push(5)
	if ms.GetMin() != 5 {
		t.Errorf("GetMin() = %v, want 5", ms.GetMin())
	}
	ms.Push(3)
	if ms.GetMin() != 3 {
		t.Errorf("GetMin() = %v, want 3", ms.GetMin())
	}
	ms.Push(7)
	if ms.GetMin() != 3 {
		t.Errorf("GetMin() = %v, want 3", ms.GetMin())
	}
	ms.Push(1)
	if ms.GetMin() != 1 {
		t.Errorf("GetMin() = %v, want 1", ms.GetMin())
	}
}

func TestMinStack_PopEmpty(t *testing.T) {
	ms := NewMinStack[int](cmp.Less[int])
	_, ok := ms.Pop()
	if ok {
		t.Error("Pop() on empty stack should return false")
	}
}

func TestMinStack_Sequence(t *testing.T) {
	ms := NewMinStack[int](cmp.Less[int])
	// Push: 2, 1, 3, 1
	ms.Push(2)
	ms.Push(1)
	ms.Push(3)
	ms.Push(1)

	if ms.GetMin() != 1 {
		t.Errorf("GetMin() = %v, want 1", ms.GetMin())
	}

	// Pop 1 -> min should still be 1
	ms.Pop()
	if ms.GetMin() != 1 {
		t.Errorf("GetMin() after pop = %v, want 1", ms.GetMin())
	}

	// Pop 3 -> min should be 1
	ms.Pop()
	if ms.GetMin() != 1 {
		t.Errorf("GetMin() after pops = %v, want 1", ms.GetMin())
	}

	// Pop 1 -> min should be 2
	ms.Pop()
	if ms.GetMin() != 2 {
		t.Errorf("GetMin() after pops = %v, want 2", ms.GetMin())
	}
}
