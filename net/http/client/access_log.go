/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"encoding/json"
	"github.com/hopeio/gox/log"
	"github.com/hopeio/gox/net/http/consts"
	stringsx "github.com/hopeio/gox/strings"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type LogLevel int8

const (
	LogLevelSilent LogLevel = iota
	LogLevelError
	LogLevelInfo
)

type Body struct {
	Data        []byte
	ContentType ContentType
}

func NewBody(data []byte, contentType ContentType) *Body {
	return &Body{Data: data, ContentType: contentType}
}

type AccessLogParam struct {
	Method, Url       string
	Request           *http.Request
	Response          *http.Response
	ReqBody, RespBody []byte
	Duration          time.Duration
}
type AccessLog func(param *AccessLogParam, err error)

func DefaultLogger(param *AccessLogParam, err error) {
	reqField, respField, statusField := zap.Skip(), zap.Skip(), zap.Skip()
	if len(param.ReqBody) > 0 {
		key := "body"
		if strings.HasPrefix(param.Request.Header.Get(consts.HeaderContentType), consts.ContentTypeJson) {
			reqField = zap.Reflect(key, json.RawMessage(param.ReqBody))
		} else {
			reqField = zap.String(key, stringsx.BytesToString(param.ReqBody))
		}
	}
	if len(param.RespBody) > 0 {
		key := "result"
		if len(param.RespBody) > 500 {
			respField = zap.String(key, "<result is too long>")
		} else {
			if strings.HasPrefix(param.Response.Header.Get(consts.HeaderContentType), consts.ContentTypeJson) {
				respField = zap.Reflect(key, json.RawMessage(param.RespBody))
			} else {
				respField = zap.String(key, stringsx.BytesToString(param.RespBody))
			}

		}

	}
	if param.Response != nil {
		statusField = zap.Int("status", param.Response.StatusCode)
	}

	log.NoCallerLogger().Logger.Info("http request", zap.String("url", param.Url),
		zap.String("method", param.Method),
		reqField,
		zap.Duration("duration", param.Duration),
		respField,
		statusField,
		zap.Error(err),
	)
}
