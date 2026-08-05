/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// init ...
func init() {
	SetDefaultLogger((&Config{
		Config: zap.Config{
			Development: true,
			Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		},
	}).NewLogger())
}

type skipLogger struct {
	*Logger
	needUpdate bool
}

var (
	defaultLogger  *Logger
	stackLogger    *Logger
	noCallerLogger *Logger
	skipLoggers    = make([]skipLogger, 10)
	mu             sync.Mutex
)

// DefaultLogger ...
func DefaultLogger() *Logger {
	return defaultLogger
}

// SetDefaultLogger ...
func SetDefaultLogger(logger *Logger) {
	mu.Lock()
	defer mu.Unlock()

	defaultLogger = logger
	stackLogger = defaultLogger.WithOptions(zap.WithCaller(true), zap.AddStacktrace(zapcore.DebugLevel))
	noCallerLogger = defaultLogger.WithOptions(zap.WithCaller(false))
	for i := range len(skipLoggers) {
		if skipLoggers[i].Logger != nil {
			skipLoggers[i].needUpdate = true
		}
	}
}

// range -3~6
func CallerSkipLogger(skip int) *Logger {
	if skip < -3 {
		panic("skip not less than -3")
	}
	if skip > 6 {
		panic("skip not great than 6")
	}
	idx := skip + 3
	if skipLoggers[idx].needUpdate || skipLoggers[idx].Logger == nil {
		mu.Lock()
		skipLoggers[idx].Logger = defaultLogger.AddCallerSkip(skip)
		skipLoggers[idx].needUpdate = false
		mu.Unlock()
	}
	return skipLoggers[idx].Logger
}

// NoCallerLogger ...
func NoCallerLogger() *Logger {
	return noCallerLogger
}

// StackLogger ...
func StackLogger() *Logger {
	return stackLogger
}

// Sync ...
func Sync() error {
	return defaultLogger.Sync()
}

// Debug ...
func Debug(args ...any) {
	if ce := defaultLogger.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Info ...
func Info(args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Warn ...
func Warn(args ...any) {
	if ce := defaultLogger.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Error returns the error message string.
func Error(args ...any) {
	if ce := defaultLogger.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Panic ...
func Panic(args ...any) {
	if ce := defaultLogger.Check(zap.PanicLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Fatal ...
func Fatal(args ...any) {
	if ce := defaultLogger.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Printf ...
func Printf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Debugf ...
func Debugf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Infof ...
func Infof(template string, args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Warnf ...
func Warnf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Errorf ...
func Errorf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Panicf ...
func Panicf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.PanicLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Fatalf ...
func Fatalf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Debugw ...
func Debugw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.DebugLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Infow ...
func Infow(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.InfoLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Warnw ...
func Warnw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.WarnLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Errorw ...
func Errorw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Panicw ...
func Panicw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.PanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Fatalw ...
func Fatalw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.FatalLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Log ...
func Log(lvl zapcore.Level, args ...any) {
	if ce := defaultLogger.Check(lvl, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Logf ...
func Logf(lvl zapcore.Level, msg string, args ...any) {
	if ce := defaultLogger.Check(lvl, ""); ce != nil {
		ce.Message = fmt.Sprintf(msg, args...)
		ce.Write()
	}
}

// Logw ...
func Logw(lvl zapcore.Level, msg string, fields ...zapcore.Field) {
	if ce := defaultLogger.Check(lvl, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Check ...
func Check(lvl zapcore.Level, args ...any) *zapcore.CheckedEntry {
	ce := defaultLogger.Check(lvl, "")
	if ce != nil {
		ce.Message = sprintln(args...)
	}
	return ce
}

// Println ...
func Println(args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}
