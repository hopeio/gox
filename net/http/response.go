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

type Body map[string]any

// RespData 主要用来接收返回，发送请使用ResAnyData
type RespData[T any] struct {
	Code errors.ErrCode `json:"code"`
	Msg  string         `json:"msg,omitempty"`
	//验证码
	Data T `json:"data,omitempty"`
}

func (res *RespData[T]) Response(w http.ResponseWriter) (int, error) {
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

func NewErrorRespData(code errors.ErrCode, msg string) *ErrRep {
	return &ErrRep{
		Code: code,
		Msg:  msg,
	}
}

func RespErrCodeMsg(w http.ResponseWriter, code errors.ErrCode, msg string) {
	NewRespData[any](code, msg, nil).Response(w)
}

func RespErrRep(w http.ResponseWriter, rep *errors.ErrRep) (int, error) {
	return (*ErrRep)(rep).Response(w)
}

func RespErrRepStatus(w http.ResponseWriter, rep *errors.ErrRep, statusCode int) (int, error) {
	return (*ErrRep)(rep).ResponseStatus(w, statusCode)
}

func RespError(w http.ResponseWriter, err error) (int, error) {
	return ErrRepFrom(err).Response(w)
}

func RespSuccess[T any](w http.ResponseWriter, msg string, data T) (int, error) {
	return NewRespData(errors.Success, msg, data).Response(w)
}

func RespSuccessMsg(w http.ResponseWriter, msg string) (int, error) {
	return NewRespData[any](errors.Success, msg, nil).Response(w)
}

func RespSuccessData(w http.ResponseWriter, data any) (int, error) {
	return NewRespData[any](errors.Success, errors.Success.String(), data).Response(w)
}

func Response[T any](w http.ResponseWriter, code errors.ErrCode, msg string, data T) (int, error) {
	return NewRespData(code, msg, data).Response(w)
}

func ResponseStatus[T any](w http.ResponseWriter, code errors.ErrCode, msg string, data T, statusCode int) (int, error) {
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

var ResponseSysErr = json.RawMessage(`{"code":-1,"msg":"system error"}`)
var ResponseOk = json.RawMessage(`{"code":0}`)

type ReceiveData = RespData[json.RawMessage]

func NewReceiveData(code errors.ErrCode, msg string, data any) *ReceiveData {
	jsonBytes, _ := json.Marshal(data)
	return &ReceiveData{
		Code: code,
		Msg:  msg,
		Data: jsonBytes,
	}
}

type HttpResponseRawBody struct {
	Status  int       `json:"status,omitempty"`
	Headers MapHeader `json:"header,omitempty"`
	Body    []byte    `json:"body,omitempty"`
}

func (res *HttpResponseRawBody) Response(w http.ResponseWriter) (int, error) {
	w.WriteHeader(res.Status)
	header := w.Header()
	for k, v := range res.Headers {
		header.Set(k, v)
	}
	return w.Write(res.Body)
}

func (res *HttpResponseRawBody) CommonResponse(w CommonResponseWriter) (int, error) {
	w.WriteHeader(res.Status)
	header := w.Header()
	for k, v := range res.Headers {
		header.Set(k, v)
	}
	return w.Write(res.Body)
}

type HttpResponse struct {
	Status  int            `json:"status,omitempty"`
	Headers MapHeader      `json:"header,omitempty"`
	Body    WriterToCloser `json:"body,omitempty"`
}

type WriterToCloser interface {
	io.WriterTo
	io.Closer
}

func (res *HttpResponse) Response(w http.ResponseWriter) (int, error) {
	w.WriteHeader(res.Status)
	for k, v := range res.Headers {
		w.Header().Set(k, v)
	}
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

func (res *HttpResponse) CommonResponse(w CommonResponseWriter) (int, error) {
	w.WriteHeader(res.Status)
	for k, v := range res.Headers {
		w.Header().Set(k, v)
	}
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

type ErrRep errors.ErrRep

func ErrRepFrom(err error) *ErrRep {
	return (*ErrRep)(errors.ErrRepFrom(err))
}

func (res *ErrRep) Response(w http.ResponseWriter) (int, error) {
	w.Header().Set(HeaderContentType, ContentTypeJsonUtf8)
	jsonBytes, _ := json.Marshal(res)
	return w.Write(jsonBytes)
}

func (res *ErrRep) ResponseStatus(w http.ResponseWriter, statusCode int) (int, error) {
	w.WriteHeader(statusCode)
	w.Header().Set(HeaderContentType, ContentTypeJsonUtf8)
	jsonBytes, _ := json.Marshal(res)
	return w.Write(jsonBytes)
}

type IHttpResponseTo interface {
	Response(w http.ResponseWriter) (int, error)
}

type HttpResponseStream struct {
	Status  int              `json:"status,omitempty"`
	Headers MapHeader        `json:"header,omitempty"`
	Body    iter.Seq[[]byte] `json:"body,omitempty"`
}

func (res *HttpResponseStream) Response(w http.ResponseWriter) (int, error) {
	return res.CommonResponse(CommonResponseWriter{w})
}

func (res *HttpResponseStream) CommonResponse(w ICommonResponseWriter) (int, error) {
	header := w.Header()
	for k, v := range res.Headers {
		header.Set(k, v)
	}
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

type RawBody []byte

func (res RawBody) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(res)
	return int64(n), err
}

func (res RawBody) Closer() error {
	return nil
}
