package stack

import "testing"

func TestNewMutexStack(t *testing.T) {
	s := NewMutexStack[int]()
	if s == nil {
		t.Fatal("NewMutexStack() returned nil")
	}
}

func TestMutexStack_PushPop(t *testing.T) {
	s := NewMutexStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	val, ok := s.Pop()
	if !ok || val != 3 {
		t.Errorf("Pop() = %v, %v, want 3, true", val, ok)
	}

	val, ok = s.Pop()
	if !ok || val != 2 {
		t.Errorf("Pop() = %v, %v, want 2, true", val, ok)
	}

	val, ok = s.Pop()
	if !ok || val != 1 {
		t.Errorf("Pop() = %v, %v, want 1, true", val, ok)
	}

	_, ok = s.Pop()
	if ok {
		t.Error("Pop() on empty stack should return false")
	}
}
