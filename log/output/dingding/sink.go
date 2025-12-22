/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package dingding

import (
	"net/url"

	"github.com/hopeio/gox/sdk/dingtalk"
	"go.uber.org/zap"
)

type sink dingtalk.RobotConfig

// TODO
func (th *sink) Write(b []byte) (n int, err error) {
	return
}

func (th *sink) Sync() error {
	return nil
}

func (th *sink) Close() error {
	return nil
}

// dingding://${token}?sercret=${sercret}
func RegisterSink() {
	_ = zap.RegisterSink("dingding", func(url *url.URL) (sinkv zap.Sink, e error) {
		th := new(sink)
		return th, nil
	})
}

func NewSink(token, secret string) zap.Sink {
	th := new(sink)
	th.Token = token
	th.Secret = secret
	return th
}
