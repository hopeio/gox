package slices

import (
	"testing"
	"unsafe"
)

type convInt int

func TestConvert_UnsafeSameKind(t *testing.T) {
	in := []convInt{1, 2, 3}
	out := Convert[[]convInt, []int](in)
	if len(out) != 3 || out[0] != 1 || out[2] != 3 {
		t.Fatalf("out=%v", out)
	}
	if unsafe.SliceData(out) != (*int)(unsafe.Pointer(unsafe.SliceData(in))) {
		t.Fatal("same-kind Convert should share backing store")
	}
}

func TestConvert_WidenDoesNotShare(t *testing.T) {
	in := []int32{1, 2}
	out := Convert[[]int32, []int64](in)
	if len(out) != 2 || out[1] != 2 {
		t.Fatalf("out=%v", out)
	}
}

func TestConvert_Empty(t *testing.T) {
	empty := Convert[[]int, []int64]([]int{})
	if len(empty) != 0 {
		t.Fatalf("empty=%v", empty)
	}
}

func TestConvert_UnsupportedPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("want panic for unsupported Convert")
		}
	}()
	_ = Convert[[]Interface, []Interface2]([]Interface{Int8(1)})
}
