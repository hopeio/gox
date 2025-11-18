/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"

	"github.com/hopeio/gox/errors"
)

// RespData 主要用来接收返回，发送请使用ResAnyData
type RespData[T any] struct {
	Code errors.ErrCode `json:"code"`
	Msg  string         `json:"msg,omitempty"`
	//验证码
	Data T `json:"data,omitempty"`
}

func (res *RespData[T]) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	w.Header().Set(HeaderContentType, "application/json; charset=utf-8")
	jsonBytes, _ := json.Marshal(res)
	return w.Write(jsonBytes)
}

type RespAnyData = RespData[any]

func NewRespData(code errors.ErrCode, msg string, data any) *RespAnyData {
	return &RespAnyData{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func NewSuccessRespData(data any) *RespAnyData {
	return &RespAnyData{
		Data: data,
	}
}

func NewErrResp(code errors.ErrCode, msg string) *ErrResp {
	return &ErrResp{
		Code: code,
		Msg:  msg,
	}
}

func RespErrCodeMsg(ctx context.Context, w http.ResponseWriter, code errors.ErrCode, msg string) (int, error) {
	return NewRespData(code, msg, nil).Respond(ctx, w)
}

func RespErrResp(ctx context.Context, w http.ResponseWriter, rep *errors.ErrResp) (int, error) {
	return (*ErrResp)(rep).Respond(ctx, w)
}

func RespError(ctx context.Context, w http.ResponseWriter, err error) (int, error) {
	return ErrRespFrom(err).Respond(ctx, w)
}

func RespSuccess(ctx context.Context, w http.ResponseWriter, msg string, data any) (int, error) {
	return NewRespData(errors.Success, msg, data).Respond(ctx, w)
}

func RespSuccessData(ctx context.Context, w http.ResponseWriter, data any) (int, error) {
	return NewRespData(errors.Success, errors.Success.String(), data).Respond(ctx, w)
}

func Resp(ctx context.Context, w http.ResponseWriter, code errors.ErrCode, msg string, data any) (int, error) {
	return NewRespData(code, msg, data).Respond(ctx, w)
}
func RespStatus(ctx context.Context, w http.ResponseWriter, code errors.ErrCode, msg string, data any, status int) (int, error) {
	w.WriteHeader(status)
	return NewRespData(code, msg, data).Respond(ctx, w)
}

var RespSysErr = json.RawMessage(`{"code":-1,"msg":"system error"}`)
var RespOk = json.RawMessage(`{"code":0}`)

type ReceiveData = RespData[json.RawMessage]

func NewReceiveData(code errors.ErrCode, msg string, data any) *ReceiveData {
	jsonBytes, _ := json.Marshal(data)
	return &ReceiveData{
		Code: code,
		Msg:  msg,
		Data: jsonBytes,
	}
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

type ErrResp errors.ErrResp

func ErrRespFrom(err error) *ErrResp {
	return (*ErrResp)(errors.ErrRespFrom(err))
}

func (res *ErrResp) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	return res.CommonRespond(ctx, ResponseWriterWrapper{w})
}

func (res *ErrResp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *ErrResp) CommonRespond(ctx context.Context, w CommonResponseWriter) (int, error) {
	w.Header().Set(HeaderContentType, ContentTypeJsonUtf8)
	jsonBytes, _ := json.Marshal(res)
	return w.Write(jsonBytes)
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
	return res.CommonRespond(ctx, ResponseWriterWrapper{w})
}

func (res *ResponseStream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.CommonRespond(r.Context(), ResponseWriterWrapper{w})
}

func (res *ResponseStream) CommonRespond(ctx context.Context, w CommonResponseWriter) (int, error) {
	header := w.Header()
	HttpHeaderIntoHeader(res.Headers, header)
	header.Set(HeaderTransferEncoding, "chunked")
	var n int
	for data := range res.Body {
		select {
		// response writer forced to close, exit.
		case <-ctx.Done():
			return n, nil
		default:
			write, err := data.WriteTo(w)
			if err != nil {
				return 0, err
			}
			n += int(write)
			w.(http.Flusher).Flush()
		}
	}
	return n, nil
}

func RespStream(ctx context.Context, w http.ResponseWriter, dataSource iter.Seq[WriterToCloser]) {
	w.Header().Set(HeaderXAccelBuffering, "no") //nginx的锅必须加
	w.Header().Set(HeaderTransferEncoding, "chunked")
	for data := range dataSource {
		select {
		// response writer forced to close, exit.
		case <-ctx.Done():
			return
		default:
			data.WriteTo(w)
			w.(http.Flusher).Flush()
		}
	}
}
