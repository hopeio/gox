/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"bytes"
	"io"
	"net/http"
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

// Write performs the operation.
func (rw *ResponseRecorder) Write(buf []byte) (int, error) {
	if len(buf) > 0 {
		if rw.Raw == nil {
			if rw.Body == nil {
				rw.Body = respPool.Get().(*bytes.Buffer)
			}
			rw.Body.Write(buf)
		}
		return rw.originWriter.Write(buf)
	}
	return 0, nil
}

// WriteHeader implements http.ResponseWriter.
func (rw *ResponseRecorder) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.originWriter.WriteHeader(statusCode)
}

// Flush implements http.Flusher. To test whether Flush was
// called, see rw.Flushed.
func (rw *ResponseRecorder) Flush() {
	rw.originWriter.(http.Flusher).Flush()
}

// Read performs the operation.
func (rw *RequestRecorder) Read(b []byte) (int, error) {
	read, err := rw.originBody.Read(b)
	if err != nil {
		return read, err
	}
	return read, err
	/*if rw.Body == nil {
		rw.Body = reqPool.Get().(*bytes.Buffer)
	}
	return rw.Body.Write(b)*/
}

// Close closes and releases resources.
func (rw *RequestRecorder) Close() error {
	return rw.originBody.Close()
}

// Reset removes or resets state.
func (rw *Recorder) Reset() {
	rw.StatusCode = http.StatusOK
	rw.RequestRecorder.Body = nil
	rw.ResponseRecorder.Body = nil
	rw.RequestRecorder.Raw = nil
	if rw.RequestRecorder.Body != nil {
		rw.RequestRecorder.Body.Reset()
		reqPool.Put(rw.RequestRecorder.Body)
	}
	if rw.ResponseRecorder.Body != nil {
		rw.ResponseRecorder.Body.Reset()
		reqPool.Put(rw.ResponseRecorder.Body)
	}
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
