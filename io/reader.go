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

func (r ReadCloserWrapper) WriteTo(w io.Writer) (int64, error) {
	return io.Copy(w, r)
}

type RawBytes []byte

func (res RawBytes) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(res)
	return int64(n), err
}

func (res RawBytes) Close() error {
	return nil
}

func (res *RawBytes) Write(p []byte) (int, error) {
	*res = append(*res, p...)
	return len(p), nil
}
