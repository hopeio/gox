package set

import (
	"testing"

	"github.com/hopeio/gox/types"
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

func TestSet_LenClear(t *testing.T) {
	s := New[int]()
	s.Add(1)
	s.Add(2)
	s.Add(3)
	if s.Len() != 3 {
		t.Errorf("Len() = %d, want 3", s.Len())
	}
	s.Clear()
	if s.Len() != 0 {
		t.Errorf("after Clear Len() = %d, want 0", s.Len())
	}
}

func TestSet_ForEachAllAny(t *testing.T) {
	s := New[int]()
	s.Add(2)
	s.Add(4)
	s.Add(6)

	var sum int
	s.ForEach(types.Consumer[int](func(v int) { sum += v }))
	if sum != 12 {
		t.Errorf("ForEach sum = %d, want 12", sum)
	}
	if !s.All(types.Predicate[int](func(v int) bool { return v%2 == 0 })) {
		t.Error("All(even) = false, want true")
	}
	if s.All(types.Predicate[int](func(v int) bool { return v > 4 })) {
		t.Error("All(>4) = true, want false")
	}
	if !s.Any(types.Predicate[int](func(v int) bool { return v == 6 })) {
		t.Error("Any(==6) = false, want true")
	}
	if s.Any(types.Predicate[int](func(v int) bool { return v == 9 })) {
		t.Error("Any(==9) = true, want false")
	}
}

func TestSet_Filter(t *testing.T) {
	s := New[int]()
	s.Add(1)
	s.Add(2)
	s.Add(3)
	s.Add(4)
	even := s.Filter(types.Predicate[int](func(v int) bool { return v%2 == 0 }))
	if even.Len() != 2 || !even.Contains(2) || !even.Contains(4) {
		t.Errorf("Filter(even) = %v, want {2,4}", even.ToSlice())
	}
}

func TestSet_Map(t *testing.T) {
	s := New[int]()
	s.Add(1)
	s.Add(2)
	s.Add(3)
	// Map to string keys (method-level generic: R independent of K)
	strs := s.Map(types.UnaryFunction[int, string](func(v int) string { return string(rune('a' + v - 1)) }))
	if strs.Len() != 3 || !strs.Contains("a") || !strs.Contains("b") || !strs.Contains("c") {
		t.Errorf("Map -> %v, want {a,b,c}", strs.ToSlice())
	}

	// MapToSlice (arbitrary R)
	doubled := s.MapToSlice(types.UnaryFunction[int, int](func(v int) int { return v * 2 }))
	if len(doubled) != 3 {
		t.Errorf("MapToSlice len = %d, want 3", len(doubled))
	}

	// MapToMap
	m := s.MapToMap(types.UnaryFunction[int, types.Pair[int, string]](func(v int) types.Pair[int, string] {
		return types.PairOf(v, string(rune('A'+v-1)))
	}))
	if len(m) != 3 || m[1] != "A" || m[3] != "C" {
		t.Errorf("MapToMap = %v, want {1:A 2:B 3:C}", m)
	}
}

func TestSet_SetOps(t *testing.T) {
	a := New[int]()
	a.Add(1)
	a.Add(2)
	a.Add(3)
	b := New[int]()
	b.Add(3)
	b.Add(4)

	u := a.Union(b)
	if u.Len() != 4 {
		t.Errorf("Union len = %d, want 4", u.Len())
	}
	inter := a.Intersect(b)
	if inter.Len() != 1 || !inter.Contains(3) {
		t.Errorf("Intersect = %v, want {3}", inter.ToSlice())
	}
	diff := a.Difference(b)
	if diff.Len() != 2 || !diff.Contains(1) || !diff.Contains(2) {
		t.Errorf("Difference = %v, want {1,2}", diff.ToSlice())
	}
}

func TestSet_Seq(t *testing.T) {
	s := New[int]()
	s.Add(1)
	s.Add(2)
	count := 0
	for v := range s.Seq() {
		if !s.Contains(v) {
			t.Errorf("Seq yielded %d not in set", v)
		}
		count++
	}
	if count != 2 {
		t.Errorf("Seq count = %d, want 2", count)
	}
}
