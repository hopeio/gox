/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

func newTestClient() *Client {
	return New().DisableLog()
}

// newNoCompressClient disables transport auto-decompress to test manual decompress logic.
func newNoCompressClient() *Client {
	tr := apiTransport()
	tr.DisableCompression = true
	return New().DisableLog().HttpClient(&http.Client{Transport: tr})
}

// --- ContentType ---

func TestContentTypeString(t *testing.T) {
	tests := []struct {
		ct   ContentType
		want string
	}{
		{ContentTypeUnset, ""},
		{ContentTypeJson, "application/json;charset=UTF-8"},
		{ContentTypeForm, "application/x-www-form-urlencoded;charset=UTF-8"},
		{ContentTypeFormData, "multipart/form-data;charset=UTF-8"},
		{ContentTypeXml, "application/xml;charset=UTF-8"},
		{ContentTypeText, "text/plain;charset=UTF-8"},
		{ContentTypeBinary, "application/octet-stream"},
		{ContentTypeImage, "application/octet-stream"},
		{ContentTypeAudio, "application/octet-stream"},
		{ContentTypeVideo, "application/octet-stream"},
	}
	for _, tt := range tests {
		got := tt.ct.String()
		if got != tt.want {
			t.Errorf("ContentType(%d).String() = %q, want %q", tt.ct, got, tt.want)
		}
	}
}

func TestContentTypeDecode(t *testing.T) {
	tests := []struct {
		input string
		want  ContentType
	}{
		{"application/json", ContentTypeJson},
		{"application/json; charset=utf-8", ContentTypeJson},
		{"application/x-www-form-urlencoded", ContentTypeForm},
		{"text/html", ContentTypeText},
		{"image/png", ContentTypeImage},
		{"video/mp4", ContentTypeVideo},
		{"audio/mpeg", ContentTypeAudio},
		{"application/octet-stream", ContentTypeApplication},
		{"unknown", ContentTypeJson},
	}
	for _, tt := range tests {
		var ct ContentType
		ct.Decode(tt.input)
		if ct != tt.want {
			t.Errorf("Decode(%q) = %d, want %d", tt.input, ct, tt.want)
		}
	}
}

// --- Basic Request ---

func TestGetJson(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"hello": "world"})
	}))
	defer srv.Close()

	var resp map[string]string
	err := newTestClient().Get(srv.URL, nil, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp["hello"] != "world" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestPostJson(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body) // echo
	}))
	defer srv.Close()

	reqBody := map[string]string{"key": "value"}
	var resp map[string]string
	err := newTestClient().Post(srv.URL, reqBody, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp["key"] != "value" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestGetRaw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("raw data"))
	}))
	defer srv.Close()

	data, err := newTestClient().GetRaw(srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "raw data" {
		t.Fatalf("unexpected raw: %s", data)
	}
}

func TestDoStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("stream content"))
	}))
	defer srv.Close()

	reader, err := newTestClient().GetStream(srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, _ := io.ReadAll(reader)
	if string(data) != "stream content" {
		t.Fatalf("unexpected stream: %s", data)
	}
}

// --- BaseUrl ---

func TestBaseUrl(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/test" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`"ok"`))
	}))
	defer srv.Close()

	var resp string
	err := newTestClient().BaseUrl(srv.URL).Get("/api/test", nil, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp != "ok" {
		t.Fatalf("unexpected: %s", resp)
	}
}

// --- 404 ---

func TestNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	var resp any
	err := newTestClient().Get(srv.URL, nil, &resp)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	var resp any
	err := newTestClient().Get(srv.URL, nil, &resp)
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Retry ---

func TestRetryOnNetworkError(t *testing.T) {
	var count atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count.Add(1) < 3 {
			// hijack and close to simulate network error
			hj := w.(http.Hijacker)
			conn, _, _ := hj.Hijack()
			conn.Close()
			return
		}
		w.Write([]byte(`"success"`))
	}))
	defer srv.Close()

	var resp string
	err := newTestClient().RetryTimesWithInterval(3, 10*time.Millisecond).Get(srv.URL, nil, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp != "success" {
		t.Fatalf("unexpected: %s", resp)
	}
	if count.Load() != 3 {
		t.Fatalf("expected 3 attempts, got %d", count.Load())
	}
}

func TestRetryExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj := w.(http.Hijacker)
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	var resp string
	err := newTestClient().RetryTimesWithInterval(2, 10*time.Millisecond).Get(srv.URL, nil, &resp)
	if err == nil {
		t.Fatal("expected error after retry exhausted")
	}
}

