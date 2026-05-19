package gox

import (
	"testing"
)

func TestTernaryOperator(t *testing.T) {
	if TernaryOperator(true, 1, 2) != 1 {
		t.Error("TernaryOperator(true, 1, 2) should be 1")
	}
	if TernaryOperator(false, 1, 2) != 2 {
		t.Error("TernaryOperator(false, 1, 2) should be 2")
	}
	// Test with strings
	if TernaryOperator(true, "a", "b") != "a" {
		t.Error("TernaryOperator(true, \"a\", \"b\") should be \"a\"")
	}
}

func TestMatch(t *testing.T) {
	if Match(true, 1, 2) != 1 {
		t.Error("Match(true, 1, 2) should be 1")
	}
	if Match(false, 1, 2) != 2 {
		t.Error("Match(false, 1, 2) should be 2")
	}
}

func TestPointer(t *testing.T) {
	p := Pointer(42)
	if p == nil || *p != 42 {
		t.Errorf("Pointer(42) = %v, want pointer to 42", p)
	}
}

func TestZero(t *testing.T) {
	z := Zero[int]()
	if z != 0 {
		t.Errorf("Zero[int]() = %v, want 0", z)
	}
	z2 := Zero[string]()
	if z2 != "" {
		t.Errorf("Zero[string]() = %q, want empty string", z2)
	}
	z3 := Zero[bool]()
	if z3 != false {
		t.Errorf("Zero[bool]() = %v, want false", z3)
	}
}

func TestNil(t *testing.T) {
	n := Nil[int]()
	if n != nil {
		t.Errorf("Nil[int]() = %v, want nil", n)
	}
}
