/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package log

import "testing"

func TestTrimLineBreak(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello\n", "hello"},
		{"hello", "hello"},
		{"a\nb\n", "a\nb"},
	}
	for _, tt := range tests {
		if got := TrimLineBreak(tt.in); got != tt.want {
			t.Errorf("TrimLineBreak(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGetMessage(t *testing.T) {
	if got := getMessage([]interface{}{"a", 1}); got != "a 1" {
		t.Fatalf("getMessage = %q", got)
	}
}

func TestFileWithLineNum(t *testing.T) {
	loc := FileWithLineNum()
	if loc == "" {
		t.Fatal("FileWithLineNum should return non-empty from test context")
	}
}
