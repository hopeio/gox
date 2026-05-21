/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package encoding

import (
	"io"
	"testing"
)

type testTextUnmarshaler struct {
	value string
}

func (t *testTextUnmarshaler) UnmarshalText(text []byte) error {
	t.value = string(text)
	return nil
}

// value receiver implements encoding.TextUnmarshaler
type testValueUnmarshaler string

func (t testValueUnmarshaler) UnmarshalText(text []byte) error {
	return nil
}

func TestUnmarshalTextFor_ValueReceiver(t *testing.T) {
	// Test with a type whose value receiver implements TextUnmarshaler
	err := UnmarshalTextFor[testValueUnmarshaler]([]byte("hello"))
	if err != nil {
		t.Errorf("UnmarshalTextFor() error: %v", err)
	}
}

func TestUnmarshalTextFor_PointerReceiver(t *testing.T) {
	// Test with a type whose pointer receiver implements TextUnmarshaler
	err := UnmarshalTextFor[testTextUnmarshaler]([]byte("hello"))
	if err != nil {
		t.Errorf("UnmarshalTextFor() error: %v", err)
	}
}

func TestFormat_Constants(t *testing.T) {
	if Json != "json" {
		t.Errorf("Json = %q, want %q", Json, "json")
	}
	if Yaml != "yaml" {
		t.Errorf("Yaml = %q, want %q", Yaml, "yaml")
	}
	if Toml != "toml" {
		t.Errorf("Toml = %q, want %q", Toml, "toml")
	}
	if Protobuf != "protobuf" {
		t.Errorf("Protobuf = %q, want %q", Protobuf, "protobuf")
	}
	if Xml != "xml" {
		t.Errorf("Xml = %q, want %q", Xml, "xml")
	}
	if Base64 != "base64" {
		t.Errorf("Base64 = %q, want %q", Base64, "base64")
	}
}

// Verify the Codec interface is satisfied by implementing it
type testCodec struct{}

func (t testCodec) Unmarshal(data []byte, v any) error { return nil }
func (t testCodec) Marshal(v any) ([]byte, error)      { return nil, nil }

var _ Codec = testCodec{}

// Verify Decoder and Encoder interfaces
type testDecoder struct{}

func (t testDecoder) Decode(r io.Reader, v any) error { return nil }

type testEncoder struct{}

func (t testEncoder) Encode(w io.Writer, v any) error { return nil }

func TestMarshalUnmarshal(t *testing.T) {
	data := map[string]int{"a": 1}
	b, err := Marshal(data)
	if err != nil {
		t.Errorf("Marshal() error: %v", err)
	}
	if len(b) == 0 {
		t.Error("Marshal() returned empty bytes")
	}

	var result map[string]int
	err = Unmarshal(b, &result)
	if err != nil {
		t.Errorf("Unmarshal() error: %v", err)
	}
	if result["a"] != 1 {
		t.Errorf("Unmarshal result = %v, want a=1", result)
	}
}
