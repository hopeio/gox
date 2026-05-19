package set

import (
	"testing"
)

func TestNew(t *testing.T) {
	s := New[int]()
	if s == nil {
		t.Error("New() returned nil")
	}
	if len(s) != 0 {
		t.Errorf("New() returned set with len %d, want 0", len(s))
	}
}

func TestSet_Add(t *testing.T) {
	s := New[string]()
	s.Add("a")
	s.Add("b")
	s.Add("a") // duplicate
	if len(s) != 2 {
		t.Errorf("after Add a,b,a: len = %d, want 2", len(s))
	}
}

func TestSet_Contains(t *testing.T) {
	s := New[int]()
	s.Add(1)
	s.Add(2)

	if !s.Contains(1) {
		t.Error("Contains(1) = false, want true")
	}
	if !s.Contains(2) {
		t.Error("Contains(2) = false, want true")
	}
	if s.Contains(3) {
		t.Error("Contains(3) = true, want false")
	}
}

func TestSet_Remove(t *testing.T) {
	s := New[int]()
	s.Add(1)
	s.Add(2)
	s.Remove(1)
	if s.Contains(1) {
		t.Error("after Remove(1), Contains(1) = true, want false")
	}
	if !s.Contains(2) {
		t.Error("after Remove(1), Contains(2) = false, want true")
	}
	// Remove non-existent key should not panic
	s.Remove(999)
}

func TestSet_ToSlice(t *testing.T) {
	s := New[int]()
	s.Add(10)
	s.Add(20)
	slice := s.ToSlice()
	if len(slice) != 2 {
		t.Errorf("ToSlice() len = %d, want 2", len(slice))
	}
	m := map[int]bool{10: true, 20: true}
	for _, v := range slice {
		if !m[v] {
			t.Errorf("ToSlice() contains unexpected value %d", v)
		}
	}
}
