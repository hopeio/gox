/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gin

import (
	"context"
	"io"
	"net/http"

	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/hopeio/gox/strstruct"
	httpx "github.com/hopeio/gox/net/http"
)

func Bind(ctx *gin.Context, obj any) error {
	return httpx.CommonBind(RequestSource{ctx}, obj)
}

type RequestSource struct {
	*gin.Context
}

func (s RequestSource) Uri() strstruct.Getter {
	return s.Params
}

func (s RequestSource) Query() strstruct.ValuesGetter {
	return (strstruct.KVsSource)(s.Request.URL.Query())
}

func (s RequestSource) Header() strstruct.ValuesGetter {
	return (httpx.HeaderSource)(s.Request.Header)
}

func (s RequestSource) Body() (context.Context, string, io.ReadCloser) {
	if s.Request.Method == http.MethodGet {
		return s, "", nil
	}
	return s, s.Request.Header.Get(httpx.HeaderContentType), s.Request.Body
}

type uriSource gin.Params

var _ strstruct.Setter = uriSource(nil)

func (param uriSource) Get(key string) ([]string, bool) {
	for i := range param {
		if param[i].Key == key {
			return []string{param[i].Value}, true
		}
	}
	return nil, false
}

// TrySet tries to set a value by request's form source (like map[string][]string)
func (param uriSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *strstruct.Options) (isSet bool, err error) {
	return strstruct.SetValueByValuesGetter(value, field, param, key, opt)
}
