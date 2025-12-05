// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encoding

import (
	"encoding"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/exp/constraints"
)

type StringConverter func(string) any
type StringConverterE func(string) (any, error)

func (c StringConverterE) IgnoreError() StringConverter {
	if c == nil {
		return nil
	}
	return func(value string) any {
		r, _ := c(value)
		return r
	}
}

var (
	invalidValue = reflect.Value{}
)

// Default converters for basic types.
/*var stringConverterMaps = map[reflect.Kind]StringConverterE{
	reflect.Bool:    stringConvertBool,
	reflect.Float32: stringConvertFloat32,
	reflect.Float64: stringConvertFloat64,
	reflect.Int:     stringConvertInt,
	reflect.Int8:    stringConvertInt8,
	reflect.Int16:   stringConvertInt16,
	reflect.Int32:   stringConvertInt32,
	reflect.Int64:   stringConvertInt64,
	reflect.String:  stringConvertString,
	reflect.Uint:    stringConvertUint,
	reflect.Uint8:   stringConvertUint8,
	reflect.Uint16:  stringConvertUint16,
	reflect.Uint32:  stringConvertUint32,
	reflect.Uint64:  stringConvertUint64,
}*/

// Deprecated: unsupported string slices array map, use mtos.SetValueByString instead
var stringConverterArrays = [...]StringConverterE{
	reflect.Invalid: nil,
	reflect.Bool:    stringConvertBool,
	reflect.Int:     stringConvertInt,
	reflect.Int8:    stringConvertInt8,
	reflect.Int16:   stringConvertInt16,
	reflect.Int32:   stringConvertInt32,
	reflect.Int64:   stringConvertInt64,
	reflect.Uint:    stringConvertUint,
	reflect.Uint8:   stringConvertUint8,
	reflect.Uint16:  stringConvertUint16,
	reflect.Uint32:  stringConvertUint32,
	reflect.Uint64:  stringConvertUint64,
	reflect.Uintptr: stringConvertUint,
	reflect.Float32: stringConvertFloat32,
	reflect.Float64: stringConvertFloat64,
}

func GetStringConverter(typ reflect.Type) StringConverter {
	return GetStringConverterE(typ).IgnoreError()
}

func GetStringConverterE(typ reflect.Type) StringConverterE {
	kind := typ.Kind()
	if kind == reflect.Slice || kind == reflect.Array {
		return getStringConvertArray(typ.Elem())
	}
	return GetStringConverterEByKind(kind)
}

func GetStringConverterByKind(kind reflect.Kind) StringConverter {
	return GetStringConverterEByKind(kind).IgnoreError()
}

func GetStringConverterEByKind(kind reflect.Kind) StringConverterE {
	if kind == reflect.String {
		return stringConvertString
	}
	if kind > reflect.Uint64 {
		return nil
	}
	return stringConverterArrays[kind]
}

func stringConvertBool(value string) (any, error) {
	if value == "on" {
		return true, nil
	}
	return strconv.ParseBool(value)
}

func StringConvertBool(value string) (bool, error) {
	if value == "on" {
		return true, nil
	}
	return strconv.ParseBool(value)
}

func stringConvertFloat32(value string) (any, error) {
	f, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

func StringConvertFloat32(value string) (float32, error) {
	f, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

func stringConvertFloat64(value string) (any, error) {
	return strconv.ParseFloat(value, 64)
}

func StringConvertFloat64(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

func stringConvertInt(value string) (any, error) {
	i, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

func stringConvertInt8(value string) (any, error) {
	i, err := strconv.ParseInt(value, 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(i), nil
}

func stringConvertInt16(value string) (any, error) {
	i, err := strconv.ParseInt(value, 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(i), nil
}

func stringConvertInt32(value string) (any, error) {
	i, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

func stringConvertInt64(value string) (any, error) {
	return strconv.ParseInt(value, 10, 64)
}

func stringConvertString(value string) (any, error) {
	return value, nil
}

func StringConvertIntFor[T constraints.Signed](value string) (T, error) {
	i, err := strconv.ParseInt(value, 10, int(unsafe.Sizeof(T(0))*8))
	if err != nil {
		return 0, err
	}
	return T(i), nil
}

func stringConvertArrayFor[T any](value string) (any, error) {
	strs := strings.Split(value, ",")
	var rets []any
	kind := reflect.TypeFor[T]().Kind()
	if kind == reflect.String {
		return strs, nil
	}
	converter := stringConverterArrays[kind]
	if converter != nil {
		for i := range strs {
			v, err := converter(strs[i])
			if err != nil {
				return nil, err
			}
			rets = append(rets, v)
		}
	}
	return rets, nil
}

func getStringConvertArray(elemTyp reflect.Type) func(value string) (any, error) {
	return func(value string) (any, error) {
		strs := strings.Split(value, ",")
		rets := make([]any, 0, len(strs))
		kind := elemTyp.Kind()
		if kind == reflect.String {
			return strs, nil
		}
		converter := stringConverterArrays[kind]
		if converter != nil {
			for i := range strs {
				v, err := converter(strs[i])
				if err != nil {
					return nil, err
				}
				rets = append(rets, reflect.ValueOf(v).Convert(elemTyp).Interface())
			}
		}
		return rets, nil
	}
}

func stringConvertUint(value string) (any, error) {
	u, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(u), nil
}

func stringConvertUint8(value string) (any, error) {
	u, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(u), nil
}

func stringConvertUint16(value string) (any, error) {
	u, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(u), nil
}

func stringConvertUint32(value string) (any, error) {
	u, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(u), nil
}

func stringConvertUint64(value string) (any, error) {
	return strconv.ParseUint(value, 10, 64)
}

func StringConvertUintFor[T constraints.Unsigned](value string) (T, error) {
	i, err := strconv.ParseInt(value, 10, int(unsafe.Sizeof(T(0))*8))
	if err != nil {
		return 0, err
	}
	return T(i), nil
}

func StringConvert(kind reflect.Kind, value string) (any, error) {
	if kind == reflect.String {
		return value, nil
	}
	converter := stringConverterArrays[kind]
	if converter != nil {
		return converter(value)
	}
	return nil, errors.New("unsupported kind")
}

func StringConvertFloatFor[T constraints.Float](value string) (T, error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return T(f), nil
}

func String(value reflect.Value) string {
	v := value.Interface()
	if t, ok := v.(encoding.TextMarshaler); ok {
		s, _ := t.MarshalText()
		return string(s)
	}

	kind := value.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64, reflect.Pointer, reflect.UnsafePointer:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Float64, reflect.Float32:
		return strconv.FormatFloat(value.Float(), 'g', -1, 64)
	case reflect.Array, reflect.Slice:
		var strs []string
		for i := 0; i < value.Len(); i++ {
			strs = append(strs, String(value.Index(i)))
		}
		return strings.Join(strs, ",")
	}
	return ""
}

func StringFor[T any](t T) string {
	v := reflect.ValueOf(t)
	return String(v)
}
