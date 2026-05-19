package list

import (
	"testing"
)

func TestNew(t *testing.T) {
	l := New[int]()
	if l == nil {
		t.Error("New() returned nil")
	}
	if l.Len() != 0 {
		t.Errorf("Len() = %d, want 0", l.Len())
	}
}

func TestList_Push(t *testing.T) {
	l := New[int]()
	l.Push(1)
	l.Push(2)
	l.Push(3)
	if l.Len() != 3 {
		t.Errorf("Len() = %d, want 3", l.Len())
	}
}

func TestList_FirstLast(t *testing.T) {
	l := New[int]()
	// Empty list
	_, ok := l.First()
	if ok {
		t.Error("First() on empty list returned ok=true")
	}
	_, ok = l.Last()
	if ok {
		t.Error("Last() on empty list returned ok=true")
	}

	l.Push(10)
	l.Push(20)
	l.Push(30)

	v, ok := l.First()
	if !ok || v != 10 {
		t.Errorf("First() = (%v, %v), want (10, true)", v, ok)
	}
	v, ok = l.Last()
	if !ok || v != 30 {
		t.Errorf("Last() = (%v, %v), want (30, true)", v, ok)
	}
}

func TestList_PushFront(t *testing.T) {
	l := New[int]()
	l.PushFront(1)
	l.PushFront(2)
	l.PushFront(3)

	v, ok := l.First()
	if !ok || v != 3 {
		t.Errorf("First() after PushFront = (%v, %v), want (3, true)", v, ok)
	}
	v, ok = l.Last()
	if !ok || v != 1 {
		t.Errorf("Last() after PushFront = (%v, %v), want (1, true)", v, ok)
	}
}

func TestList_Pop(t *testing.T) {
	l := New[int]()
	l.Push(1)
	l.Push(2)
	l.Push(3)

	v, ok := l.Pop()
	if !ok || v != 1 {
		t.Errorf("Pop() = (%v, %v), want (1, true)", v, ok)
	}
	v, ok = l.Pop()
	if !ok || v != 2 {
		t.Errorf("Pop() = (%v, %v), want (2, true)", v, ok)
	}
	if l.Len() != 1 {
		t.Errorf("Len() = %d, want 1", l.Len())
	}
	// Pop last element
	v, ok = l.Pop()
	if !ok || v != 3 {
		t.Errorf("Pop() = (%v, %v), want (3, true)", v, ok)
	}
	if l.Len() != 0 {
		t.Errorf("Len() = %d, want 0", l.Len())
	}
	// Pop on empty list
	_, ok = l.Pop()
	if ok {
		t.Error("Pop() on empty list returned ok=true")
	}
}

func TestList_PushAt(t *testing.T) {
	l := New[int]()
	l.Push(1)
	l.Push(3)
	// PushAt(1, 2) inserts after node at position 1
	// With [1, 3], node at idx=1 is 3, result: [1, 3, 2]
	l.PushAt(1, 2)

	if l.Len() != 3 {
		t.Errorf("Len() = %d, want 3", l.Len())
	}

	// Pop order: 1, 3, 2
	v, ok := l.Pop()
	if !ok || v != 1 {
		t.Errorf("Pop() = (%v, %v), want (1, true)", v, ok)
	}
	v, ok = l.Pop()
	if !ok || v != 3 {
		t.Errorf("Pop() = (%v, %v), want (3, true)", v, ok)
	}
	v, ok = l.Pop()
	if !ok || v != 2 {
		t.Errorf("Pop() = (%v, %v), want (2, true)", v, ok)
	}

	// PushAt beginning (idx=0 delegates to PushFront)
	l2 := New[int]()
	l2.Push(2)
	l2.PushAt(0, 1)
	v, ok = l2.First()
	if !ok || v != 1 {
		t.Errorf("First() after PushAt(0,1) = (%v, %v), want (1, true)", v, ok)
	}

	// PushAt end (idx=size delegates to Push)
	l3 := New[int]()
	l3.Push(1)
	l3.PushAt(1, 2)
	v, ok = l3.Last()
	if !ok || v != 2 {
		t.Errorf("Last() after PushAt(1,2) = (%v, %v), want (2, true)", v, ok)
	}
}

func TestList_PushAt_Panic(t *testing.T) {
	l := New[int]()
	defer func() {
		if r := recover(); r == nil {
			t.Error("PushAt with out-of-range index should panic")
		}
	}()
	l.PushAt(-1, 1)
}
