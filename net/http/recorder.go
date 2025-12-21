/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/http/httpguts"
)

type Unwrapper interface {
	Unwrap() http.ResponseWriter
}

type RequestRecorder interface {
	RecordRequest(contentType string, raw []byte, v any)
}

type ResponseRecorder interface {
	RecordResponse(contentType string, raw []byte, v any)
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

// Recorder is an implementation of http.ResponseWriter that
// records its mutations for later inspection in tests.
type Recorder struct {
	originWriter http.ResponseWriter
	originBody   io.ReadCloser
	Code         int
	Request      Record
	Reponse      Record
	result       *http.Response // cache of Result's return value
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
		originWriter: w,
		originBody:   r.Body,
		Code:         http.StatusOK,
	}
}

func (rw *Recorder) Header() http.Header {
	return rw.originWriter.Header()
}

func (rw *Recorder) Write(buf []byte) (int, error) {
	if len(buf) > 0 {
		if rw.Reponse.Raw == nil {
			if rw.Reponse.Body == nil {
				rw.Reponse.Body = respPool.Get().(*bytes.Buffer)
			}
			rw.Reponse.Body.Write(buf)
		}
		return rw.originWriter.Write(buf)
	}
	return 0, nil
}

// WriteHeader implements http.ResponseWriter.
func (rw *Recorder) WriteHeader(code int) {
	rw.Code = code
	rw.originWriter.WriteHeader(code)
}

// Flush implements http.Flusher. To test whether Flush was
// called, see rw.Flushed.
func (rw *Recorder) Flush() {
	rw.originWriter.(http.Flusher).Flush()
}

func (rw *Recorder) Result() *http.Response {
	if rw.result != nil {
		return rw.result
	}
	header := rw.originWriter.Header()
	res := &http.Response{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		StatusCode: rw.Code,
		Header:     header,
	}
	rw.result = res
	if res.StatusCode == 0 {
		res.StatusCode = 200
	}
	res.Status = fmt.Sprintf("%03d %s", res.StatusCode, http.StatusText(res.StatusCode))
	if rw.Reponse.Body != nil {
		res.Body = io.NopCloser(bytes.NewReader(rw.Reponse.Body.Bytes()))
	} else {
		res.Body = http.NoBody
	}
	res.ContentLength = parseContentLength(res.Header.Get("Content-Length"))

	if trailers, ok := header["Trailer"]; ok {
		res.Trailer = make(http.Header, len(trailers))
		for _, k := range trailers {
			k = http.CanonicalHeaderKey(k)
			if !httpguts.ValidTrailerHeader(k) {
				// Ignore since forbidden by RFC 7230, section 4.1.2.
				continue
			}
			vv, ok := header[k]
			if !ok {
				continue
			}
			vv2 := make([]string, len(vv))
			copy(vv2, vv)
			res.Trailer[k] = vv2
		}
	}
	for k, vv := range header {
		if !strings.HasPrefix(k, http.TrailerPrefix) {
			continue
		}
		if res.Trailer == nil {
			res.Trailer = make(http.Header)
		}
		for _, v := range vv {
			res.Trailer.Add(strings.TrimPrefix(k, http.TrailerPrefix), v)
		}
	}
	return res
}

func parseContentLength(cl string) int64 {
	cl = textproto.TrimString(cl)
	if cl == "" {
		return -1
	}
	n, err := strconv.ParseUint(cl, 10, 63)
	if err != nil {
		return -1
	}
	return int64(n)
}

func (rw *Recorder) Read(b []byte) (int, error) {
	if rw.Request.Body == nil {
		rw.Request.Body = reqPool.Get().(*bytes.Buffer)
	}
	read, err := rw.originBody.Read(b)
	if err != nil {
		return read, err
	}
	return rw.Request.Body.Write(b)
}

func (rw *Recorder) Close() error {
	return rw.originBody.Close()
}

func (rw *Recorder) Reset() {
	rw.Code = http.StatusOK
	rw.Request.Body = nil
	rw.Reponse.Body = nil
	rw.Request.Raw = nil
	if rw.Request.Body != nil {
		rw.Request.Body.Reset()
		reqPool.Put(rw.Request.Body)
	}
	if rw.Reponse.Body != nil {
		rw.Reponse.Body.Reset()
		reqPool.Put(rw.Reponse.Body)
	}
}

func (rw *Recorder) RecordRequest(contentType string, raw []byte, v any) {
	rw.Request.Raw = raw
	rw.Request.Value = v
	rw.Request.ContentType = contentType
}

func (rw *Recorder) RecordResponse(contentType string, raw []byte, v any) {
	rw.Reponse.Raw = raw
	rw.Reponse.Value = v
	rw.Reponse.ContentType = contentType
}
