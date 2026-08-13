/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Unwrapper interface {
	Unwrap() http.ResponseWriter
}

type RecordBodyer interface {
	RecordBody(raw []byte, v any)
}

var reqPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

var respPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

type Recorder struct {
	RequestRecorder
	ResponseRecorder
}

type RequestRecorder struct {
	Record
	originBody io.ReadCloser
}

type ResponseRecorder struct {
	Record
	originWriter http.ResponseWriter
	StatusCode   int
	skipRecord   bool
}

// MaxRecordBodySize caps how many response bytes are buffered for access logging.
// Once exceeded, buffering is abandoned for that response (the client still
// receives everything); large downloads must not be duplicated in memory.
var MaxRecordBodySize = 64 << 10

// recordableContentType reports whether a response body is text-like and worth
// recording. Empty content type is treated as recordable (small JSON default
// paths often set it late); binary types are skipped.
func recordableContentType(ct string) bool {
	if ct == "" {
		return true
	}
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	if strings.HasPrefix(ct, "text/") {
		return true
	}
	switch ct {
	case "application/json", "application/xml", "application/x-www-form-urlencoded",
		"application/x-protobuf", "application/protobuf", "application/grpc":
		return true
	}
	return strings.HasSuffix(ct, "+json") || strings.HasSuffix(ct, "+xml")
}

type Record struct {
	ContentType string
	Body        *bytes.Buffer
	Raw         []byte
	Value       any
}

// NewRecorder returns an initialized Recorder.
func NewRecorder(w http.ResponseWriter, r *http.Request) *Recorder {
	return &Recorder{
		RequestRecorder: RequestRecorder{
			originBody: r.Body,
		},
		ResponseRecorder: ResponseRecorder{
			originWriter: w,
		},
	}
}

// Header returns the result.
func (rw *ResponseRecorder) Header() http.Header {
	return rw.originWriter.Header()
}

// Write forwards to the underlying writer, buffering a copy for access logging
// only when the content type is text-like and the total stays under
// MaxRecordBodySize.
func (rw *ResponseRecorder) Write(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	if rw.Raw == nil && !rw.skipRecord {
		if rw.Body == nil {
			if recordableContentType(rw.originWriter.Header().Get(HeaderContentType)) {
				rw.Body = respPool.Get().(*bytes.Buffer)
			} else {
				rw.skipRecord = true
			}
		}
		if rw.Body != nil {
			if rw.Body.Len()+len(buf) > MaxRecordBodySize {
				// 超限即放弃录制并归还缓冲，避免大响应在内存里复制一份
				rw.Body.Reset()
				respPool.Put(rw.Body)
				rw.Body = nil
				rw.skipRecord = true
			} else {
				rw.Body.Write(buf)
			}
		}
	}
	return rw.originWriter.Write(buf)
}

// WriteHeader implements http.ResponseWriter.
func (rw *ResponseRecorder) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.originWriter.WriteHeader(statusCode)
}

// Flush implements http.Flusher; it is a no-op when the underlying writer
// does not support flushing instead of panicking.
func (rw *ResponseRecorder) Flush() {
	if f, ok := rw.originWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker so protocol upgrades (e.g. websocket)
// keep working through the recorder.
func (rw *ResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.originWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Unwrap exposes the underlying writer for http.ResponseController.
func (rw *ResponseRecorder) Unwrap() http.ResponseWriter {
	return rw.originWriter
}

// Read performs the operation.
func (rw *RequestRecorder) Read(b []byte) (int, error) {
	return rw.originBody.Read(b)
}

// Close closes and releases resources.
func (rw *RequestRecorder) Close() error {
	return rw.originBody.Close()
}

// Reset returns pooled buffers and clears recorded state.
// 注意先归还再置 nil：曾因先置 nil 导致缓冲永远不回池，每个请求都重新分配。
func (rw *Recorder) Reset() {
	rw.StatusCode = http.StatusOK
	rw.skipRecord = false
	if rw.RequestRecorder.Body != nil {
		rw.RequestRecorder.Body.Reset()
		reqPool.Put(rw.RequestRecorder.Body)
		rw.RequestRecorder.Body = nil
	}
	if rw.ResponseRecorder.Body != nil {
		rw.ResponseRecorder.Body.Reset()
		respPool.Put(rw.ResponseRecorder.Body)
		rw.ResponseRecorder.Body = nil
	}
	rw.RequestRecorder.Raw = nil
	rw.ResponseRecorder.Raw = nil
	rw.RequestRecorder.Value = nil
	rw.ResponseRecorder.Value = nil
	rw.RequestRecorder.ContentType = ""
	rw.ResponseRecorder.ContentType = ""
}

// RecordBody performs the operation.
func (rw *Record) RecordBody(raw []byte, v any) {
	if len(raw) > 0 {
		rw.Raw = raw
	}
	if v != nil {
		rw.Value = v
	}
}
