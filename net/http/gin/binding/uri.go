/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package binding

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/hopeio/gox/mtos"
)

type uriSource gin.Params

var _ mtos.Setter = uriSource(nil)

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
func (param uriSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *mtos.Options) (isSet bool, err error) {
	return mtos.SetValueByKVsWithStructField(value, field, param, key, opt)
}
