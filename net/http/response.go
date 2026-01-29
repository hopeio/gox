/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"context"
	"fmt"
	"io"
	"iter"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	errorsx "github.com/hopeio/gox/errors"
	iox "github.com/hopeio/gox/io"
	"github.com/hopeio/gox/strings"
)

type ResponseWriter interface {
	WriteHeader(code int)
	HeaderX() Header
	Write([]byte) (int, error)
}

type Responder interface {
	Respond(ctx context.Context, w http.ResponseWriter) (int, error)
}

var errCodeHttpStatusMap = map[errorsx.ErrCode]int{
	errorsx.Success: http.StatusOK,
}

func RegisterErrCodeHttpStatus(code errorsx.ErrCode, status int) {
	errCodeHttpStatusMap[code] = status
}

type errHeaderKey struct{}

var ErrHeaderKey errHeaderKey

func RespodWithErrHeader(ctx context.Context) context.Context {
	return context.WithValue(ctx, ErrHeaderKey, true)
}

func StatusFromErrCode(code errorsx.ErrCode) int {
	if status, ok := errCodeHttpStatusMap[code]; ok {
		return status
	}
	return http.StatusOK
}

// CommonResp 主要用来接收返回，发送请使用 CommonAnyResp
type CommonResp[T any] struct {
	Code errorsx.ErrCode `json:"code"`
	Msg  string          `json:"msg,omitempty"`
	//验证码
	Data T `json:"data,omitempty"`
}

func (res *CommonResp[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *CommonResp[T]) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	if res.Code != errorsx.Success {
		w.WriteHeader(StatusFromErrCode(res.Code))
	} else {
		w.WriteHeader(http.StatusOK)
	}
	data, contentType := DefaultMarshal(ctx, res)
	if wx, ok := w.(ResponseWriter); ok {
		header := wx.HeaderX()
		if res.Code != errorsx.Success && ctx.Value(ErrHeaderKey) != nil {
			header.Set(HeaderErrorCode, strconv.Itoa(int(res.Code)))
			header.Set(HeaderErrorMsg, res.Msg)
		}

		header.Set(HeaderContentType, contentType)
	} else {
		header := w.Header()
		if res.Code != errorsx.Success && ctx.Value(ErrHeaderKey) != nil {
			header.Set(HeaderErrorCode, strconv.Itoa(int(res.Code)))
			header.Set(HeaderErrorMsg, res.Msg)
		}
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

type CommonAnyResp = CommonResp[any]

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
	Status  int                `json:"status,omitempty"`
	Headers http.Header        `json:"header,omitempty"`
	Body    iox.WriterToCloser `json:"body,omitempty"`
}

func (res *Response) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *Response) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	w.WriteHeader(res.Status)
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
	if res.Code != errorsx.Success {
		w.WriteHeader(StatusFromErrCode(res.Code))
	} else {
		w.WriteHeader(http.StatusOK)
	}
	data, contentType := DefaultMarshal(ctx, res)
	if wx, ok := w.(ResponseWriter); ok {
		header := wx.HeaderX()
		if res.Code != errorsx.Success && ctx.Value(ErrHeaderKey) != nil {
			header.Set(HeaderErrorCode, strconv.Itoa(int(res.Code)))
			header.Set(HeaderErrorMsg, res.Msg)
		}
		header.Set(HeaderContentType, contentType)
	} else {
		header := w.Header()
		if res.Code != errorsx.Success && ctx.Value(ErrHeaderKey) != nil {
			header.Set(HeaderErrorCode, strconv.Itoa(int(res.Code)))
			header.Set(HeaderErrorMsg, res.Msg)
		}
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

type ResponseStream struct {
	Status  int                          `json:"status,omitempty"`
	Headers http.Header                  `json:"header,omitempty"`
	Body    iter.Seq[iox.WriterToCloser] `json:"body,omitempty"`
}

func (res *ResponseStream) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	w.WriteHeader(res.Status)
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
	return RespondStream(ctx, w, res.Body)
}

func (res *ResponseStream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func RespondStream(ctx context.Context, w http.ResponseWriter, dataSource iter.Seq[iox.WriterToCloser]) (int, error) {
	if wx, ok := w.(ResponseWriter); ok {
		header := wx.HeaderX()
		header.Set(HeaderCacheControl, CacheControlNoCache)
		header.Set(HeaderTransferEncoding, TransferEncodingChunked)
	} else {
		header := w.Header()
		header.Set(HeaderCacheControl, CacheControlNoCache)
		header.Set(HeaderTransferEncoding, TransferEncodingChunked)
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

type SSEData string

func (data SSEData) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(strings.ToBytes(fmt.Sprintf("data: %s\n\n", data)))
	return int64(n), err
}
func (data SSEData) Close() error {
	return nil
}

func RespondSSE[T ~string](ctx context.Context, w http.ResponseWriter, dataSource iter.Seq[T]) (int, error) {
	if wx, ok := w.(ResponseWriter); ok {
		header := wx.HeaderX()
		header.Set(HeaderContentType, ContentTypeTextEventStream)
	} else {
		header := w.Header()
		header.Set(HeaderContentType, ContentTypeTextEventStream)
	}
	return RespondStream(ctx, w, func(yield func(iox.WriterToCloser) bool) {
		for data := range dataSource {
			yield(SSEData(data))
		}
	})
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

type ResponseFile struct {
	Name        string             `json:"name"`
	Body        iox.WriterToCloser `json:"body"`
	ContentType string
}

func (res *ResponseFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res.Respond(r.Context(), w)
}

func (res *ResponseFile) Respond(ctx context.Context, w http.ResponseWriter) (int, error) {
	contentType := res.ContentType
	var limitedWriter iox.LimitedWriter
	if contentType == "" {
		if ext := filepath.Ext(res.Name); ext != "" {
			contentType = mime.TypeByExtension(ext)
			if contentType == "" {
				contentType = ContentTypeOctetStream
			}
		} else {
			limitedWriter = iox.NewLimitedWriter(512)
			_, err := res.Body.WriteTo(&limitedWriter)
			if err != nil {
				return 0, err
			}
			contentType = http.DetectContentType(limitedWriter)
		}
	}
	contentDisposition := "inline"
	if res.Name != "" {
		contentDisposition = fmt.Sprintf(AttachmentTmpl, res.Name)
	}
	if wx, ok := w.(ResponseWriter); ok {
		header := wx.HeaderX()
		header.Set(HeaderContentType, contentType)
		header.Set(HeaderContentDisposition, contentDisposition)
	} else {
		header := w.Header()
		header.Set(HeaderContentType, contentType)
		header.Set(HeaderContentDisposition, contentDisposition)
	}
	if len(limitedWriter) > 0 {
		n, err := w.Write(limitedWriter)
		if err != nil {
			return n, err
		}
	}
	n, err := res.Body.WriteTo(w)
	res.Body.Close()
	if len(limitedWriter) > 0 {
		n += int64(len(limitedWriter))
	}
	return int(n), err
}
