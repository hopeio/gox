package log

import (
	"context"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewOtelLogger creates and returns a new instance.
func NewOtelLogger(name string, opts ...otelzap.Option) *Logger {
	return &Logger{zap.New(otelzap.NewCore(name, opts...), zap.AddCallerSkip(1))}
}

// WithContext ...
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	if !spanContext.IsValid() {
		return l
	}
	return l.With(zap.String(FieldTraceId, spanContext.TraceID().String()), zap.String(FieldSpanId, spanContext.SpanID().String()), zapcore.Field{
		Type:      zapcore.SkipType,
		Interface: ctx,
	})
}

// Context ...
func Context(ctx context.Context) zapcore.Field {
	return zapcore.Field{
		Type: zapcore.InlineMarshalerType,
		Interface: contextObjectMarshaler{
			Context: ctx,
		},
	}
}

type contextObjectMarshaler struct {
	context.Context
}

// MarshalLogObject ...
func (m contextObjectMarshaler) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	spanContext := trace.SpanContextFromContext(m.Context)
	if spanContext.IsValid() {
		enc.AddString(FieldTraceId, spanContext.TraceID().String())
		enc.AddString(FieldSpanId, spanContext.SpanID().String())
	}
	return nil
}
