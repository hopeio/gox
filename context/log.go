/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package context

import (
	"github.com/hopeio/gox/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (c *Context) Log(lvl zapcore.Level, msg string, fields ...zap.Field) {
	log.Logw(lvl, msg, append(fields, zap.String(log.FieldTraceId, c.traceID))...)
}

func (c *Context) ErrorLog(err error, fields ...zap.Field) {
	log.Errorw(err.Error(), append(fields, zap.String(log.FieldTraceId, c.traceID))...)
}

func (c *Context) RespErrorLog(respErr, originErr error, flag string, fields ...zap.Field) error {
	// caller 用原始logger skip刚好
	fields = append(fields, zap.String(log.FieldTraceId, c.traceID),
		zap.String(log.FieldPosition, flag))
	log.CallerSkipLogger(1).Errorw(originErr.Error(), fields...)
	return respErr
}
