/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package strings

import (
	"testing"
)

func TestSigned(t *testing.T) {
	got, err := Signed[int32]("-42")
	if err != nil || got != -42 {
		t.Fatalf("Signed[int32] = %d, err = %v", got, err)
	}
	if _, err := Signed[int8]("128"); err == nil {
		t.Fatal("want overflow for int8")
	}
	if _, err := Signed[int]("abc"); err == nil {
		t.Fatal("want parse error")
	}
}

func TestSignedP(t *testing.T) {
	p, err := SignedP[int16]("1000")
	if err != nil || p == nil || *p != 1000 {
		t.Fatalf("SignedP = %#v, err = %v", p, err)
	}
	if _, err := SignedP[int16]("bad"); err == nil {
		t.Fatal("want error")
	}
}

func TestSignedSlice(t *testing.T) {
	got, err := SignedSlice[int]("1,2,3", ",")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Fatalf("got %#v", got)
	}
	if _, err := SignedSlice[int]("1,x,3", ","); err == nil {
		t.Fatal("want error on bad element")
	}
}

func TestUnSigned(t *testing.T) {
	got, err := UnSigned[uint16]("65535")
	if err != nil || got != 65535 {
		t.Fatalf("UnSigned = %d, err = %v", got, err)
	}
	if _, err := UnSigned[uint8]("256"); err == nil {
		t.Fatal("want overflow")
	}
}

func TestUnSignedP(t *testing.T) {
	p, err := UnSignedP[uint32]("42")
	if err != nil || p == nil || *p != 42 {
		t.Fatalf("UnSignedP = %#v, err = %v", p, err)
	}
}

func TestUnsignedSlice(t *testing.T) {
	got, err := UnsignedSlice[uint8]("10,20,30", ",")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[1] != 20 {
		t.Fatalf("got %#v", got)
	}
}

func TestFloatAndSlice(t *testing.T) {
	f, err := Float[float64]("3.14")
	if err != nil || f != 3.14 {
		t.Fatalf("Float = %v, err = %v", f, err)
	}
	sl, err := FloatSlice[float32]("1.1,2.2", ",")
	if err != nil || len(sl) != 2 {
		t.Fatalf("FloatSlice = %#v, err = %v", sl, err)
	}
}

func TestBoolSlice(t *testing.T) {
	got, err := BoolSlice("true,false,1", ",")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || !got[0] || got[1] || !got[2] {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSlice(t *testing.T) {
	got, err := StringSlice("a,b,c", ",")
	if err != nil || len(got) != 3 || got[2] != "c" {
		t.Fatalf("got %#v, err = %v", got, err)
	}
}

func TestBytesAndSlice(t *testing.T) {
	imported := "aGVsbG8=" // hello in std base64
	b, err := Bytes(imported)
	if err != nil || string(b) != "hello" {
		t.Fatalf("Bytes = %q, err = %v", b, err)
	}
	if _, err := Bytes("!!!"); err == nil {
		t.Fatal("want decode error")
	}
}

func TestIntFamily(t *testing.T) {
	if v, err := Int64("-9223372036854775808"); err != nil || v != -9223372036854775808 {
		t.Fatalf("Int64 = %d, err = %v", v, err)
	}
	if v, err := Uint64("18446744073709551615"); err != nil || v != 18446744073709551615 {
		t.Fatalf("Uint64 = %d, err = %v", v, err)
	}
	if v, err := IntSlice("1,2", ","); err != nil || len(v) != 2 {
		t.Fatalf("IntSlice = %#v, err = %v", v, err)
	}
}

func TestFloat64Float32Slice(t *testing.T) {
	if v, err := Float32("1.5"); err != nil || v != 1.5 {
		t.Fatalf("Float32 = %v, err = %v", v, err)
	}
	if v, err := Float64Slice("1,2", ","); err != nil || len(v) != 2 {
		t.Fatalf("Float64Slice = %#v, err = %v", v, err)
	}
}

func TestBoolP(t *testing.T) {
	p, err := BoolP("0")
	if err != nil || p == nil || *p {
		t.Fatalf("BoolP = %#v, err = %v", p, err)
	}
}
