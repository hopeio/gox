/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package io

import (
	"bufio"
	"io"
)

// ReadLines performs the operation.
func ReadLines(reader io.Reader, f func(line string) bool) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if !f(scanner.Text()) {
			return nil
		}
	}
	return scanner.Err()
}

type ReadCloserWrapper struct {
	io.ReadCloser
}

// WriteTo performs the operation.
func (r ReadCloserWrapper) WriteTo(w io.Writer) (int64, error) {
	return io.Copy(w, r.ReadCloser)
}

type RawByter interface {
	Raw() []byte
}

type RawBytes []byte

// WriteTo performs the operation.
func (res RawBytes) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(res)
	return int64(n), err
}

// Close closes and releases resources.
func (res RawBytes) Close() error {
	return nil
}

// Write performs the operation.
func (res *RawBytes) Write(p []byte) (int, error) {
	*res = append(*res, p...)
	return len(p), nil
}

// Read consumes the buffer. 值接收者版本对切片头的推进不生效，
// 会永远从头读且不返回 EOF，导致 io.Copy 死循环。
func (res *RawBytes) Read(p []byte) (int, error) {
	if len(*res) == 0 {
		return 0, io.EOF
	}
	n := copy(p, *res)
	*res = (*res)[n:]
	return n, nil
}

// Raw returns the result.
func (res RawBytes) Raw() []byte {
	return res
}

type ReadWrapper struct {
	io.Reader
	close func() error
}

// WriteTo performs the operation.
func (r *ReadWrapper) WriteTo(w1 io.Writer) (int64, error) {
	return io.Copy(w1, r.Reader)
}

// Close closes and releases resources.
func (r *ReadWrapper) Close() error {
	if r.close == nil {
		return nil
	}
	return r.close()
}

// WrapReader returns the result.
func WrapReader(r io.Reader, close func() error) *ReadWrapper {
	return &ReadWrapper{
		Reader: r,
		close:  close,
	}
}
