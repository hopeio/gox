/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package unsafe

import (
	"testing"
	"unsafe"
)

func TestCast(t *testing.T) {
	type inner struct {
		x int
		y int
	}
	v := inner{x: 42, y: 99}
	p := Cast[int](&v)
	if *p != 42 {
		t.Errorf("Cast() = %v, want 42", *p)
	}
}

func TestCastSlice(t *testing.T) {
	// CastSlice preserves len(s), so []int32 of len 2 -> []byte of len 2
	src := []int32{0x01, 0x02}
	dst := CastSlice[byte](src)
	if len(dst) != len(src) {
		t.Errorf("CastSlice len = %d, want %d", len(dst), len(src))
	}
}

func TestBinary(t *testing.T) {
	val := uint32(0x12345678)
	b := Binary(unsafe.Pointer(&val), 4)
	if len(b) != 4 {
		t.Errorf("Binary len = %d, want 4", len(b))
	}
}

func TestClear(t *testing.T) {
	val := 42
	Clear(&val)
	if val != 0 {
		t.Errorf("Clear() failed, val = %v, want 0", val)
	}
}

func TestNoEscape(t *testing.T) {
	x := 42
	p := unsafe.Pointer(&x)
	result := NoEscape(p)
	if result == nil {
		t.Error("NoEscape() returned nil")
	}
}

func TestUnsafePtr_Roundtrip(t *testing.T) {
	original := "hello world"
	p := UnsafePtr(original)
	result := FromUnsafePtr(p, int64(len(original)))
	if result != original {
		t.Errorf("Roundtrip: got %q, want %q", result, original)
	}
}

func TestIndexChar(t *testing.T) {
	s := "hello"
	p := IndexChar(s, 0)
	if p == nil {
		t.Error("IndexChar() returned nil")
	}
}

func TestIndexByte(t *testing.T) {
	b := []byte("hello")
	p := IndexByte(b, 0)
	if p == nil {
		t.Error("IndexByte() returned nil")
	}
}
