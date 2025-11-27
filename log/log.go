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

func init() {
	SetDefaultLogger(&Config{Development: true, Level: zapcore.DebugLevel})
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

func DefaultLogger() *Logger {
	return defaultLogger
}

func SetDefaultLogger(lf *Config, cores ...zapcore.Core) {
	mu.Lock()
	defer mu.Unlock()

	defaultLogger = lf.NewLogger(cores...)
	stackLogger = defaultLogger.WithOptions(zap.WithCaller(true), zap.AddStacktrace(zapcore.ErrorLevel))
	noCallerLogger = defaultLogger.WithOptions(zap.WithCaller(false))
	clf := *lf
	clf.SkipLineEnding = true
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
		skipLoggers[idx].Logger = defaultLogger.AddSkip(skip)
		skipLoggers[idx].needUpdate = false
		mu.Unlock()
	}
	return skipLoggers[idx].Logger
}

func NoCallerLogger() *Logger {
	return noCallerLogger
}
func StackLogger() *Logger {
	return stackLogger
}
func Sync() error {
	return defaultLogger.Sync()
}

func Debug(args ...any) {
	if ce := defaultLogger.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func Info(args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func Warn(args ...any) {
	if ce := defaultLogger.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func Error(args ...any) {
	if ce := defaultLogger.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func Panic(args ...any) {
	if ce := defaultLogger.Check(zap.PanicLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func Fatal(args ...any) {
	if ce := defaultLogger.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func Printf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

func Debugf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

func Infof(template string, args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

func Warnf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

func Errorf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

func Panicf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.PanicLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

func Fatalf(template string, args ...any) {
	if ce := defaultLogger.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

func Debugw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.DebugLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func Infow(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.InfoLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func Warnw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.WarnLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func Errorw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func Panicw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.PanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func Fatalw(msg string, fields ...zap.Field) {
	if ce := defaultLogger.Check(zap.FatalLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func Log(lvl zapcore.Level, args ...any) {
	if ce := defaultLogger.Check(lvl, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func Logf(lvl zapcore.Level, msg string, args ...any) {
	if ce := defaultLogger.Check(lvl, ""); ce != nil {
		ce.Message = fmt.Sprintf(msg, args...)
		ce.Write()
	}
}

func Logw(lvl zapcore.Level, msg string, fields ...zapcore.Field) {
	if ce := defaultLogger.Check(lvl, msg); ce != nil {
		ce.Write(fields...)
	}
}

func Check(lvl zapcore.Level, args ...any) *zapcore.CheckedEntry {
	ce := defaultLogger.Check(lvl, "")
	if ce != nil {
		ce.Message = sprintln(args...)
	}
	return ce
}

func Println(args ...any) {
	if ce := defaultLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}
