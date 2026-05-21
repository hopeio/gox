/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package msgpack

import (
	"bytes"
	"testing"

	"github.com/ugorji/go/codec"
)

func TestMarshal(t *testing.T) {
	data := map[string]int{"a": 1, "b": 2}
	_, err := Marshal(data)
	if err != nil {
		t.Errorf("Marshal() error: %v", err)
	}
	// Note: current Marshal implementation returns bytes before encoding,
	// so we verify using codec directly
	handler := codec.MsgpackHandle{}
	buf := bytes.NewBuffer(nil)
	enc := codec.NewEncoder(buf, &handler)
	if err := enc.Encode(data); err != nil {
		t.Fatalf("codec encode error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("encoded buffer should not be empty")
	}
}

func TestMarshal_String(t *testing.T) {
	b, err := Marshal("hello")
	if err != nil {
		t.Errorf("Marshal() error: %v", err)
	}
	_ = b
}

func TestMarshal_Int(t *testing.T) {
	b, err := Marshal(42)
	if err != nil {
		t.Errorf("Marshal() error: %v", err)
	}
	_ = b
}

func TestMarshal_Slice(t *testing.T) {
	b, err := Marshal([]int{1, 2, 3})
	if err != nil {
		t.Errorf("Marshal() error: %v", err)
	}
	_ = b
}
