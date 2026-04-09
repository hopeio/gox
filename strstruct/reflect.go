/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package strstruct

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

func ParseStringSetReflectValue(value reflect.Value, val string, field *reflect.StructField) error {
	if val == "" {
		return nil
	}
	anyV := value.Interface()
	tuV, ok := anyV.(encoding.TextUnmarshaler)
	if !ok {
		tuV, ok = value.Addr().Interface().(encoding.TextUnmarshaler)
	}
	if ok {
		return tuV.UnmarshalText(stringsx.ToBytes(val))
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
		switch anyV.(type) {
		case time.Duration:
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
		switch anyV.(type) {
		case time.Time:
			return setTimeField(val, field, value)
		}
		return setIntField(val, 64, value)
	default:
		return errUnknownType
	}
	return nil
}

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

func setTimeField(val string, structField *reflect.StructField, value reflect.Value) error {
	timeFormat := time.RFC3339
	l := time.Local
	if structField != nil {
		timeFormat = structField.Tag.Get("format")
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

		if val == "" {
			value.Set(reflect.ValueOf(time.Time{}))
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

func setArray(vals []string, value reflect.Value, field *reflect.StructField) error {
	for i, s := range vals {
		err := ParseStringSetReflectValue(value.Index(i), s, field)
		if err != nil {
			return err
		}
	}
	return nil
}

func setSlice(vals []string, value reflect.Value, field *reflect.StructField) error {
	slice := reflect.MakeSlice(value.Type(), len(vals), len(vals))
	err := setArray(vals, slice, field)
	if err != nil {
		return err
	}
	value.Set(slice)
	return nil
}

func setTimeDuration(val string, value reflect.Value) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(d))
	return nil
}
