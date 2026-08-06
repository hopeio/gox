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
	*logx.Logger
	*logger.Config
}

// NewLogger creates a new instance.
func NewLogger(logcfg *logx.Config, conf *logger.Config) logger.Interface {
	if conf == nil {
		conf = &logger.Config{LogLevel: logger.Warn}
	}
	loger := logcfg.NewLogger()
	loger = loger.With(zap.String("component", "gorm"))
	return &Logger{Logger: loger, Config: conf}
}

// LogMode returns a clone with the given gorm log level (same as gorm default logger).
// Zap cannot safely remap level here: IncreaseLevel only raises the threshold and
// WithOptions must not mutate the shared parent logger.
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	nl := *l
	cfg := *l.Config
	cfg.LogLevel = level
	nl.Config = &cfg
	return &nl
}

// Info print info
func (l *Logger) Info(ctx context.Context, msg string, data ...any) {
	l.Logger.Info(fmt.Sprintf(strings.TrimRight(msg, "\n"), data...), logx.Context(ctx))
}

// Warn print warn messages
func (l *Logger) Warn(ctx context.Context, msg string, data ...any) {
	l.Logger.Warn(fmt.Sprintf(strings.TrimRight(msg, "\n"), data...), logx.Context(ctx))
}

// Error print error messages
func (l *Logger) Error(ctx context.Context, msg string, data ...any) {
	l.Logger.Error(fmt.Sprintf(strings.TrimRight(msg, "\n"), data...), logx.Context(ctx))
}

// Trace performs the operation.
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
	fields := []zap.Field{elapsedms, sqlField, rowsField, caller, logx.Context(ctx)}
	entry := l.Check(zapcore.Level(4-level), msg)
	entry.Write(fields...)
}
