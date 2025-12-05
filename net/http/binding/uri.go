/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package binding

import (
	"net/http"
	"reflect"

	"github.com/hopeio/gox/mtos"
)

type UriSource http.Request

var _ mtos.Setter = (*UriSource)(nil)

func (req *UriSource) Peek(key string) ([]string, bool) {
	if req.Pattern == "" {
		return nil, false
	}
	v := (*http.Request)(req).PathValue(key)
	return []string{v}, v != ""
}

func (req *UriSource) HasValue(key string) bool {
	v := (*http.Request)(req).PathValue(key)
	return v != ""
}

// TrySet tries to set a value by request's form source (like map[string][]string)
func (req *UriSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *mtos.Options) (isSet bool, err error) {
	return mtos.SetValueByKVsWithStructField(value, field, req, key, opt)
}
