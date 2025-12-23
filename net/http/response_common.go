package http

import (
	"context"
	"io"
	"iter"
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

func (w ResponseWriterWrapper) RespondStream(ctx context.Context, seq iter.Seq[WriterToCloser]) (int, error) {
	return RespondStream(ctx, w.ResponseWriter, seq)
}

type CommonResponder interface {
	CommonRespond(ctx context.Context, w CommonResponseWriter) (int, error)
}

type CommonRequestWriter interface {
	io.Reader
}

type RespondStreamer interface {
	RespondStream(ctx context.Context, seq iter.Seq[WriterToCloser]) (int, error)
}

func CommonRespond(ctx context.Context, w CommonResponseWriter, res any) (int, error) {
	header := w.Header()
	if err, ok := res.(error); ok {
		return ErrRespFrom(err).CommonRespond(ctx, w)
	}
	data, contentType := DefaultMarshal("", res)
	header.Set(HeaderContentType, contentType)
	return w.Write(data)
}
