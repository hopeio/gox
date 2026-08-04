/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package json

import (
	"reflect"
	"testing"
)

func TestMarshalUnmarshalSmoke(t *testing.T) {
	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	orig := payload{Name: "test", Count: 3}
	data, err := Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}
	var decoded payload
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(orig, decoded) {
		t.Fatalf("round trip = %#v", decoded)
	}
}

func TestMarshalToStringUnmarshalFromString(t *testing.T) {
	s, err := MarshalToString(map[string]int{"a": 1})
	if err != nil || s == "" {
		t.Fatalf("MarshalToString = %q, err = %v", s, err)
	}
	var m map[string]int
	if err := UnmarshalFromString(s, &m); err != nil || m["a"] != 1 {
		t.Fatalf("UnmarshalFromString err=%v m=%#v", err, m)
	}
}

func TestUnquoteCases(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{`"plain"`, "plain", true},
		{`"line\nbreak"`, "line\nbreak", true},
		{`"\u0041"`, "A", true},
		{`"\u8bf7\u6c42"`, "请求", true},
		{`noquotes`, "", false},
		{`"bad\uZZZZ"`, "", false},
	}
	for _, tt := range tests {
		got, ok := Unquote([]byte(tt.in))
		if ok != tt.ok {
			t.Errorf("Unquote(%q) ok = %v, want %v", tt.in, ok, tt.ok)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("Unquote(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDecodeHelpers(t *testing.T) {
	s, rest, err := DecodeString([]byte(`"svc",`))
	if err != nil || s != "svc" {
		t.Fatalf("DecodeString = %q, rest=%q, err=%v", s, rest, err)
	}
	n, rest, err := DecodeInt([]byte(`42,`))
	if err != nil || n != 42 {
		t.Fatalf("DecodeInt = %d, rest=%q, err=%v", n, rest, err)
	}
	f, rest, err := DecodeFloat([]byte(`1.5,`))
	if err != nil || f != 1.5 {
		t.Fatalf("DecodeFloat = %v, rest=%q, err=%v", f, rest, err)
	}
	b, _, err := DecodeBool([]byte(`true}`))
	if err != nil || !b {
		t.Fatalf("DecodeBool = %v, err=%v", b, err)
	}
}

func TestDecodeIntMissingComma(t *testing.T) {
	if _, _, err := DecodeInt([]byte("123")); err == nil {
		t.Fatal("want error when comma missing")
	}
}
