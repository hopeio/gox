/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gorm

import (
	"context"
	"fmt"
	"strings"
	"time"

	logx "github.com/hopeio/gox/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type Logger struct {
	*zap.Logger
	*logger.Config
}

// NewLogger creates a new instance.
func NewLogger(loger *zap.Logger, conf *logger.Config) logger.Interface {
	if conf == nil {
		conf = &logger.Config{LogLevel: logger.Warn}
	}
	loger = loger.With(zap.String("component", "gorm"))
	return &Logger{Logger: loger, Config: conf}
}

// LogMode returns a clone and raises the logx.Logger level to match gorm LogLevel.
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	lc := *l.Config
	lc.LogLevel = level
	l.Config = &lc
	return &Logger{Logger: l.Logger, Config: &lc}
}

// Info print info
func (l *Logger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel < logger.Info {
		return
	}
	l.Logger.Info(fmt.Sprintf(strings.TrimRight(msg, "\n"), data...), logx.Context(ctx))
}

// Warn print warn messages
func (l *Logger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel < logger.Warn {
		return
	}
	l.Logger.Warn(fmt.Sprintf(strings.TrimRight(msg, "\n"), data...), logx.Context(ctx))
}

// Error print error messages
func (l *Logger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel < logger.Error {
		return
	}
	l.Logger.Error(fmt.Sprintf(strings.TrimRight(msg, "\n"), data...), logx.Context(ctx))
}

// Trace performs the operation.
func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
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
	fields := []zap.Field{elapsedms, sqlField, rowsField, caller, logx.Context(ctx)}
	entry := l.Check(zapcore.Level(4-level), msg)
	entry.Write(fields...)
}
