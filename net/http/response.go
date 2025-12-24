/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"context"
	"io"
	"iter"
	"net/http"
	"strconv"

	errorsx "github.com/hopeio/gox/errors"
)

type ResponseWriter interface {
	WriteHeader(code int)
	HeaderX() Header
	Write([]byte) (int, error)
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
}

func (w ResponseWriterWrapper) HeaderX() Header {
	return (HttpHeader)(w.ResponseWriter.Header())
}

func (w ResponseWriterWrapper) RespondStream(ctx context.Context, seq iter.Seq[WriterToCloser]) {
	RespondStream(ctx, w.ResponseWriter, seq)
}

type Responder interface {
	Respond(ctx context.Context, w http.ResponseWriter)
}

// CommonResp 主要用来接收返回，发送请使用ResAnyData
type CommonResp[T any] struct {
	Code errorsx.ErrCode `json:"code"`
	Msg  string          `json:"msg,omitempty"`
	//验证码
	Data T `json:"data,omitempty"`
}

func (res *CommonResp[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *CommonResp[T]) Respond(ctx context.Context, w http.ResponseWriter) {
	data, contentType := DefaultMarshal("", res)
	if wx, ok := w.(ResponseWriter); ok {
		header := wx.HeaderX()
		header.Set(HeaderErrorCode, strconv.Itoa(int(res.Code)))
		header.Set(HeaderContentType, contentType)
	} else {
		header := w.Header()
		header.Set(HeaderErrorCode, strconv.Itoa(int(res.Code)))
		header.Set(HeaderContentType, contentType)
	}
	ow := w
	if ww, ok := w.(Unwrapper); ok {
		ow = ww.Unwrap()
	}
	if recorder, ok := ow.(RecordBody); ok {
		recorder.RecordBody(data, res)
	}
	w.Write(data)
}

type CommonAnyResp = CommonResp[any]

func NewCommonAnyResp(code errorsx.ErrCode, msg string, data any) *CommonAnyResp {
	return &CommonAnyResp{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func RespondErrCodeMsg(w http.ResponseWriter, r *http.Request, code errorsx.ErrCode, msg string) {
	NewCommonAnyResp(code, msg, nil).ServeHTTP(w, r)
}

func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	ErrRespFrom(err).ServeHTTP(w, r)
}

func RespondSuccess(w http.ResponseWriter, r *http.Request, res any) {
	data, contentType := DefaultMarshal(r.Header.Get(HeaderAccept), res)
	if wx, ok := w.(ResponseWriter); ok {
		wx.HeaderX().Set(HeaderContentType, contentType)
	} else {
		w.Header().Set(HeaderContentType, contentType)
	}
	ow := w
	if ww, ok := w.(Unwrapper); ok {
		ow = ww.Unwrap()
	}
	if recorder, ok := ow.(RecordBody); ok {
		recorder.RecordBody(data, res)
	}
	w.Write(data)
}

func Respond(w http.ResponseWriter, r *http.Request, data any) {
	if err, ok := data.(error); ok {
		RespondError(w, r, err)
	}
	RespondSuccess(w, r, data)
}

type Response struct {
	Status  int            `json:"status,omitempty"`
	Headers http.Header    `json:"header,omitempty"`
	Body    WriterToCloser `json:"body,omitempty"`
}

type WriterToCloser interface {
	io.WriterTo
	io.Closer
}

func (res *Response) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *Response) Respond(ctx context.Context, w http.ResponseWriter) {
	if wx, ok := res.Body.(ResponseWriter); ok {
		header := wx.HeaderX()
		for k, v := range res.Headers {
			for _, vv := range v {
				header.Add(k, vv)
			}
		}
	} else {
		CopyHttpHeader(w.Header(), res.Headers)
	}
	w.WriteHeader(res.Status)
	res.Body.WriteTo(w)
	res.Body.Close()
}

type ErrResp errorsx.ErrResp

func NewErrResp(code errorsx.ErrCode, msg string) *ErrResp {
	return &ErrResp{
		Code: code,
		Msg:  msg,
	}
}

func ErrRespFrom(err error) *ErrResp {
	if err == nil {
		return nil
	}
	if errresp, ok := err.(*ErrResp); ok {
		return errresp
	}
	return (*ErrResp)(errorsx.ErrRespFrom(err))
}

func (res *ErrResp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *ErrResp) Respond(ctx context.Context, w http.ResponseWriter) {
	data, contentType := DefaultMarshal("", res)
	if wx, ok := w.(ResponseWriter); ok {
		header := wx.HeaderX()
		header.Set(HeaderErrorCode, strconv.Itoa(int(res.Code)))
		header.Set(HeaderContentType, contentType)
	} else {
		header := w.Header()
		header.Set(HeaderErrorCode, strconv.Itoa(int(res.Code)))
		header.Set(HeaderContentType, contentType)
	}

	if recorder, ok := w.(RecordBody); ok {
		recorder.RecordBody(data, res)
	}
	w.Write(data)
}

func (res *ErrResp) ErrResp() *errorsx.ErrResp {
	return (*errorsx.ErrResp)(res)
}

func (res *ErrResp) Error() string {
	return res.ErrResp().Error()
}

type RespondStreamer interface {
	RespondStream(ctx context.Context, seq iter.Seq[WriterToCloser])
}

type ResponseStream struct {
	Status  int                      `json:"status,omitempty"`
	Headers http.Header              `json:"header,omitempty"`
	Body    iter.Seq[WriterToCloser] `json:"body,omitempty"`
}

func (res *ResponseStream) Respond(ctx context.Context, w http.ResponseWriter) {
	if sw, ok := w.(RespondStreamer); ok {
		sw.RespondStream(ctx, res.Body)
	}
}

func (res *ResponseStream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	RespondStream(r.Context(), w, res.Body)
}

func RespondStream(ctx context.Context, w http.ResponseWriter, dataSource iter.Seq[WriterToCloser]) (int, error) {
	if wx, ok := w.(ResponseWriter); ok {
		header := wx.HeaderX()
		header.Set(HeaderCacheControl, "no-cache")
		header.Set(HeaderTransferEncoding, "chunked")
	} else {
		header := w.Header()
		header.Set(HeaderCacheControl, "no-cache")
		header.Set(HeaderTransferEncoding, "chunked")
	}
	flusher := w.(http.Flusher)
	var n int
	for data := range dataSource {
		select {
		// response writer forced to close, exit.
		case <-ctx.Done():
			return n, nil
		default:
			write, err := data.WriteTo(w)
			if err != nil {
				return n, err
			}
			n += int(write)
			flusher.Flush()
		}
	}
	return n, nil
}

type XXXResponseBody interface {
	XXX_ResponseBody() any
}

type ResponseBody interface {
	ResponseBody() []byte
}

type StatusCode interface {
	StatusCode(v any) int
}
