/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gin

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hopeio/gox/kvstruct"
	httpx "github.com/hopeio/gox/net/http"
)

func Bind(ctx *gin.Context, obj any) error {
	return httpx.CommonBind(RequestSource{ctx}, obj)
}

type RequestSource struct {
	*gin.Context
}

func (s RequestSource) Uri() kvstruct.Setter {
	return (uriSource)(s.Params)
}

func (s RequestSource) Query() kvstruct.Setter {
	return (kvstruct.KVsSource)(s.Request.URL.Query())
}

func (s RequestSource) Header() kvstruct.Setter {
	return (httpx.HeaderSource)(s.Request.Header)
}

func (s RequestSource) MultipartForm() kvstruct.Setter {
	contentType := s.Request.Header.Get(httpx.HeaderContentType)
	if strings.HasPrefix(contentType, httpx.ContentTypeMultipart) {
		err := s.Request.ParseMultipartForm(httpx.DefaultMemory)
		if err != nil {
			return nil
		}
		return (*httpx.MultipartSource)(s.Request.MultipartForm)
	}
	return nil
}

func (s RequestSource) BodyBind(obj any) error {
	if s.Request.Method == http.MethodGet {
		return nil
	}
	contentType := s.Request.Header.Get(httpx.HeaderContentType)
	if strings.HasPrefix(contentType, httpx.ContentTypeMultipart) {
		return nil
	}
	data, err := io.ReadAll(s.Request.Body)
	if err != nil {
		return fmt.Errorf("read body error: %w", err)
	}
	if recorder, ok := s.Request.Body.(httpx.RequestRecorder); ok {
		recorder.RecordRequest(contentType, data, obj)
	}
	if len(data) == 0 {
		return nil
	}
	return httpx.DefaultUnmarshal(contentType, data, obj)
}

type uriSource gin.Params

var _ kvstruct.Setter = uriSource(nil)

func (param uriSource) Peek(key string) ([]string, bool) {
	for i := range param {
		if param[i].Key == key {
			return []string{param[i].Value}, true
		}
	}
	return nil, false
}

func (param uriSource) HasValue(key string) bool {
	for i := range param {
		if param[i].Key == key {
			return true
		}
	}
	return false
}

// TrySet tries to set a value by request's form source (like map[string][]string)
func (param uriSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *kvstruct.Options) (isSet bool, err error) {
	return kvstruct.SetValueByKVsWithStructField(value, field, param, key, opt)
}
