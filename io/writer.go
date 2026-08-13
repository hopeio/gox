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

// Close closes and releases resources.
func (w *WriterToWrapper) Close() error {
	if w.close == nil {
		return nil
	}
	return w.close()
}

// WrapWriterTo returns the result.
func WrapWriterTo(w io.WriterTo, close func() error) *WriterToWrapper {
	return &WriterToWrapper{
		WriterTo: w,
		close:    close,
	}
}

type LimitedWriter []byte

// NewLimitedWriter creates and returns a new instance.
func NewLimitedWriter(max int64) LimitedWriter {
	return make([]byte, 0, max)
}

// Write appends up to the remaining capacity and reports io.EOF once truncation
// happens, so io.Copy/WriteTo callers can tell the limit was hit instead of
// silently believing everything was written.
func (lw *LimitedWriter) Write(p []byte) (int, error) {
	b := *lw
	l, c := len(b), cap(b)
	if l >= c {
		return 0, io.EOF
	}

	n := len(p)
	remaining := c - l
	if n > remaining {
		p = p[:remaining]
	}
	*lw = append(b, p...)
	if len(p) < n {
		// 旧实现截断后仍报告全量写入成功，调用方无从得知数据被丢弃
		return len(p), io.EOF
	}
	return len(p), nil
}
