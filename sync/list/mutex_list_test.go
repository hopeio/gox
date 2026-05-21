package list

import "testing"

func TestNewMutexList(t *testing.T) {
	l := NewMutexList[int]()
	if l == nil {
		t.Fatal("NewMutexList() returned nil")
	}
	if l.Len() != 0 {
		t.Errorf("Len() = %d, want 0", l.Len())
	}
}

func TestMutexList_PushPop(t *testing.T) {
	l := NewMutexList[int]()
	l.Push(1)
	l.Push(2)
	l.Push(3)

	if l.Len() != 3 {
		t.Errorf("Len() = %d, want 3", l.Len())
	}

	val, ok := l.Pop()
	if !ok || val != 1 {
		t.Errorf("Pop() = %v, %v, want 1, true", val, ok)
	}

	val, ok = l.Pop()
	if !ok || val != 2 {
		t.Errorf("Pop() = %v, %v, want 2, true", val, ok)
	}

	if l.Len() != 1 {
		t.Errorf("Len() = %d, want 1", l.Len())
	}
}

func TestMutexList_PopEmpty(t *testing.T) {
	l := NewMutexList[int]()
	_, ok := l.Pop()
	if ok {
		t.Error("Pop() on empty list should return false")
	}
}

func TestMutexList_SingleElement(t *testing.T) {
	l := NewMutexList[string]()
	l.Push("hello")

	val, ok := l.Pop()
	if !ok || val != "hello" {
		t.Errorf("Pop() = %v, %v, want hello, true", val, ok)
	}

	if l.Len() != 0 {
		t.Errorf("Len() after pop = %d, want 0", l.Len())
	}
}
