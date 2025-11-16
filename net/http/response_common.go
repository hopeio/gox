package http

import (
	"io"
	"net/http"
)

type ICommonResponseWriter interface {
	Status(code int)
	Header() Header
	io.Writer
}

type CommonResponseWriter struct {
	http.ResponseWriter
}

func (w CommonResponseWriter) Status(code int) {
	w.WriteHeader(code)
}
func (w CommonResponseWriter) Header() Header {
	return (HttpHeader)(w.ResponseWriter.Header())
}
func (w CommonResponseWriter) Write(p []byte) (int, error) {
	return w.ResponseWriter.Write(p)
}

type ICommonRespond interface {
	CommonRespond(w ICommonResponseWriter) (int, error)
}
