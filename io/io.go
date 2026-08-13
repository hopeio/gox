/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package io

import (
	"context"
	"io"

	"golang.org/x/time/rate"
)

type limitReader struct {
	r       io.Reader
	ctx     context.Context
	limiter *rate.Limiter
}

// LimitReader returns a reader that is rate limited by
// the given token bucket. Each token in the bucket
// represents one byte.
func LimitReader(r io.Reader, ctx context.Context, limiter *rate.Limiter) io.Reader {
	return &limitReader{
		r:       r,
		ctx:     ctx,
		limiter: limiter,
	}
}

// Read reads at most one burst worth of bytes and waits for tokens matching
// the bytes actually read (short reads don't overpay the limiter).
// 旧实现的 for 循环 end 从不更新且内层 n 遮蔽外层，读偏移错位、短 buf 直接返回 (0,nil)。
func (r *limitReader) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	chunk := min(len(buf), r.limiter.Burst())
	n, err := r.r.Read(buf[:chunk])
	if n > 0 {
		if werr := r.limiter.WaitN(r.ctx, n); werr != nil {
			return n, werr
		}
	}
	return n, err
}

type limitWriter struct {
	w       io.Writer
	ctx     context.Context
	limiter *rate.Limiter
}

// LimitWriter returns a reader that is rate limited by
// the given token bucket. Each token in the bucket
// represents one byte.
func LimitWriter(w io.Writer, ctx context.Context, limiter *rate.Limiter) io.Writer {
	return &limitWriter{
		w:       w,
		ctx:     ctx,
		limiter: limiter,
	}
}

// Write writes buf in burst-sized chunks, waiting for tokens before each chunk.
// io.Writer 契约要求全量写出，因此循环推进直到写完或出错。
func (w *limitWriter) Write(buf []byte) (int, error) {
	burst := w.limiter.Burst()
	var n int
	for n < len(buf) {
		chunk := min(len(buf)-n, burst)
		if err := w.limiter.WaitN(w.ctx, chunk); err != nil {
			return n, err
		}
		m, err := w.w.Write(buf[n : n+chunk])
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, nil
}
