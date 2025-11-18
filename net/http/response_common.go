package http

import (
	"context"
	"io"
	"net/http"
)

type CommonResponseWriter interface {
	Status(code int)
	Header() Header
	io.Writer
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
}

func (w ResponseWriterWrapper) Status(code int) {
	w.WriteHeader(code)
}
func (w ResponseWriterWrapper) Header() Header {
	return (HttpHeader)(w.ResponseWriter.Header())
}
func (w ResponseWriterWrapper) Write(p []byte) (int, error) {
	return w.ResponseWriter.Write(p)
}

type CommonResponder interface {
	CommonRespond(ctx context.Context, w CommonResponseWriter) (int, error)
}

type CommonRequestWriter interface {
	io.Reader
}
