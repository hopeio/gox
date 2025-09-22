/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package binding

import (
	"github.com/hopeio/gox/reflect/mtos"
	stringsx "github.com/hopeio/gox/strings"
	"github.com/valyala/fasthttp"
	"reflect"
)

type ArgsSource fasthttp.Args

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form *ArgsSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt mtos.SetOptions) (isSet bool, err error) {
	return mtos.SetValueByKVsWithStructField(value, field, form, key, opt)
}

func (form *ArgsSource) Peek(key string) ([]string, bool) {
	v := stringsx.BytesToString((*fasthttp.Args)(form).Peek(key))
	return []string{v}, v != ""
}

func (form *ArgsSource) HasValue(key string) bool {
	v := stringsx.BytesToString((*fasthttp.Args)(form).Peek(key))
	return v != ""
}

type CtxSource fasthttp.RequestCtx

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form *CtxSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt mtos.SetOptions) (isSet bool, err error) {
	return mtos.SetValueByKVsWithStructField(value, field, form, key, opt)
}

func (form *CtxSource) Peek(key string) ([]string, bool) {
	v := (*fasthttp.RequestCtx)(form).UserValue(key).(string)
	return []string{v}, v != ""
}

type HeaderSource fasthttp.RequestHeader

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form *HeaderSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt mtos.SetOptions) (isSet bool, err error) {
	return mtos.SetValueByKVsWithStructField(value, field, form, key, opt)
}

func (form *HeaderSource) Peek(key string) ([]string, bool) {
	v := stringsx.BytesToString((*fasthttp.RequestHeader)(form).Peek(key))
	return []string{v}, v != ""
}
