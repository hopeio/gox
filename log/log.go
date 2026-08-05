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

// init initializes package state.
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

// DefaultLogger returns the result.
func DefaultLogger() *Logger {
	return defaultLogger
}

// SetDefaultLogger updates or inserts a value.
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

// NoCallerLogger returns the result.
func NoCallerLogger() *Logger {
	return noCallerLogger
}

// StackLogger returns the result.
func StackLogger() *Logger {
	return stackLogger
}

// Sync performs the operation.
func Sync() error {
	return defaultLogger.Sync()
}

// Debug performs the operation.
func Debug(args ...any) {
	if ce := defaultLogger.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Info performs the operation.
func Info(args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Warn performs the operation.
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

// Panic performs the operation.
func Panic(args ...any) {
	if ce := defaultLogger.Check(zap.PanicLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Fatal performs the operation.
func Fatal(args ...any) {
	if ce := defaultLogger.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Printf performs the operation.
func Printf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Debugf performs the operation.
func Debugf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Infof performs the operation.
func Infof(template string, args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Warnf performs the operation.
func Warnf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Errorf performs the operation.
func Errorf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Panicf performs the operation.
func Panicf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.PanicLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Fatalf performs the operation.
func Fatalf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

// Debugw performs the operation.
func Debugw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.DebugLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Infow performs the operation.
func Infow(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.InfoLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Warnw performs the operation.
func Warnw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.WarnLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Errorw performs the operation.
func Errorw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Panicw performs the operation.
func Panicw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.PanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Fatalw performs the operation.
func Fatalw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.FatalLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Log performs the operation.
func Log(lvl zapcore.Level, args ...any) {
	if ce := defaultLogger.Check(lvl, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

// Logf performs the operation.
func Logf(lvl zapcore.Level, msg string, args ...any) {
	if ce := defaultLogger.Check(lvl, ""); ce != nil {
		ce.Message = fmt.Sprintf(msg, args...)
		ce.Write()
	}
}

// Logw performs the operation.
func Logw(lvl zapcore.Level, msg string, fields ...zapcore.Field) {
	if ce := defaultLogger.Check(lvl, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Check returns the result.
func Check(lvl zapcore.Level, args ...any) *zapcore.CheckedEntry {
	ce := defaultLogger.Check(lvl, "")
	if ce != nil {
		ce.Message = sprintln(args...)
	}
	return ce
}

// Println performs the operation.
func Println(args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}
