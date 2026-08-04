/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package strings

import (
	"encoding"
	"testing"
	"time"
)

type textVal string

func (t *textVal) UnmarshalText(b []byte) error {
	*t = textVal("T:" + string(b))
	return nil
}

var _ encoding.TextUnmarshaler = (*textVal)(nil)

func TestParseFor_DurationAndHex(t *testing.T) {
	d, err := ParseFor[time.Duration]("500ms")
	if err != nil {
		t.Fatal(err)
	}
	if d != 500*time.Millisecond {
		t.Fatalf("got %v", d)
	}
	n, err := ParseFor[int]("0x10")
	if err != nil {
		t.Fatal(err)
	}
	if n != 16 {
		t.Fatalf("hex got %d", n)
	}
}

func TestParseFor_OverflowAndUnsupported(t *testing.T) {
	if _, err := ParseFor[int8]("128"); err == nil {
		t.Fatal("want overflow error")
	}
	if _, err := ParseFor[struct{}]("x"); err == nil {
		t.Fatal("want unsupported type")
	}
}

func TestParseFor_PointerAndTextUnmarshaler(t *testing.T) {
	p, err := ParseFor[*int]("42")
	if err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 42 {
		t.Fatalf("got %#v", p)
	}
	tv, err := ParseFor[textVal]("hi")
	if err != nil {
		t.Fatal(err)
	}
	if tv != "T:hi" {
		t.Fatalf("got %q", tv)
	}
}

func TestParsePtrFor(t *testing.T) {
	p, err := ParsePtrFor[int]("7")
	if err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 7 {
		t.Fatalf("got %#v", p)
	}
}