func TestResponseHandlerRetry(t *testing.T) {
	var count atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`"data"`))
	}))
	defer srv.Close()

	c := newTestClient().RetryTimesWithInterval(3, 10*time.Millisecond)
	c.responseHandler = func(response *http.Response) (bool, io.ReadCloser, error) {
		if count.Add(1) < 3 {
			return true, nil, nil // request retry
		}
		return false, nil, nil
	}

	var resp string
	err := c.Get(srv.URL, nil, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp != "data" {
		t.Fatalf("unexpected: %s", resp)
	}
	if count.Load() != 3 {
		t.Fatalf("expected handler called 3 times, got %d", count.Load())
	}
}

// --- Context cancellation ---

func TestContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj := w.(http.Hijacker)
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	var resp string
	err := NewRequest(http.MethodGet, srv.URL).
		Client(newTestClient().RetryTimesWithInterval(5, time.Second)).
		Context(ctx).
		Do(nil, &resp)
	if err == nil {
		t.Fatal("expected context error")
	}
}

// --- Decompression ---

func TestGzipDecompression(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		gw.Write([]byte(`{"msg":"gzipped"}`))
		gw.Close()
	}))
	defer srv.Close()

	var resp map[string]string
	err := newNoCompressClient().Get(srv.URL, nil, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp["msg"] != "gzipped" {
		t.Fatalf("unexpected: %v", resp)
	}
}

func TestBrotliDecompression(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "br")
		bw := brotli.NewWriter(w)
		bw.Write([]byte(`{"msg":"brotli"}`))
		bw.Close()
	}))
	defer srv.Close()

	var resp map[string]string
	err := newNoCompressClient().Get(srv.URL, nil, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp["msg"] != "brotli" {
		t.Fatalf("unexpected: %v", resp)
	}
}

func TestZstdDecompression(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "zstd")
		zw, _ := zstd.NewWriter(w)
		zw.Write([]byte(`{"msg":"zstd"}`))
		zw.Close()
	}))
	defer srv.Close()

	var resp map[string]string
	err := newNoCompressClient().Get(srv.URL, nil, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp["msg"] != "zstd" {
		t.Fatalf("unexpected: %v", resp)
	}
}

// --- respBodyHandler ---

func TestRespBodyHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":"encrypted"}`))
	}))
	defer srv.Close()

	c := newTestClient()
	c.respBodyHandler = func(data []byte) ([]byte, error) {
		// simulate "decryption": replace encrypted with decrypted
		return []byte(`{"data":"decrypted"}`), nil
	}

	var resp map[string]string
	err := c.Get(srv.URL, nil, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp["data"] != "decrypted" {
		t.Fatalf("unexpected: %v", resp)
	}
}

// --- ResponseBodyCheck ---

type checkedResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (r *checkedResp) CheckError() error {
	if r.Code != 0 {
		return io.ErrUnexpectedEOF // arbitrary error for test
	}
	return nil
}

func TestResponseBodyCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"code":1,"msg":"fail"}`))
	}))
	defer srv.Close()

	var resp checkedResp
	err := newTestClient().Get(srv.URL, nil, &resp)
	if err == nil {
		t.Fatal("expected CheckError to return error")
	}
}

// --- Clone ---

func TestClone(t *testing.T) {
	c := New().DisableLog().AddHeader("X-Test", "original").RetryTimes(3)
	clone := c.Clone()

	// modify clone header should not affect original
	clone.header.Set("X-Test", "modified")
	if c.header.Get("X-Test") != "original" {
		t.Fatal("clone header modification affected original")
	}

	// modify clone retryTimes should not affect original
	clone.retryTimes = 10
	if c.retryTimes != 3 {
		t.Fatal("clone field modification affected original")
	}
}

// --- Header merging ---

func TestHeaderMerge(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Custom")
		w.Write([]byte(`"ok"`))
	}))
	defer srv.Close()

	c := newTestClient().AddHeader("X-Custom", "from-client")
	var resp string
	err := NewRequest(http.MethodGet, srv.URL).
		Client(c).
		AddHeader("X-Custom", "from-request").
		Do(nil, &resp)
	if err != nil {
		t.Fatal(err)
	}
	// CopyHttpHeader appends client headers, Header.Get returns first value (request-level)
	if gotHeader != "from-request" {
		t.Fatalf("expected request header first, got: %s", gotHeader)
	}
}
