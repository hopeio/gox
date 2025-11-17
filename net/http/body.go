/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"io"
)

type NoCloseBody struct {
	data       []byte
	index      int64 // current reading index
	prevRune   int   // index of previous rune; or < 0
	closeTimes int64
}

func (r *NoCloseBody) Len() int {
	if r.index >= int64(len(r.data)) {
		return 0
	}
	return int(int64(len(r.data)) - r.index)
}

func (r *NoCloseBody) Size() int64 { return int64(len(r.data)) }

func (r *NoCloseBody) Read(b []byte) (n int, err error) {
	if r.index >= int64(len(r.data)) {
		return 0, io.EOF
	}
	r.prevRune = -1
	n = copy(b, r.data[r.index:])
	r.index += int64(n)
	return
}

func (r *NoCloseBody) Close() error {
	r.index = 0
	r.closeTimes++
	return nil
}

// 适用于轮询
func NewNoCloseBody(s []byte) *NoCloseBody { return &NoCloseBody{s, 0, -1, 0} }
