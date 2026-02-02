/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import (
	"context"
	"log/slog"
	"runtime"

	"github.com/hopeio/gox/slices"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ slog.Handler = &Logger{}

func (l *Logger) NewSLogger() *slog.Logger {
	return slog.New(l)
}

func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.Logger.Core().Enabled(zapcore.Level(level / 4))
}

func (l *Logger) Handle(ctx context.Context, record slog.Record) error {
	core := l.Logger.Core()
	ent := zapcore.Entry{
		LoggerName: l.Name(),
		Time:       record.Time,
		Level:      zapcore.Level(record.Level / 4),
		Message:    record.Message,
	}
	ce := core.Check(ent, nil)
	fs := runtime.CallersFrames([]uintptr{record.PC})
	frame, _ := fs.Next()
	ce.Caller = zapcore.EntryCaller{
		Defined:  frame.PC != 0,
		PC:       frame.PC,
		File:     frame.File,
		Line:     frame.Line,
		Function: frame.Function,
	}

	ce.Write()
	return nil
}

func (l *Logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return l.With(slices.Map(attrs, func(attr slog.Attr) zap.Field {
		return zap.String(attr.Key, attr.Value.String())
	})...)
}

func (l *Logger) WithGroup(name string) slog.Handler {
	return l.Named(name)
}
