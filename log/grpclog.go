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

// grpclog
func (l *Logger) V(level int) bool {
	level -= 2
	return l.Logger.Core().Enabled(zapcore.Level(level))
}

// // 等同于xxxln,为了实现某些接口 如grpclog
func (l *Logger) Infoln(args ...any) {
	if ce := l.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = fmt.Sprint(args...)
		ce.Write()
	}
}

func (l *Logger) Warning(args ...any) {
	if ce := l.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = fmt.Sprint(args...)
		ce.Write()
	}
}

func (l *Logger) Warningln(args ...any) {
	if ce := l.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = fmt.Sprint(args...)
		ce.Write()
	}
}

// grpclog
func (l *Logger) Warningf(template string, args ...any) {
	if ce := l.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

func (l *Logger) Errorln(args ...any) {
	if ce := l.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = fmt.Sprint(args...)
		ce.Write()
	}
}

func (l *Logger) Fatalln(args ...any) {
	if ce := l.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = fmt.Sprint(args...)
		ce.Write()
	}
}

// InfoDepth logs to INFO log at the specified depth. Arguments are handled in the manner of fmt.Println.
func (l *Logger) InfoDepth(depth int, args ...any) {
	if ce := l.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = trimLineBreak(fmt.Sprintln(args...))
		ce.Write()
	}
}

// WarningDepth logs to WARNING log at the specified depth. Arguments are handled in the manner of fmt.Println.
func (l *Logger) WarningDepth(depth int, args ...any) {
	if ce := l.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = trimLineBreak(fmt.Sprintln(args...))
		ce.Write()
	}
}

// ErrorDepth logs to ERROR log at the specified depth. Arguments are handled in the manner of fmt.Println.
func (l *Logger) ErrorDepth(depth int, args ...any) {
	if ce := l.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = trimLineBreak(fmt.Sprintln(args...))
		ce.Write()
	}
}

// FatalDepth logs to FATAL log at the specified depth. Arguments are handled in the manner of fmt.Println.
func (l *Logger) FatalDepth(depth int, args ...any) {
	if ce := l.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = trimLineBreak(fmt.Sprintln(args...))
		ce.Write()
	}
}
