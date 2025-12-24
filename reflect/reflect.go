/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package reflect

import (
	"fmt"
	"reflect"
)

func SetValue(fieldValue reflect.Value, value any) error {
	if !fieldValue.IsValid() {
		return fmt.Errorf("field value invalid")
	}
	if !fieldValue.CanSet() {
		return fmt.Errorf("cannot set field value")
	}

	fieldType := fieldValue.Type()
	val := reflect.ValueOf(value)

	valTypeKind := val.Type().Kind()
	fieldTypeKind := fieldType.Kind()
	if fieldType != val.Type() && val.CanConvert(fieldType) {
		val = val.Convert(fieldType)
	} else {
		return fmt.Errorf("provided value type %s didn't match obj field type %s", valTypeKind, fieldTypeKind)
	}
	fieldValue.Set(val)
	return nil
}

func CanCast(t1, t2 reflect.Type, strict bool) bool {
	t1kind, t2kind := t1.Kind(), t2.Kind()
	if strict {
		if (t1 == t2) || (t1kind == reflect.String && t2kind == reflect.String) || (t1kind <= reflect.Complex128 && t2kind <= reflect.Complex128 && t1kind == t2kind) {
			return true
		}
	} else {
		if t1kind <= reflect.Float64 && t2kind <= reflect.Float64 {
			return true
		}
		if (t1kind == reflect.Complex64 || t1kind == reflect.Complex128) && (t2kind == reflect.Complex64 || t2kind == reflect.Complex128) {
			return true
		}
	}

	switch t1kind {
	case reflect.String:
		return t1kind == t2kind
	case reflect.Ptr, reflect.Array, reflect.Chan, reflect.Slice, reflect.Map:
		if t1kind == reflect.Map {
			if !CanCast(t1.Key(), t2.Key(), true) {
				return false
			}
		}
		if t1kind == reflect.Array && t1.Len() != t2.Len() {
			return false
		}
		return CanCast(t1.Elem(), t2.Elem(), true)
	case reflect.Struct:
		if t1.NumField() != t2.NumField() {
			return false
		}
		for i := range t1.NumField() {
			if !CanCast(t1.Field(i).Type, t2.Field(i).Type, true) {
				return false
			}
		}
	case reflect.Interface, reflect.UnsafePointer:
		return t1 == t2
	}
	return true
}

func InitValue(v reflect.Value) {
	v = InitPtr(v)
	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			field := v.Field(i)
			InitValue(field)
		}
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			InitValue(v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			InitValue(v.MapIndex(key))
		}
	case reflect.Interface:
		v = v.Elem()
		if v.IsValid() {
			InitValue(v)
		}
	}
}
