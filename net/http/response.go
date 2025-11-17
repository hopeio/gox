/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
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

func (res *RespData[T]) Respond(w http.ResponseWriter) (int, error) {
	w.Header().Set(HeaderContentType, "application/json; charset=utf-8")
	jsonBytes, _ := json.Marshal(res)
	return w.Write(jsonBytes)
}

func (res *RespData[T]) ResponseStatus(w http.ResponseWriter, statusCode int) (int, error) {
	w.WriteHeader(statusCode)
	w.Header().Set(HeaderContentType, "application/json; charset=utf-8")
	jsonBytes, _ := json.Marshal(res)
	return w.Write(jsonBytes)
}

func NewRespData[T any](code errors.ErrCode, msg string, data T) *RespData[T] {
	return &RespData[T]{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

type RespAnyData = RespData[any]

func NewRespAnyData(code errors.ErrCode, msg string, data any) *RespAnyData {
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

func NewErrorRespData(code errors.ErrCode, msg string) *ErrResp {
	return &ErrResp{
		Code: code,
		Msg:  msg,
	}
}

func RespErrCodeMsg(w http.ResponseWriter, code errors.ErrCode, msg string) {
	NewRespData[any](code, msg, nil).Respond(w)
}

func RespErrRep(w http.ResponseWriter, rep *errors.ErrResp) (int, error) {
	return (*ErrResp)(rep).Respond(w)
}

func RespErrRepStatus(w http.ResponseWriter, rep *errors.ErrResp, statusCode int) (int, error) {
	return (*ErrResp)(rep).RespondStatus(w, statusCode)
}

func RespError(w http.ResponseWriter, err error) (int, error) {
	return ErrRespFrom(err).Respond(w)
}

func RespSuccess[T any](w http.ResponseWriter, msg string, data T) (int, error) {
	return NewRespData(errors.Success, msg, data).Respond(w)
}

func RespSuccessMsg(w http.ResponseWriter, msg string) (int, error) {
	return NewRespData[any](errors.Success, msg, nil).Respond(w)
}

func RespSuccessData(w http.ResponseWriter, data any) (int, error) {
	return NewRespData[any](errors.Success, errors.Success.String(), data).Respond(w)
}

func Respond[T any](w http.ResponseWriter, code errors.ErrCode, msg string, data T) (int, error) {
	return NewRespData(code, msg, data).Respond(w)
}

func RespondStatus[T any](w http.ResponseWriter, code errors.ErrCode, msg string, data T, statusCode int) (int, error) {
	return NewRespData(code, msg, data).ResponseStatus(w, statusCode)
}

func RespStreamWrite(w http.ResponseWriter, dataSource iter.Seq[[]byte]) {
	w.Header().Set(HeaderXAccelBuffering, "no") //nginx的锅必须加
	w.Header().Set(HeaderTransferEncoding, "chunked")
	notifyClosed := w.(http.CloseNotifier).CloseNotify()
	for data := range dataSource {
		select {
		// response writer forced to close, exit.
		case <-notifyClosed:
			return
		default:
			w.Write(data)
			w.(http.Flusher).Flush()
		}
	}
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

func (res *Response) Respond(w http.ResponseWriter) (int, error) {
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

func (res *Response) CommonRespond(w ICommonResponseWriter) (int, error) {
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

func (res *ErrResp) Respond(w http.ResponseWriter) (int, error) {
	w.Header().Set(HeaderContentType, ContentTypeJsonUtf8)
	jsonBytes, _ := json.Marshal(res)
	return w.Write(jsonBytes)
}

func (res *ErrResp) RespondStatus(w http.ResponseWriter, statusCode int) (int, error) {
	w.WriteHeader(statusCode)
	w.Header().Set(HeaderContentType, ContentTypeJsonUtf8)
	jsonBytes, _ := json.Marshal(res)
	return w.Write(jsonBytes)
}

type IRespond interface {
	Respond(w http.ResponseWriter) (int, error)
}

type ResponseStream struct {
	Status  int              `json:"status,omitempty"`
	Headers http.Header      `json:"header,omitempty"`
	Body    iter.Seq[[]byte] `json:"body,omitempty"`
}

func (res *ResponseStream) Respond(w http.ResponseWriter) (int, error) {
	return res.CommonRespond(CommonResponseWriter{w})
}

func (res *ResponseStream) CommonRespond(w ICommonResponseWriter) (int, error) {
	header := w.Header()
	HttpHeaderIntoHeader(res.Headers, header)
	header.Set(HeaderTransferEncoding, "chunked")
	notifyClosed := w.(http.CloseNotifier).CloseNotify()
	var n int
	for data := range res.Body {
		select {
		// response writer forced to close, exit.
		case <-notifyClosed:
			return n, nil
		default:
			write, err := w.Write(data)
			if err != nil {
				return 0, err
			}
			n += write
			w.(http.Flusher).Flush()
		}
	}
	return n, nil
}
