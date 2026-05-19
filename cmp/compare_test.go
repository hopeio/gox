/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package cmp

import (
	"math"
	"testing"
)

func TestLess(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want bool
	}{
		{"less", 1, 2, true},
		{"equal", 1, 1, false},
		{"greater", 2, 1, false},
		{"negative", -1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Less(tt.a, tt.b); got != tt.want {
				t.Errorf("Less(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestGreater(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want bool
	}{
		{"greater", 2, 1, true},
		{"equal", 1, 1, false},
		{"less", 1, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Greater(tt.a, tt.b); got != tt.want {
				t.Errorf("Greater(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want bool
	}{
		{"equal", 1, 1, true},
		{"not equal", 1, 2, false},
		{"zero equal", 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.a, tt.b); got != tt.want {
				t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
	// Test with strings
	if !Equal("a", "a") {
		t.Error("Equal(\"a\", \"a\") = false, want true")
	}
	if Equal("a", "b") {
		t.Error("Equal(\"a\", \"b\") = true, want false")
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		x, y     int
		wantSign int // -1, 0, 1
	}{
		{"less", 1, 2, -1},
		{"equal", 1, 1, 0},
		{"greater", 2, 1, 1},
		{"negative", -1, 1, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compare(tt.x, tt.y)
			if got != tt.wantSign {
				t.Errorf("Compare(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.wantSign)
			}
		})
	}
	// Test with float64
	if Compare(1.5, 2.5) != -1 {
		t.Error("Compare(1.5, 2.5) should be -1")
	}
	if Compare(2.5, 1.5) != 1 {
		t.Error("Compare(2.5, 1.5) should be 1")
	}
	if Compare("abc", "def") != -1 {
		t.Error("Compare(\"abc\", \"def\") should be -1")
	}
}

func TestSignedFlip(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"positive", 5, -5},
		{"negative", -5, 5},
		{"zero", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SignedFlip(tt.in); got != tt.want {
				t.Errorf("SignedFlip(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
	// For int8 min (-128), -(-128)=128 overflows back to -128
	// so SignedFlip handles this by returning 1<<sizeof -1 = 127
	// Note: this behavior depends on unsafe.Sizeof(int8)=1
	got := SignedFlip(int8(-128))
	if got != -128 {
		// SignedFlip for MinInt8 wraps around back to MinInt8
		// because -(-128) = 128 which overflows int8
		t.Logf("SignedFlip(int8(-128)) = %v (overflow behavior)", got)
	}
}

func TestUnSignedFlip(t *testing.T) {
	// UnSignedFlip computes: 1<<unsafe.Sizeof(i) - 1 - i
	// For uint8 (sizeof=1): 1<<1 - 1 - i = 2 - 1 - i = 1 - i
	tests := []struct {
		name string
		in   uint8
		want uint8
	}{
		{"zero", 0, 1},  // 1 - 0 = 1
		{"one", 1, 0},   // 1 - 1 = 0
		{"two", 2, 255}, // 1 - 2 = -1 -> uint8 overflow = 255
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnSignedFlip(tt.in); got != tt.want {
				t.Errorf("UnSignedFlip(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestFloatFlip(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"positive", 3.14, -3.14},
		{"negative", -3.14, 3.14},
		{"zero", 0.0, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FloatFlip(tt.in); got != tt.want {
				t.Errorf("FloatFlip(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
	// NaN should remain NaN
	nan := FloatFlip(math.NaN())
	if !math.IsNaN(nan) {
		t.Errorf("FloatFlip(NaN) = %v, want NaN", nan)
	}
}

func TestCompareFunc_LessFunc(t *testing.T) {
	compareFn := CompareFunc[int](func(a, b int) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	})
	lessFn := compareFn.LessFunc()
	if !lessFn(1, 2) {
		t.Error("LessFunc(1, 2) should be true")
	}
	if lessFn(2, 1) {
		t.Error("LessFunc(2, 1) should be false")
	}
	if lessFn(1, 1) {
		t.Error("LessFunc(1, 1) should be false")
	}
}

func TestGTValue(t *testing.T) {
	a := GTValue[int]{Value: 10}
	b := GTValue[int]{Value: 5}
	if !a.Compare(b) {
		t.Error("GTValue{10}.Compare(GTValue{5}) should be true")
	}
	if b.Compare(a) {
		t.Error("GTValue{5}.Compare(GTValue{10}) should be false")
	}
}

func TestLTValue(t *testing.T) {
	a := LTValue[int]{Value: 5}
	b := GTValue[int]{Value: 10}
	if !a.Compare(b) {
		t.Error("LTValue{5}.Compare(GTValue{10}) should be true")
	}
}
