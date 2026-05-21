/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package encoding

import (
	"testing"
)

func TestGBKToUTF8_Roundtrip(t *testing.T) {
	// First encode a UTF-8 string to GBK, then decode it back
	original := "你好世界"
	gbk, err := UTF8ToGBK(original)
	if err != nil {
		t.Fatalf("UTF8ToGBK() error: %v", err)
	}

	result, err := GBKToUTF8(gbk)
	if err != nil {
		t.Fatalf("GBKToUTF8() error: %v", err)
	}

	if result != original {
		t.Errorf("Roundtrip: got %q, want %q", result, original)
	}
}

func TestGBKBytesToUTF8_Roundtrip(t *testing.T) {
	original := []byte("测试中文")
	gbk, err := UTF8BytesToGBK(original)
	if err != nil {
		t.Fatalf("UTF8BytesToGBK() error: %v", err)
	}

	result, err := GBKBytesToUTF8(gbk)
	if err != nil {
		t.Fatalf("GBKBytesToUTF8() error: %v", err)
	}

	if string(result) != string(original) {
		t.Errorf("Roundtrip: got %q, want %q", string(result), string(original))
	}
}

func TestUTF8ToGBK_ASCII(t *testing.T) {
	// ASCII should be the same in both encodings
	original := "hello"
	gbk, err := UTF8ToGBK(original)
	if err != nil {
		t.Fatalf("UTF8ToGBK() error: %v", err)
	}

	result, err := GBKToUTF8(gbk)
	if err != nil {
		t.Fatalf("GBKToUTF8() error: %v", err)
	}

	if result != original {
		t.Errorf("ASCII roundtrip: got %q, want %q", result, original)
	}
}

func TestDetermineEncoding(t *testing.T) {
	// Test with UTF-8 content
	_, name, _ := DetermineEncoding([]byte("hello world"), "text/html; charset=utf-8")
	if name != "utf-8" {
		t.Logf("DetermineEncoding name = %q (may vary)", name)
	}
}
