/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package dingcore

import (
	"testing"

	"github.com/hopeio/gox/log"
	"go.uber.org/zap/zapcore"
)

func TestDingDing(t *testing.T) {

	log.SetDefaultLogger(&log.Config{
		Development: false,
		Level:       zapcore.DebugLevel,
		OutputPaths: log.OutPutPaths{},
		Name:        "",
	}, NewCore("", "", zapcore.DebugLevel, &zapcore.EncoderConfig{}))
	log.Info("测试")
}
