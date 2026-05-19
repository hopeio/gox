package maps

import (
	"sort"
	"testing"
)

func TestMap(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	result := Map(m, func(k string, v int) string {
		return k
	})
	sort.Strings(result)
	if len(result) != 3 {
		t.Errorf("Map() len = %d, want 3", len(result))
	}
}

func TestKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	keys := Keys(m)
	sort.Strings(keys)
	if len(keys) != 2 {
		t.Errorf("Keys() len = %d, want 2", len(keys))
	}
	expected := []string{"a", "b"}
	for i, k := range keys {
		if k != expected[i] {
			t.Errorf("Keys()[%d] = %v, want %v", i, k, expected[i])
		}
	}
}

func TestValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	vals := Values(m)
	sort.Ints(vals)
	if len(vals) != 2 {
		t.Errorf("Values() len = %d, want 2", len(vals))
	}
	expected := []int{1, 2}
	for i, v := range vals {
		if v != expected[i] {
			t.Errorf("Values()[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestKeysMap(t *testing.T) {
	m := map[int]string{1: "a", 2: "b"}
	result := KeysMap(m, func(k int) string {
		return string(rune('0' + k))
	})
	if len(result) != 2 {
		t.Errorf("KeysMap() len = %d, want 2", len(result))
	}
}

func TestValuesMap(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	result := ValuesMap(m, func(v int) string {
		return string(rune('0' + v))
	})
	if len(result) != 2 {
		t.Errorf("ValuesMap() len = %d, want 2", len(result))
	}
}

func TestForEach(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	sum := 0
	ForEach(m, func(k string, v int) {
		sum += v
	})
	if sum != 3 {
		t.Errorf("ForEach sum = %d, want 3", sum)
	}
}

func TestForEachValue(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	sum := 0
	ForEachValue(m, func(v int) {
		sum += v
	})
	if sum != 3 {
		t.Errorf("ForEachValue sum = %d, want 3", sum)
	}
}

func TestForEachKey(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	keys := []string{}
	ForEachKey(m, func(k string) {
		keys = append(keys, k)
	})
	if len(keys) != 2 {
		t.Errorf("ForEachKey count = %d, want 2", len(keys))
	}
}

func TestMerge(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 3, "c": 4}
	merged := Merge(m1, m2)
	if len(merged) != 3 {
		t.Errorf("Merge() len = %d, want 3", len(merged))
	}
	if merged["b"] != 3 {
		t.Errorf("Merge() b = %d, want 3 (last wins)", merged["b"])
	}
}

func TestTransform(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	result := Transform[map[string]int, string, int, map[int]string, int, string](m, func(k string, v int) (int, string) {
		return v, k
	})
	if result[1] != "a" {
		t.Errorf("Transform()[1] = %v, want a", result[1])
	}
	if result[2] != "b" {
		t.Errorf("Transform()[2] = %v, want b", result[2])
	}
}

func TestMultiKeys(t *testing.T) {
	m1 := map[string]int{"a": 1}
	m2 := map[string]int{"b": 2}
	keys := MultiKeys(m1, m2)
	sort.Strings(keys)
	if len(keys) != 2 {
		t.Errorf("MultiKeys() len = %d, want 2", len(keys))
	}
}

func TestMultiValues(t *testing.T) {
	m1 := map[string]int{"a": 1}
	m2 := map[string]int{"b": 2}
	vals := MultiValues(m1, m2)
	sort.Ints(vals)
	if len(vals) != 2 {
		t.Errorf("MultiValues() len = %d, want 2", len(vals))
	}
}
