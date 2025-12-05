/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package context

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	Tracer = otel.Tracer("context")
	Meter  = otel.Meter("context")
)

const (
	KindKey = attribute.Key("context.key")
)

var (
	serverKeyValue = KindKey.String("server")
)

func StartSpan(ctx context.Context, name string, o ...trace.SpanStartOption) (context.Context, trace.Span) {
	if name == "" {
		pc, _, _, _ := runtime.Caller(4)
		name = runtime.FuncForPC(pc).Name()
	}
	return Tracer.Start(ctx, name, o...)
}

func (c *Context) StartSpan(name string, o ...trace.SpanStartOption) trace.Span {
	ctx, span := Tracer.Start(c.ctx, name, o...)
	c.ctx = ctx
	if c.traceID == "" {
		c.traceID = span.SpanContext().TraceID().String()
	}
	return span
}

func (c *Context) StartSpanEnd(name string, o ...trace.SpanStartOption) func(options ...trace.SpanEndOption) {
	ctx, span := Tracer.Start(c.ctx, name, o...)
	c.ctx = ctx
	if c.traceID == "" {
		c.traceID = span.SpanContext().TraceID().String()
	}
	return span.End
}
