/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package kvstruct

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	stringsx "github.com/hopeio/gox/strings"
)

var (
	errUnknownType  = errors.New("unknown type")
	errUnknownField = errors.New("unknown field")
)

// ParseStringSetReflectValue parses the input.
func ParseStringSetReflectValue(value reflect.Value, val string, field *reflect.StructField) error {
	if val == "" {
		return nil
	}
	if !value.CanInterface() {
		return errUnknownField
	}
	// 优先经指针断言 TextUnmarshaler：指针方法集含值方法，且修改能落到原值上；
	// 值断言只对指针类型的 value 有意义（值接收者实现改的是装箱副本，改动会丢失）
	if value.CanAddr() {
		if tuV, ok := value.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return tuV.UnmarshalText(stringsx.ToBytes(val))
		}
	} else if value.Kind() == reflect.Pointer && !value.IsNil() {
		if tuV, ok := value.Interface().(encoding.TextUnmarshaler); ok {
			return tuV.UnmarshalText(stringsx.ToBytes(val))
		}
	}
	switch kind := value.Kind(); kind {
	case reflect.Int:
		return setIntField(val, 0, value)
	case reflect.Int8:
		return setIntField(val, 8, value)
	case reflect.Int16:
		return setIntField(val, 16, value)
	case reflect.Int32:
		return setIntField(val, 32, value)
	case reflect.Int64:
		if _, ok := value.Interface().(time.Duration); ok {
			return setTimeDuration(val, value)
		}
		return setIntField(val, 64, value)
	case reflect.Uint:
		return setUintField(val, 0, value)
	case reflect.Uint8:
		return setUintField(val, 8, value)
	case reflect.Uint16:
		return setUintField(val, 16, value)
	case reflect.Uint32:
		return setUintField(val, 32, value)
	case reflect.Uint64:
		return setUintField(val, 64, value)
	case reflect.Bool:
		return setBoolField(val, value)
	case reflect.Float32:
		return setFloatField(val, 32, value)
	case reflect.Float64:
		return setFloatField(val, 64, value)
	case reflect.String:
		value.SetString(val)
	case reflect.Array, reflect.Slice:
		typ := value.Type()
		subType := typ.Elem()
		eKind := subType.Kind()
		if eKind == reflect.Array || eKind == reflect.Slice || eKind == reflect.Map {
			return fmt.Errorf("unsupported sub type %v", subType)
		}
		strs := strings.Split(val, ",")
		if kind == reflect.Array {
			if len(strs) != value.Len() {
				return fmt.Errorf("%q is not valid value for %s", strs, value.Type().String())
			}
		}
		if kind == reflect.Slice {
			value.Set(reflect.MakeSlice(typ, len(strs), len(strs)))
		}
		for i := 0; i < value.Len(); i++ {
			if err := ParseStringSetReflectValue(value.Index(i), strs[i], nil); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		if _, ok := value.Interface().(time.Time); ok {
			return setTimeField(val, field, value)
		}
		// SetInt 对 struct Value 会 panic，未知结构体直接报不支持
		return errUnknownType
	default:
		return errUnknownType
	}
	return nil
}

// ParseStringsSetReflectValue parses the input.
func ParseStringsSetReflectValue(value reflect.Value, vals []string, field *reflect.StructField) error {
	if len(vals) == 0 {
		return nil
	}
	switch value.Kind() {
	case reflect.Slice:
		return setSlice(vals, value, field)
	case reflect.Array:
		if len(vals) != value.Len() {
			return fmt.Errorf("%q is not valid value for %s", vals, value.Type().String())
		}
		return setArray(vals, value, field)
	default:
		return ParseStringSetReflectValue(value, vals[0], field)
	}
}

// setIntField performs the operation.
func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		return nil
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

// setUintField performs the operation.
func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		return nil
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

// setBoolField performs the operation.
func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		return nil
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

// setFloatField performs the operation.
func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		return nil
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

// setTimeField performs the operation.
func setTimeField(val string, structField *reflect.StructField, value reflect.Value) error {
	timeFormat := time.RFC3339
	l := time.Local
	if structField != nil {
		// tag 缺失时保持 RFC3339 默认值，不能被空串覆盖
		if tf := structField.Tag.Get("format"); tf != "" {
			timeFormat = tf
		}
		switch tf := strings.ToLower(timeFormat); tf {
		case "unix", "unixnano":
			tv, err := strconv.ParseInt(val, 10, 0)
			if err != nil {
				return err
			}

			d := time.Duration(1)
			if tf == "unixnano" {
				d = time.Second
			}

			t := time.Unix(tv/int64(d), tv%int64(d))
			value.Set(reflect.ValueOf(t))
			return nil

		}

		if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
			l = time.UTC
		}

		if locTag := structField.Tag.Get("time_location"); locTag != "" {
			loc, err := time.LoadLocation(locTag)
			if err != nil {
				return err
			}
			l = loc
		}
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(t))
	return nil
}

// setArray performs the operation.
func setArray(vals []string, value reflect.Value, field *reflect.StructField) error {
	for i, s := range vals {
		err := ParseStringSetReflectValue(value.Index(i), s, field)
		if err != nil {
			return err
		}
	}
	return nil
}

// setSlice performs the operation.
func setSlice(vals []string, value reflect.Value, field *reflect.StructField) error {
	slice := reflect.MakeSlice(value.Type(), len(vals), len(vals))
	err := setArray(vals, slice, field)
	if err != nil {
		return err
	}
	value.Set(slice)
	return nil
}

// setTimeDuration performs the operation.
func setTimeDuration(val string, value reflect.Value) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(d))
	return nil
}
