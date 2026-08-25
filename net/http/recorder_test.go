package http

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseRecorder_RecordsTextLikeBody(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &ResponseRecorder{originWriter: w}
	rw.Header().Set(HeaderContentType, "application/json; charset=utf-8")
	if _, err := rw.Write([]byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	if rw.Body == nil || rw.Body.String() != `{"a":1}` {
		t.Fatalf("Body=%v want recorded json", rw.Body)
	}
	if w.Body.String() != `{"a":1}` {
		t.Fatalf("origin writer got %q", w.Body.String())
	}
}

func TestResponseRecorder_SkipsBinaryBody(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &ResponseRecorder{originWriter: w}
	rw.Header().Set(HeaderContentType, "image/jpeg")
	if _, err := rw.Write([]byte{0xff, 0xd8, 0xff}); err != nil {
		t.Fatal(err)
	}
	if rw.Body != nil {
		t.Fatalf("binary body should not be recorded, got %d bytes", rw.Body.Len())
	}
	if w.Body.Len() != 3 {
		t.Fatalf("origin writer got %d bytes, want 3", w.Body.Len())
	}
}

func TestResponseRecorder_CapsOversizedBody(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &ResponseRecorder{originWriter: w}
	rw.Header().Set(HeaderContentType, "application/json")
	chunk := bytes.Repeat([]byte("x"), 32<<10)
	for i := 0; i < 4; i++ { // 128KB > MaxRecordBodySize(64KB)
		if _, err := rw.Write(chunk); err != nil {
			t.Fatal(err)
		}
	}
	if rw.Body != nil {
		t.Fatalf("oversized body should abandon recording, kept %d bytes", rw.Body.Len())
	}
	if w.Body.Len() != 128<<10 {
		t.Fatalf("origin writer got %d bytes, want full 128KB", w.Body.Len())
	}
}

func TestRecorder_ResetReturnsState(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader("body"))
	rec := NewRecorder(w, r)
	rec.ResponseRecorder.Header().Set(HeaderContentType, "text/plain")
	rec.ResponseRecorder.Write([]byte("hello"))
	rec.ResponseRecorder.RecordBody([]byte("raw"), "value")
	rec.Reset()
	if rec.ResponseRecorder.Body != nil || rec.ResponseRecorder.Raw != nil || rec.ResponseRecorder.Value != nil {
		t.Fatalf("Reset should clear response state: %+v", rec.ResponseRecorder.Record)
	}
	if rec.ResponseRecorder.ContentType != "" || rec.RequestRecorder.ContentType != "" {
		t.Fatal("Reset should clear content types")
	}
	// 复用后仍能正常录制
	rec.ResponseRecorder.Write([]byte("again"))
	if rec.ResponseRecorder.Body == nil || rec.ResponseRecorder.Body.String() != "again" {
		t.Fatal("recorder should be reusable after Reset")
	}
}

func TestResponseRecorder_WriteImpliesOKStatus(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &ResponseRecorder{originWriter: w}
	if rw.StatusCode != 0 {
		t.Fatalf("fresh StatusCode=%d want 0", rw.StatusCode)
	}
	if _, err := rw.Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	if rw.StatusCode != 200 {
		t.Fatalf("StatusCode=%d want 200 after Write without WriteHeader", rw.StatusCode)
	}
}

func TestResponseRecorder_FlushAndUnwrap(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &ResponseRecorder{originWriter: w}
	rw.Flush() // httptest.ResponseRecorder implements Flusher
	if !w.Flushed {
		t.Fatal("Flush should reach the origin writer")
	}
	if rw.Unwrap() != w {
		t.Fatal("Unwrap should expose the origin writer")
	}
	if _, _, err := rw.Hijack(); err == nil {
		t.Fatal("Hijack on non-hijackable writer should return an error")
	}
}
