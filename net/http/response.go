/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"context"
	"errors"
	"io"
	"iter"
	"net/http"
	"strconv"

	errorsx "github.com/hopeio/gox/errors"
)

// RespData 主要用来接收返回，发送请使用ResAnyData
type RespData[T any] struct {
	Code errorsx.ErrCode `json:"code"`
	Msg  string          `json:"msg,omitempty"`
	//验证码
	Data T `json:"data,omitempty"`
}

func (res *RespData[T]) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	contentType := DefaultMarshaler.ContentType(res)
	w.Header().Set(HeaderErrorCode, strconv.Itoa(int(res.Code)))
	var data []byte
	var err error
	data, err = DefaultMarshaler.Marshal(res)
	if err != nil {
		contentType = ContentTypeText
		data = []byte(err.Error())
	}
	w.Header().Set(HeaderContentType, contentType)
	ow := w
	if ww, ok := w.(Unwrapper); ok {
		ow = ww.Unwrap()
	}
	if recorder, ok := ow.(ResponseRecorder); ok {
		recorder.RecordResponse(contentType, data, res)
	}
	return w.Write(data)
}

type RespAnyData = RespData[any]

func NewRespData(code errorsx.ErrCode, msg string, data any) *RespAnyData {
	return &RespAnyData{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func RespondErrCodeMsg(ctx context.Context, w http.ResponseWriter, code errorsx.ErrCode, msg string) (int, error) {
	return NewRespData(code, msg, nil).Respond(ctx, w)
}

func RespondError(ctx context.Context, w http.ResponseWriter, err error) (int, error) {
	return ErrRespFrom(err).Respond(ctx, w)
}

func RespondSuccess(ctx context.Context, w http.ResponseWriter, res any) (int, error) {
	contentType := DefaultMarshaler.ContentType(res)

	data, err := DefaultMarshaler.Marshal(res)
	if err != nil {
		contentType = ContentTypeText
		data = []byte(err.Error())
	}
	w.Header().Set(HeaderContentType, contentType)
	ow := w
	if ww, ok := w.(Unwrapper); ok {
		ow = ww.Unwrap()
	}
	if recorder, ok := ow.(ResponseRecorder); ok {
		recorder.RecordResponse(contentType, data, res)
	}
	return w.Write(data)
}

func Respond(ctx context.Context, w http.ResponseWriter, data any) (int, error) {
	if err, ok := data.(error); ok {
		return RespondError(ctx, w, err)
	}
	return RespondSuccess(ctx, w, data)
}
func RespStatus(ctx context.Context, w http.ResponseWriter, status int, data any) (int, error) {
	w.WriteHeader(status)
	return Respond(ctx, w, data)
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

func (res *Response) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	w.WriteHeader(res.Status)
	CopyHttpHeader(w.Header(), res.Headers)
	i, err := res.Body.WriteTo(w)
	if err != nil {
		return int(i), err
	}
	err = res.Body.Close()
	if err != nil {
		return int(i), err
	}
	return int(i), err
}

func (res *Response) CommonRespond(ctx context.Context, w CommonResponseWriter) (int, error) {
	w.Status(res.Status)
	HttpHeaderIntoHeader(res.Headers, w.Header())
	i, err := res.Body.WriteTo(w)
	if err != nil {
		return int(i), err
	}
	err = res.Body.Close()
	if err != nil {
		return int(i), err
	}
	return int(i), err
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

func (res *ErrResp) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	return res.CommonRespond(ctx, ResponseWriterWrapper{w})
}

func (res *ErrResp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *ErrResp) CommonRespond(ctx context.Context, w CommonResponseWriter) (int, error) {
	contentType := DefaultMarshaler.ContentType(res)

	w.Header().Set(HeaderErrorCode, strconv.Itoa(int(res.Code)))
	data, err := DefaultMarshaler.Marshal(res)
	if err != nil {
		contentType = ContentTypeText
		data = []byte(err.Error())
	}
	w.Header().Set(HeaderContentType, contentType)
	if recorder, ok := w.(ResponseRecorder); ok {
		recorder.RecordResponse(contentType, data, res)
	}
	return w.Write(data)
}

func (res *ErrResp) ErrResp() *errorsx.ErrResp {
	return (*errorsx.ErrResp)(res)
}

func (res *ErrResp) Error() string {
	return res.ErrResp().Error()
}

type Responder interface {
	Respond(ctx context.Context, w http.ResponseWriter) (int, error)
}

type ResponseStream struct {
	Status  int                      `json:"status,omitempty"`
	Headers http.Header              `json:"header,omitempty"`
	Body    iter.Seq[WriterToCloser] `json:"body,omitempty"`
}

func (res *ResponseStream) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	return RespondStream(ctx, w, res.Body)
}

func (res *ResponseStream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	RespondStream(r.Context(), w, res.Body)
}

func (res *ResponseStream) CommonRespond(ctx context.Context, w CommonResponseWriter) (int, error) {
	if sw, ok := w.(RespondStreamer); ok {
		return sw.RespondStream(ctx, res.Body)
	}
	return 0, errors.New("not support stream")
}

func RespondStream(ctx context.Context, w http.ResponseWriter, dataSource iter.Seq[WriterToCloser]) (int, error) {
	w.Header().Set(HeaderCacheControl, "no-cache")
	w.Header().Set(HeaderTransferEncoding, "chunked")
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
