/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

// Named adds a sub-scope to the logger's name. See Logger.Named for details.
func (l *Logger) Named(name string) *Logger {
	return &Logger{l.Logger.Named(name)}
}

// WithOptions wrap the zap WithOptions, applies the supplied Options, and
// returns the resulting Logger. It's safe to use concurrently.
func (l *Logger) WithOptions(opts ...zap.Option) *Logger {
	return &Logger{l.Logger.WithOptions(opts...)}
}

// With wrap the zap With. Fields added
// to the child don't affect the parent, and vice versa.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

// Sugar wrap the zap Sugar.
func (l *Logger) Zap() *zap.Logger {
	return l.Logger
}

// Sugar wrap the zap Sugar.
func (l *Logger) Sugar() *zap.SugaredLogger {
	l.WithOptions(zap.AddCallerSkip(-1))
	return l.Logger.Sugar()
}

// AddCore wrap the zap AddCore.
func (l *Logger) AddCore(newCore zapcore.Core) *Logger {
	return l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, newCore)
	}))
}

// AddSkip wrap the zap AddCallerSkip.
func (l *Logger) AddSkip(skip int) *Logger {
	return &Logger{l.Logger.WithOptions(zap.AddCallerSkip(skip))}
}

func (l *Logger) Printf(template string, args ...any) {
	if ce := l.Check(zap.InfoLevel, ""); ce != nil {
		ce.Write()
	}
}

// 兼容gormv1
func (l *Logger) Print(args ...any) {
	if ce := l.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Debug uses fmt.Sprint to construct and log a message.
func (l *Logger) Debug(args ...any) {
	if ce := l.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Info uses fmt.Sprint to construct and log a message.
func (l *Logger) Info(args ...any) {
	if ce := l.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Warn uses fmt.Sprint to construct and log a message.
func (l *Logger) Warn(args ...any) {
	if ce := l.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Error uses fmt.Sprint to construct and log a message.
func (l *Logger) Error(args ...any) {
	if ce := l.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (l *Logger) DPanic(args ...any) {
	if ce := l.Check(zap.DPanicLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (l *Logger) Panic(args ...any) {
	if ce := l.Check(zap.PanicLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (l *Logger) Fatal(args ...any) {
	if ce := l.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Debugw(msg string, fields ...zap.Field) {
	if ce := l.Check(zap.DebugLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Infow(msg string, fields ...zap.Field) {
	if ce := l.Check(zap.InfoLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Warnw(msg string, fields ...zap.Field) {
	if ce := l.Check(zap.WarnLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Logger) Errorw(msg string, fields ...zap.Field) {
	if ce := l.Check(zap.ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func (l *Logger) DPanicw(msg string, fields ...zap.Field) {
	if ce := l.Check(zap.DPanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (l *Logger) Panicw(msg string, fields ...zap.Field) {
	if ce := l.Check(zap.PanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (l *Logger) Fatalw(msg string, fields ...zap.Field) {
	if ce := l.Check(zap.FatalLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Debugf uses fmt.Sprintf to log a templated message.
func (l *Logger) Debugf(template string, args ...any) {
	if ce := l.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *Logger) Infof(template string, args ...any) {
	if ce := l.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Warnf uses fmt.Sprintf to log a templated message.
func (l *Logger) Warnf(template string, args ...any) {
	if ce := l.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *Logger) Errorf(template string, args ...any) {
	if ce := l.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (l *Logger) DPanicf(template string, args ...any) {
	if ce := l.Check(zap.DPanicLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (l *Logger) Panicf(template string, args ...any) {
	if ce := l.Check(zap.PanicLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (l *Logger) Fatalf(template string, args ...any) {
	if ce := l.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

func (l *Logger) Println(args ...any) {
	if ce := l.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = fmt.Sprint(args...)
		ce.Write()
	}
}

type StdOutLevel zapcore.Level

func (l StdOutLevel) Enabled(lvl zapcore.Level) bool {
	return lvl >= zapcore.Level(l) && lvl < zapcore.ErrorLevel
}

type StdErrLevel zapcore.Level

func (l StdErrLevel) Enabled(lvl zapcore.Level) bool {
	return lvl >= zapcore.Level(l) && lvl >= zapcore.ErrorLevel
}
