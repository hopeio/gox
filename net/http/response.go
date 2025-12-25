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

type Responder interface {
	Respond(ctx context.Context, w http.ResponseWriter) (int, error)
}

// CommonResp 主要用来接收返回，发送请使用 CommonAnyResp
type CommonResp[T any] struct {
	Code errorsx.ErrCode `json:"code"`
	Msg  string          `json:"msg,omitempty"`
	//验证码
	Data T `json:"data,omitempty"`
}

type CommonAnyResp CommonResp[any]

func (res *CommonAnyResp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *CommonAnyResp) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {

	data, contentType := DefaultMarshal(ctx, res)
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
	if uw, ok := w.(Unwrapper); ok {
		ow = uw.Unwrap()
	}
	if recorder, ok := ow.(RecordBody); ok {
		recorder.RecordBody(data, res)
	}
	return w.Write(data)
}

func NewCommonAnyResp(code errorsx.ErrCode, msg string, data any) *CommonAnyResp {
	return &CommonAnyResp{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func ServeErrCodeMsg(w http.ResponseWriter, r *http.Request, code errorsx.ErrCode, msg string) {
	NewErrResp(code, msg).ServeHTTP(w, r)
}

func RespondErrCodeMsg(ctx context.Context, w http.ResponseWriter, code errorsx.ErrCode, msg string) {
	NewErrResp(code, msg).Respond(ctx, w)
}

func ServeError(w http.ResponseWriter, r *http.Request, err error) {
	ErrRespFrom(err).ServeHTTP(w, r)
}

func RespondError(ctx context.Context, w http.ResponseWriter, err error) (int, error) {
	return ErrRespFrom(err).Respond(ctx, w)
}

func ServeSuccess(w http.ResponseWriter, r *http.Request, res any) {
	RespondSuccess(r.Context(), w, res)
}

func RespondSuccess(ctx context.Context, w http.ResponseWriter, res any) (int, error) {
	data, contentType := DefaultMarshal(ctx, res)
	if wx, ok := w.(ResponseWriter); ok {
		wx.HeaderX().Set(HeaderContentType, contentType)
	} else {
		w.Header().Set(HeaderContentType, contentType)
	}
	ow := w
	if uw, ok := w.(Unwrapper); ok {
		ow = uw.Unwrap()
	}
	if recorder, ok := ow.(RecordBody); ok {
		recorder.RecordBody(data, res)
	}
	return w.Write(data)
}

func Serve(w http.ResponseWriter, r *http.Request, data any) {
	if err, ok := data.(error); ok {
		ServeError(w, r, err)
	}
	ServeSuccess(w, r, data)
}

func Respond(ctx context.Context, w http.ResponseWriter, data any) (int, error) {
	if err, ok := data.(error); ok {
		return RespondError(ctx, w, err)
	}
	return RespondSuccess(ctx, w, data)
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

func (res *Response) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	if wx, ok := w.(ResponseWriter); ok {
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
	n, err := res.Body.WriteTo(w)
	res.Body.Close()
	return int(n), err
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

func (res *ErrResp) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {

	data, contentType := DefaultMarshal(ctx, res)
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
	if uw, ok := w.(Unwrapper); ok {
		ow = uw.Unwrap()
	}
	if recorder, ok := ow.(RecordBody); ok {
		recorder.RecordBody(data, res)
	}
	return w.Write(data)
}

func (res *ErrResp) ErrResp() *errorsx.ErrResp {
	return (*errorsx.ErrResp)(res)
}

func (res *ErrResp) Error() string {
	return res.ErrResp().Error()
}

type RespondStreamer interface {
	RespondStream(ctx context.Context, seq iter.Seq[WriterToCloser]) (int, error)
}

type ResponseStream struct {
	Status  int                      `json:"status,omitempty"`
	Headers http.Header              `json:"header,omitempty"`
	Body    iter.Seq[WriterToCloser] `json:"body,omitempty"`
}

func (res *ResponseStream) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	if sw, ok := w.(RespondStreamer); ok {
		return sw.RespondStream(ctx, res.Body)
	}
	return 0, nil
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
	var n int
	flusher := w.(http.Flusher)
	for data := range dataSource {
		select {
		case <-ctx.Done():
			return n, ctx.Err()
		default:
			writen, err := data.WriteTo(w)
			n += int(writen)
			if err != nil {
				return n, err
			}
			flusher.Flush()
		}
	}
	return n, nil
}

type XXXResponseBody interface {
	XXX_ResponseBody() any
}

type ResponseBody interface {
	ResponseBody() ([]byte, string)
}

type StatusCode interface {
	StatusCode(v any) int
}
