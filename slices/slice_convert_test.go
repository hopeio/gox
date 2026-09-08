package slices

import (
	"reflect"
	"testing"
)

func TestCanReinterpretRejectsString(t *testing.T) {
	if canReinterpret(reflect.TypeOf("")) {
		t.Fatal("string must not be reinterpreted: it contains a Go pointer")
	}
	type withString struct {
		S string
		N int
	}
	if canReinterpret(reflect.TypeOf(withString{})) {
		t.Fatal("struct with string field must not be reinterpreted")
	}
	if !canReinterpret(reflect.TypeOf(int64(0))) {
		t.Fatal("plain int64 should allow reinterpret")
	}
}

func TestConvertIdenticalString(t *testing.T) {
	in := []string{"a", "b"}
	out := Convert[[]string, []string](in)
	if len(out) != 2 || out[0] != "a" || out[1] != "b" {
		t.Fatalf("got %v", out)
	}
}

func TestConvertSameSizedNumeric(t *testing.T) {
	type convInt int64
	in := []convInt{1, 2, 3}
	out := Convert[[]convInt, []int64](in)
	if len(out) != 3 || out[0] != 1 || out[2] != 3 {
		t.Fatalf("got %v", out)
	}
}
