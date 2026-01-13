package io

import (
	"io"
)

type WriterToCloser interface {
	io.WriterTo
	io.Closer
}

type ReadWriterToCloser interface {
	io.Reader
	io.WriterTo
	io.Closer
}

type WriterToWrapper struct {
	io.WriterTo
	close func() error
}

func (w *WriterToWrapper) Close() error {
	if w.close == nil {
		return nil
	}
	return w.close()
}

func WrapWriterTo(w io.WriterTo, close func() error) *WriterToWrapper {
	return &WriterToWrapper{
		WriterTo: w,
		close:    close,
	}
}

type LimitedWriter []byte

func NewLimitedWriter(max int64) LimitedWriter {
	return make([]byte, 0, max)
}

func (lw *LimitedWriter) Write(p []byte) (int, error) {
	b := *lw
	l, c := len(b), cap(b)
	if l >= c {
		return 0, io.EOF
	}

	remaining := c - l
	if len(p) > remaining {
		p = p[:remaining]
	}
	*lw = append(b, p...)
	if len(p) > remaining {
		return len(p), io.EOF
	}
	return len(p), nil
}
