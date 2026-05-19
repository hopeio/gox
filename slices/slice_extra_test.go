package slices

import (
	"slices"
	"sort"
	"testing"
)

func TestEvery(t *testing.T) {
	s := []int{2, 4, 6}
	if !Every(s, func(v int) bool { return v%2 == 0 }) {
		t.Error("Every() on all even should be true")
	}
	s2 := []int{2, 3, 6}
	if Every(s2, func(v int) bool { return v%2 == 0 }) {
		t.Error("Every() on mixed should be false")
	}
}

func TestSome(t *testing.T) {
	s := []int{1, 3, 4}
	if !Some(s, func(v int) bool { return v%2 == 0 }) {
		t.Error("Some() with one even should be true")
	}
	s2 := []int{1, 3, 5}
	if Some(s2, func(v int) bool { return v%2 == 0 }) {
		t.Error("Some() with no even should be false")
	}
}

func TestZip(t *testing.T) {
	s1 := []int{1, 2, 3}
	s2 := []int{4, 5, 6}
	result := Zip(s1, s2)
	if len(result) != 3 {
		t.Errorf("Zip() len = %d, want 3", len(result))
	}
	if result[0] != [2]int{1, 4} {
		t.Errorf("Zip()[0] = %v, want [1,4]", result[0])
	}
}

