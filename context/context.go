/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package context

import (
	"context"
	"sync"
	"time"

	"github.com/hopeio/gox/idgen"
	"go.opentelemetry.io/otel/trace"
)

type Context struct {
	ctx      context.Context
	rootSpan trace.Span
	traceID  string
	sync.RWMutex
}

func New(ctx context.Context) *Context {
	c, ok := FromContext(ctx)
	if ok {
		return c
	}
	var traceId string
	var rootSpan trace.Span
	if ctx != nil {
		span := trace.SpanFromContext(ctx)
		if spanContext := span.SpanContext(); spanContext.IsValid() {
			traceId = spanContext.TraceID().String()
			rootSpan = span
		}
	} else {
		ctx = context.Background()
	}
	if rootSpan == nil || traceId == "" {
		ctx, rootSpan = StartSpan(ctx, "NewContext")
		if spanContext := rootSpan.SpanContext(); spanContext.IsValid() {
			traceId = spanContext.TraceID().String()
		} else {
			traceId = idgen.UniqueID().String()
		}
	}
	return &Context{
		ctx:      ctx,
		rootSpan: rootSpan,
		traceID:  traceId,
	}
}

func (c *Context) TraceID() string {
	return c.traceID
}

type ctxKey struct{}

func (c *Context) Wrapper() context.Context {
	return context.WithValue(c.ctx, ctxKey{}, c)
}

func WrapperKey() ctxKey {
	return ctxKey{}
}

func FromContext(ctx context.Context) (*Context, bool) {
	if ctx == nil {
		return nil, false
	}
	ctxi := ctx.Value(ctxKey{})
	c, ok := ctxi.(*Context)
	if !ok {
		return nil, false
	}
	c.ctx = ctx
	return c, ok
}

func (c *Context) Base() context.Context {
	return c.ctx
}

func (c *Context) SetBase(ctx context.Context) {
	c.ctx = ctx
}

func (c *Context) RootSpan() trace.Span {
	return c.rootSpan
}

func (c *Context) WithTimeout(timeout time.Duration) context.CancelFunc {
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithTimeout(c.ctx, timeout)
	return cancel
}

func (c *Context) WithTimeoutCause(timeout time.Duration, cause error) context.CancelFunc {
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithTimeoutCause(c.ctx, timeout, cause)
	return cancel
}

func (c *Context) WithCancel() context.CancelFunc {
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithCancel(c.ctx)
	return cancel
}

func (c *Context) WithoutCancel() {
	c.ctx = context.WithoutCancel(c.ctx)
}

func (c *Context) WithValue(key, val any) {
	c.ctx = context.WithValue(c.ctx, key, val)
}

func (c *Context) WithCancelCause() context.CancelCauseFunc {
	var cancel context.CancelCauseFunc
	c.ctx, cancel = context.WithCancelCause(c.ctx)
	return cancel
}

func (c *Context) WithDeadline(d time.Time) context.CancelFunc {
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithDeadline(c.ctx, d)
	return cancel
}

func (c *Context) WithDeadlineCause(d time.Time, cause error) context.CancelFunc {
	var cancel context.CancelFunc
	c.ctx, cancel = context.WithDeadlineCause(c.ctx, d, cause)
	return cancel
}

func (c *Context) Value(key any) any {
	return c.ctx.Value(key)
}
