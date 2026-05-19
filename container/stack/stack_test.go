package stack

import (
	"testing"
)

func TestNew(t *testing.T) {
	s := New[int]()
	if s == nil {
		t.Error("New() returned nil")
	}
	if len(s) != 0 {
		t.Errorf("New() returned stack with len %d, want 0", len(s))
	}
}

func TestStack_PushPop(t *testing.T) {
	s := New[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	v, ok := s.Pop()
	if !ok || v != 3 {
		t.Errorf("Pop() = (%v, %v), want (3, true)", v, ok)
	}
	v, ok = s.Pop()
	if !ok || v != 2 {
		t.Errorf("Pop() = (%v, %v), want (2, true)", v, ok)
	}
	v, ok = s.Pop()
	if !ok || v != 1 {
		t.Errorf("Pop() = (%v, %v), want (1, true)", v, ok)
	}
	// Empty stack
	v, ok = s.Pop()
	if ok {
		t.Errorf("Pop() on empty stack returned ok=true, v=%v", v)
	}
}

func TestStack_String(t *testing.T) {
	s := New[string]()
	s.Push("a")
	s.Push("b")
	v, ok := s.Pop()
	if !ok || v != "b" {
		t.Errorf("Pop() = (%v, %v), want (b, true)", v, ok)
	}
}
