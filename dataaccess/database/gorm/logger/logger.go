/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package logger

import (
	"context"
	"fmt"
	logi "github.com/hopeio/gox/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"strings"
	"time"
)

type Logger struct {
	*zap.Logger
	*logger.Config
}

type Config struct {
	SlowThreshold time.Duration
	Colorful      bool
	LogLevel      zapcore.Level
}

func New(loger *zap.Logger, conf *logger.Config) logger.Interface {
	if conf == nil {
		conf = &logger.Config{LogLevel: logger.Warn}
	}
	loger.Core().With([]zap.Field{zap.String("comment", "gorm")}).Enabled(zapcore.Level(4 - conf.LogLevel))
	return &Logger{Logger: loger, Config: conf}
}

// LogMode log mode
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	l.Logger.Core().Enabled(zapcore.Level(4 - level))
	l.LogLevel = level
	return l
}

// Info print info
func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Info(fmt.Sprintf(strings.TrimRight(msg, "\n"), data...), zap.String(logi.FieldTraceId, TraceId(ctx)))
}

// Warn print warn messages
func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Warn(fmt.Sprintf(strings.TrimRight(msg, "\n"), data...), zap.String(logi.FieldTraceId, TraceId(ctx)))
}

// Error print error messages
func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Error(fmt.Sprintf(strings.TrimRight(msg, "\n"), data...), zap.String(logi.FieldTraceId, TraceId(ctx)))
}

// Trace print sql message 只有这里的context不是background,看了代码,也没用
func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel == logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	elapsedms := zap.Float64("elapsedms", float64(elapsed.Nanoseconds())/1e6)
	level := logger.Info
	var msg string
	switch {
	case err != nil:
		level = logger.Error
		msg = err.Error()
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		level = logger.Warn
		msg = fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
	}
	if l.LogLevel < level {
		return
	}
	switch level {
	case logger.Error:
		msg = err.Error()
	case logger.Warn:
		msg = fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
	}
	sql, rows := fc()
	sqlField := zap.String("sql", sql)
	rowsField := zap.Int64("rows", rows)
	caller := zap.String("caller", utils.FileWithLineNum())
	fields := []zap.Field{elapsedms, sqlField, rowsField, caller, zap.String(logi.FieldTraceId, TraceId(ctx))}
	entry := l.Check(zapcore.Level(4-level), msg)
	// entry.Caller = zapcore.NewEntryCaller(0, "", 0, false) utils.FileWithLineNum() or 获取到gorm的gormSourceDir
	entry.Write(fields...)
}

func TraceId(ctx context.Context) string {
	if traceId, ok := ctx.Value(traceIdKey{}).(string); ok {
		return traceId
	}
	return "unknown"
}

type traceIdKey struct{}

func SetTranceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, traceIdKey{}, traceId)
}
