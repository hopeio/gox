/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import (
	"fmt"

	"go.uber.org/zap"
)

// with stack
func StackError(args ...any) {
	if ce := stackLogger.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func StackErrorf(template string, args ...any) {
	if ce := stackLogger.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = fmt.Sprintf(template, args...)
		ce.Write()
	}
}

func StackErrorw(msg string, fields ...zap.Field) {
	if ce := stackLogger.Check(zap.ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// no caller

func NoCallerDebug(args ...any) {
	if ce := noCallerLogger.Check(zap.DebugLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func NoCallerInfo(args ...any) {
	if ce := noCallerLogger.Check(zap.InfoLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func NoCallerWarn(args ...any) {
	if ce := noCallerLogger.Check(zap.WarnLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func NoCallerError(args ...any) {
	if ce := noCallerLogger.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func NoCallerPanic(args ...any) {
	if ce := noCallerLogger.Check(zap.PanicLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func NoCallerFatal(args ...any) {
	if ce := noCallerLogger.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func NoCallerErrorf(template string, args ...any) {
	if ce := noCallerLogger.Check(zap.ErrorLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func NoCallerFatalf(template string, args ...any) {
	if ce := noCallerLogger.Check(zap.FatalLevel, ""); ce != nil {
		ce.Message = sprintln(args...)
		ce.Write()
	}
}

func NoCallerErrorw(msg string, fields ...zap.Field) {
	if ce := noCallerLogger.Check(zap.ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func NoCallerPanicw(msg string, fields ...zap.Field) {
	if ce := noCallerLogger.Check(zap.PanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func NoCallerFatalw(msg string, fields ...zap.Field) {
	if ce := noCallerLogger.Check(zap.FatalLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}
