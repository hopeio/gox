/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package limit

import (
	"context"
	"io"

	"golang.org/x/time/rate"
)

type reader struct {
	r       io.Reader
	ctx     context.Context
	limiter *rate.Limiter
}

// Reader returns a reader that is rate limited by
// the given token bucket. Each token in the bucket
// represents one byte.
func Reader(r io.Reader, ctx context.Context, limiter *rate.Limiter) io.Reader {
	return &reader{
		r:       r,
		ctx:     ctx,
		limiter: limiter,
	}
}

func (r *reader) Read(buf []byte) (int, error) {
	burst := r.limiter.Burst()
	l := len(buf)
	var n int
	for end := n + burst; end <= l; n = end {
		err := r.limiter.WaitN(r.ctx, burst)
		if err != nil {
			return n, err
		}
		n, err := r.r.Read(buf[n:end])
		if n <= 0 {
			return n, err
		}
	}
	return n, nil
}

type writer struct {
	w       io.Writer
	ctx     context.Context
	limiter *rate.Limiter
}

// Writer returns a reader that is rate limited by
// the given token bucket. Each token in the bucket
// represents one byte.
func Writer(w io.Writer, ctx context.Context, limiter *rate.Limiter) io.Writer {
	return &writer{
		w:       w,
		ctx:     ctx,
		limiter: limiter,
	}
}

func (w *writer) Write(buf []byte) (int, error) {
	burst := w.limiter.Burst()
	l := len(buf)
	var n int
	for end := n + burst; end <= l; n = end {
		err := w.limiter.WaitN(w.ctx, burst)
		if err != nil {
			return n, err
		}
		n, err := w.w.Write(buf[n:end])
		if n <= 0 {
			return n, err
		}
	}
	return n, nil
}
