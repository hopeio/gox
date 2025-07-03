/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package dingding

import (
	"github.com/hopeio/utils/sdk/dingtalk"
	"go.uber.org/zap/zapcore"
)

func NewCore(token, secret string, level zapcore.Level, encoderConfig *zapcore.EncoderConfig) zapcore.Core {
	return &core{
		RobotConfig: dingtalk.RobotConfig{
			Token:  token,
			Secret: secret,
		},
		encoder: newDingEncoder(encoderConfig),
		Level:   level,
	}
}

type core struct {
	dingtalk.RobotConfig
	zapcore.Level
	encoder zapcore.Encoder
}

func (c *core) With(fields []zapcore.Field) zapcore.Core {
	for i := range fields {
		fields[i].AddTo(c.encoder)
	}
	return c
}
func (c *core) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}
func (c *core) Write(e zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.encoder.EncodeEntry(e, fields)
	if err != nil {
		return err
	}
	return dingtalk.RobotSendMarkDownMessageWithSecret(c.Token, c.Secret, "日志", buf.String(), nil)
}
func (*core) Sync() error { return nil }
