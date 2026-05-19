package io

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReadLines(t *testing.T) {
	input := "line1\nline2\nline3"
	reader := strings.NewReader(input)
	var lines []string
	err := ReadLines(reader, func(line string) bool {
		lines = append(lines, line)
		return true
	})
	if err != nil {
		t.Errorf("ReadLines() error: %v", err)
	}
	if len(lines) != 3 {
		t.Errorf("ReadLines() got %d lines, want 3", len(lines))
	}
	if lines[0] != "line1" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "line1")
	}
}

func TestReadLines_EarlyStop(t *testing.T) {
	input := "line1\nline2\nline3"
	reader := strings.NewReader(input)
	var lines []string
	err := ReadLines(reader, func(line string) bool {
		lines = append(lines, line)
		return len(lines) < 2 // stop after 2 lines
	})
	if err != nil {
		t.Errorf("ReadLines() error: %v", err)
	}
	if len(lines) != 2 {
		t.Errorf("ReadLines() with early stop: got %d lines, want 2", len(lines))
	}
}

func TestRawBytes_WriteTo(t *testing.T) {
	data := RawBytes("hello")
	var buf bytes.Buffer
	n, err := data.WriteTo(&buf)
	if err != nil {
		t.Errorf("WriteTo() error: %v", err)
	}
	if n != 5 {
		t.Errorf("WriteTo() n = %d, want 5", n)
	}
	if buf.String() != "hello" {
		t.Errorf("WriteTo() content = %q, want %q", buf.String(), "hello")
	}
}

func TestRawBytes_Write(t *testing.T) {
	var data RawBytes
	n, err := data.Write([]byte("hello"))
	if err != nil {
		t.Errorf("Write() error: %v", err)
	}
	if n != 5 {
		t.Errorf("Write() n = %d, want 5", n)
	}
	if string(data) != "hello" {
		t.Errorf("Write() data = %q, want %q", string(data), "hello")
	}
}

func TestRawBytes_Read(t *testing.T) {
	data := RawBytes("hello")
	buf := make([]byte, 5)
	n, err := data.Read(buf)
	if err != nil {
		t.Errorf("Read() error: %v", err)
	}
	if n != 5 {
		t.Errorf("Read() n = %d, want 5", n)
	}
	if string(buf) != "hello" {
		t.Errorf("Read() content = %q, want %q", string(buf), "hello")
	}
}

func TestRawBytes_Close(t *testing.T) {
	data := RawBytes("hello")
	if err := data.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}
}

func TestRawBytes_Raw(t *testing.T) {
	data := RawBytes("hello")
	if string(data.Raw()) != "hello" {
		t.Errorf("Raw() = %q, want %q", string(data.Raw()), "hello")
	}
}

func TestLimitedWriter(t *testing.T) {
	lw := NewLimitedWriter(10)
	n, err := lw.Write([]byte("hello"))
	if err != nil {
		t.Errorf("Write() error: %v", err)
	}
	if n != 5 {
		t.Errorf("Write() n = %d, want 5", n)
	}
	// Write more than remaining capacity (5 bytes left)
	// LimitedWriter truncates to remaining and writes without EOF
	// Note: the EOF condition in the implementation is unreachable due to truncation
	n, err = lw.Write([]byte("1234567890"))
	if n != 5 {
		t.Errorf("Write() over capacity: n = %d, want 5", n)
	}
	// After filling capacity, next write returns EOF
	n, err = lw.Write([]byte("x"))
	if err != io.EOF {
		t.Errorf("Write() at full capacity: err = %v, want EOF", err)
	}
	if n != 0 {
		t.Errorf("Write() at full capacity: n = %d, want 0", n)
	}
}

func TestWrapReader(t *testing.T) {
	r := strings.NewReader("hello")
	called := false
	wrapper := WrapReader(r, func() error {
		called = true
		return nil
	})
	buf := make([]byte, 5)
	wrapper.Read(buf)
	if string(buf) != "hello" {
		t.Errorf("Read() = %q, want %q", string(buf), "hello")
	}
	wrapper.Close()
	if !called {
		t.Error("Close() did not call close function")
	}
}

func TestWrapReader_NilClose(t *testing.T) {
	r := strings.NewReader("hello")
	wrapper := WrapReader(r, nil)
	if err := wrapper.Close(); err != nil {
		t.Errorf("Close() with nil close func returned error: %v", err)
	}
}
