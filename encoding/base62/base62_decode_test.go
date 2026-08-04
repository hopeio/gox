package base62

import (
	"bytes"
	"testing"
)

func TestDecodeString_RoundTrip(t *testing.T) {
	cases := [][]byte{
		{},
		{0},
		{0, 0, 0},
		{1},
		{61},
		{62},
		{255},
		{0, 1},
		{1, 0},
		{255, 255},
		{0, 0, 255},
		{1, 2, 3, 4, 5},
	}
	for _, in := range cases {
		enc := EncodeToString(in)
		got, err := DecodeString(enc)
		if err != nil {
			t.Fatalf("DecodeString(%q) err=%v", enc, err)
		}
		if !bytes.Equal(got, in) {
			t.Fatalf("roundtrip %v -> %q -> %v", in, enc, got)
		}
	}
}

func TestDecodeString_AlphabetAndErrors(t *testing.T) {
	cases := map[string][]byte{
		"a": {10},
		"z": {35},
		"A": {36},
		"Z": {61},
	}
	for s, want := range cases {
		got, err := DecodeString(s)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%q -> %v want %v", s, got, want)
		}
	}
	if _, err := DecodeString("!!!"); err == nil {
		t.Fatal("want invalid character error")
	}
	got, err := DecodeString("")
	if err != nil || len(got) != 0 {
		t.Fatalf("empty: %v %v", got, err)
	}
}