func TestDeduplicate(t *testing.T) {
	s := []int{1, 2, 2, 3, 3, 3}
	result := Deduplicate(s)
	if len(result) != 3 {
		t.Errorf("Deduplicate() len = %d, want 3", len(result))
	}
	sort.Ints(result)
	expected := []int{1, 2, 3}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Deduplicate()[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

func TestToMap(t *testing.T) {
	type item struct {
		ID   int
		Name string
	}
	s := []item{{1, "a"}, {2, "b"}}
	m := ToMap(s, func(i item) (int, string) { return i.ID, i.Name })
	if m[1] != "a" {
		t.Errorf("ToMap()[1] = %q, want %q", m[1], "a")
	}
	if m[2] != "b" {
		t.Errorf("ToMap()[2] = %q, want %q", m[2], "b")
	}
}

func TestClassify(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	m := Classify(s, func(v int) (string, int) {
		if v%2 == 0 {
			return "even", v
		}
		return "odd", v
	})
	if len(m["even"]) != 3 {
		t.Errorf("Classify()[\"even\"] len = %d, want 3", len(m["even"]))
	}
	if len(m["odd"]) != 3 {
		t.Errorf("Classify()[\"odd\"] len = %d, want 3", len(m["odd"]))
	}
}

func TestMap(t *testing.T) {
	s := []int{1, 2, 3}
	result := Map(s, func(v int) string {
		return string(rune('0' + v))
	})
	if len(result) != 3 {
		t.Errorf("Map() len = %d, want 3", len(result))
	}
	if result[0] != "1" {
		t.Errorf("Map()[0] = %q, want %q", result[0], "1")
	}
}

func TestFilter(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	result := Filter(func(v int) bool { return v%2 == 0 }, s)
	if len(result) != 2 {
		t.Errorf("Filter() len = %d, want 2", len(result))
	}
	if result[0] != 2 || result[1] != 4 {
		t.Errorf("Filter() = %v, want [2,4]", result)
	}
}

func TestReduce(t *testing.T) {
	s := []int{1, 2, 3, 4}
	result := Reduce(s, func(a, b int) int { return a + b })
	if result != 10 {
		t.Errorf("Reduce() = %d, want 10", result)
	}
}

func TestCopy(t *testing.T) {
	s := []int{1, 2, 3}
	c := Copy(s)
	if len(c) != len(s) {
		t.Errorf("Copy() len = %d, want %d", len(c), len(s))
	}
	c[0] = 99
	if s[0] == 99 {
		t.Error("Copy() should not modify original")
	}
}

func TestGroupBy(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	m := GroupBy(s, func(v int) string {
		if v%2 == 0 {
			return "even"
		}
		return "odd"
	})
	if len(m["even"]) != 3 {
		t.Errorf("GroupBy()[\"even\"] len = %d, want 3", len(m["even"]))
	}
	if len(m["odd"]) != 3 {
		t.Errorf("GroupBy()[\"odd\"] len = %d, want 3", len(m["odd"]))
	}
}

func TestSum(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	if Sum(s) != 15 {
		t.Errorf("Sum() = %d, want 15", Sum(s))
	}
}

func TestRemove(t *testing.T) {
	s := []int{1, 2, 3, 4}
	result := Remove(s, 2) // remove index 2
	if len(result) != 3 {
		t.Errorf("Remove() len = %d, want 3", len(result))
	}
	expected := []int{1, 2, 4}
	if !slices.Equal(result, expected) {
		t.Errorf("Remove() = %v, want %v", result, expected)
	}
}

func TestToPtrs(t *testing.T) {
	s := []int{1, 2, 3}
	ptrs := ToPtrs(s)
	if len(ptrs) != 3 {
		t.Errorf("ToPtrs() len = %d, want 3", len(ptrs))
	}
	if *ptrs[0] != 1 || *ptrs[1] != 2 || *ptrs[2] != 3 {
		t.Errorf("ToPtrs() values incorrect")
	}
}

func TestCollector(t *testing.T) {
	c := Collector[[]int, int]{}
	builder := c.Builder()
	c.Append(builder, 1)
	c.Append(builder, 2)
	c.Append(builder, 3)
	result := c.Finish(builder)
	if len(result) != 3 {
		t.Errorf("Collector result len = %d, want 3", len(result))
	}
	if result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("Collector result = %v, want [1,2,3]", result)
	}
}

// FilterPlace already tested in slice_test.go

func TestForEach(t *testing.T) {
	s := []int{1, 2, 3}
	sum := 0
	ForEach(s, func(idx int, v int) { sum += v })
	if sum != 6 {
		t.Errorf("ForEach() sum = %d, want 6", sum)
	}
}

func TestForEachValue(t *testing.T) {
	s := []int{1, 2, 3}
	sum := 0
	ForEachValue(s, func(v int) { sum += v })
	if sum != 6 {
		t.Errorf("ForEachValue() sum = %d, want 6", sum)
	}
}

// HasCoincide already tested in set_test.go

func TestIntersection(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{2, 3, 4}
	result := Intersection(a, b)
	if len(result) != 2 {
		t.Errorf("Intersection() len = %d, want 2", len(result))
	}
	sort.Ints(result)
	if result[0] != 2 || result[1] != 3 {
		t.Errorf("Intersection() = %v, want [2,3]", result)
	}
}

func TestUnion(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{2, 3, 4}
	result := Union(a, b)
	if len(result) != 4 {
		t.Errorf("Union() len = %d, want 4", len(result))
	}
}

func TestDifferenceSet(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{2, 3, 4}
	result := DifferenceSet(a, b)
	sort.Ints(result)
	if len(result) != 1 || result[0] != 1 {
		t.Errorf("DifferenceSet() = %v, want [1]", result)
	}
}

// Difference already tested in set_test.go

func TestRemoveDuplicates(t *testing.T) {
	s := []int{1, 2, 2, 3, 3}
	result := RemoveDuplicates(s)
	if len(result) != 3 {
		t.Errorf("RemoveDuplicates() len = %d, want 3", len(result))
	}
}

func TestHasCoincide_LargeArray(t *testing.T) {
	// Test with arrays >= SmallArrayLen(32) that use the map-based algorithm
	// The map algorithm adds s1 elements then checks s2 elements concurrently.
	// For overlap to be detected, s2 elements must appear after s1 elements
	// are already in the map.
	a := make([]int, 100)
	b := make([]int, 100)
	for i := range a {
		a[i] = i // 0..99
		b[i] = i // 0..99 - exact overlap
	}
	if !HasCoincide(a, b) {
		t.Error("HasCoincide with identical large arrays should be true")
	}
}
