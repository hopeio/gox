/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package binding

import (
	"net/textproto"
	"reflect"

	"github.com/hopeio/gox/mtos"
)

type HeaderSource map[string][]string

var _ mtos.Setter = HeaderSource(nil)

func (hs HeaderSource) Peek(key string) ([]string, bool) {
	v, ok := hs[textproto.CanonicalMIMEHeaderKey(key)]
	return v, ok
}

func (hs HeaderSource) HasValue(key string) bool {
	_, ok := hs[textproto.CanonicalMIMEHeaderKey(key)]
	return ok
}
func (hs HeaderSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *mtos.Options) (isSet bool, err error) {
	return mtos.SetValueByKVsWithStructField(value, field, hs, key, opt)
}
