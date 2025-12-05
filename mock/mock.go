/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package mock

import (
	"math/rand/v2"
	"reflect"
	"time"

	rand2 "github.com/hopeio/gox/math/rand"
)

var timeType = reflect.TypeOf(time.Time{})

func New[T any]() T {
	var v T
	Mock(&v)
	return v
}

func Mock(v interface{}) {
	value := reflect.ValueOf(v)
	typMap := make(map[reflect.Type]uint8)
	mock(value, nil, typMap)
}

// 数组长度
const length = 1

// 一个类型最大重复次数
const times = 3

func mock(value reflect.Value, field *reflect.StructField, typMap map[reflect.Type]uint8) {
	typ := value.Type()
	var tag string
	if field != nil {
		tag = field.Tag.Get("mock")
	}
	switch value.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if tag == "" {
			value.SetUint(rand.Uint64N(256))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if tag == "" {
			value.SetInt(rand.Int64N(128))
		}
	case reflect.Float32, reflect.Float64:
		if tag == "" {
			value.SetFloat(rand.ExpFloat64())
		}
	case reflect.String:
		if tag == "" {
			value.SetString(rand2.English())
		}
	case reflect.Ptr:
		if value.IsNil() && value.CanSet() {
			value.Set(reflect.New(typ.Elem()))
		}
		mock(value.Elem(), field, typMap)
	case reflect.Struct:
		if typ == timeType {
			value.Set(reflect.ValueOf(time.Now()))
			return
		}
		if count := typMap[typ]; count == times {
			return
		}
		typMap[typ] = typMap[typ] + 1
		for i := 0; i < value.NumField(); i++ {
			sf := typ.Field(i)
			if !sf.IsExported() {
				continue
			}
			fieldValue := value.Field(i)
			mock(fieldValue, &sf, typMap)
		}
	case reflect.Array:
		for i := 0; i < value.Len(); i++ {
			mock(value.Index(i), field, typMap)
		}
	case reflect.Slice:
		value.Set(reflect.MakeSlice(typ, length, length))
		for i := 0; i < length; i++ {
			mock(value.Index(i), field, typMap)
		}
	case reflect.Map:
		value.Set(reflect.MakeMapWithSize(typ, length))
		for i := 0; i < length; i++ {
			mk := reflect.New(typ.Key()).Elem()
			mock(mk, field, typMap)
			mv := reflect.New(typ.Elem()).Elem()
			mock(mv, field, typMap)
			value.SetMapIndex(mk, mv)
		}
	}
}
