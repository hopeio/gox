// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

import (
	"encoding"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/exp/constraints"
)

// ParseFor ...
func ParseFor[T any](str string) (T, error) {
	var t T
	a, ap := any(t), any(&t)
	itv, ok := a.(encoding.TextUnmarshaler)
	if !ok {
		itv, ok = ap.(encoding.TextUnmarshaler)
	}
	if ok {
		err := itv.UnmarshalText(ToBytes(str))
		if err != nil {
			return t, err
		}
		return t, nil
	}

	if err := setReflectValueFromString(reflect.ValueOf(&t).Elem(), str); err != nil {
		return t, err
	}
	return t, nil
}

// setReflectValueFromString ...
func setReflectValueFromString(v reflect.Value, str string) error {
	if !v.IsValid() || !v.CanSet() {
		return fmt.Errorf("cannot set value of kind %v", v.Kind())
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString(str)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		v.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(str)
			if err != nil {
				return err
			}
			v.SetInt(int64(d))
			return nil
		}
		i, err := strconv.ParseInt(str, 0, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := strconv.ParseUint(str, 0, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetUint(u)
		return nil
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(str, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetFloat(f)
		return nil
	case reflect.Pointer:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return setReflectValueFromString(v.Elem(), str)
	default:
		return fmt.Errorf("unsupported type %s", v.Type())
	}
}

// ParsePtrFor ...
func ParsePtrFor[T any](str string) (*T, error) {
	return toPtr(ParseFor[T](str))
}

// Signed ...
func Signed[T constraints.Signed](value string) (T, error) {
	i, err := strconv.ParseInt(value, 10, int(unsafe.Sizeof(T(0))*8))
	if err != nil {
		return 0, err
	}
	return T(i), nil
}

// SignedP ...
func SignedP[T constraints.Signed](value string) (*T, error) {
	return toPtr(Signed[T](value))
}

// SignedSlice converts 'val' where individual integers are separated by
// 'sep' into a Signed slice.
func SignedSlice[T constraints.Signed](val, sep string) ([]T, error) {
	s := strings.Split(val, sep)
	values := make([]T, len(s))
	for i, v := range s {
		value, err := Signed[T](v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// UnSigned ...
func UnSigned[T constraints.Unsigned](value string) (T, error) {
	i, err := strconv.ParseUint(value, 10, int(unsafe.Sizeof(T(0))*8))
	if err != nil {
		return 0, err
	}
	return T(i), nil
}

// UnSignedP ...
func UnSignedP[T constraints.Unsigned](value string) (*T, error) {
	return toPtr(UnSigned[T](value))
}

// UnsignedSlice converts 'val' where individual integers are separated by
// 'sep' into a Unsigned slice.
func UnsignedSlice[T constraints.Unsigned](val, sep string) ([]T, error) {
	s := strings.Split(val, sep)
	values := make([]T, len(s))
	for i, v := range s {
		value, err := UnSigned[T](v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Float ...
func Float[T constraints.Float](value string) (T, error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return T(f), nil
}

// FloatP ...
func FloatP[T constraints.Float](value string) (*T, error) {
	return toPtr(Float[T](value))
}

// FloatSlice converts 'val' where individual floating point numbers are separated by
// 'sep' into a float slice.
func FloatSlice[T constraints.Float](val, sep string) ([]T, error) {
	s := strings.Split(val, sep)
	values := make([]T, len(s))
	for i, v := range s {
		value, err := Float[T](v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// String returns the string representation.
func String(val string) (string, error) {
	return val, nil
}

// StringSlice converts 'val' where individual strings are separated by
// 'sep' into a string slice.
func StringSlice(val, sep string) ([]string, error) {
	return strings.Split(val, sep), nil
}

// Bool ...
func Bool(value string) (bool, error) {
	return strconv.ParseBool(value)
}

// BoolP ...
func BoolP(value string) (*bool, error) {
	return toPtr(Bool(value))
}

// BoolSlice converts 'val' where individual booleans are separated by
// 'sep' into a bool slice.
func BoolSlice(val, sep string) ([]bool, error) {
	s := strings.Split(val, sep)
	values := make([]bool, len(s))
	for i, v := range s {
		value, err := Bool(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Bytes converts the given string representation of a byte sequence into a slice of bytes
// A bytes sequence is encoded in URL-safe base64 without padding
func Bytes(val string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		b, err = base64.URLEncoding.DecodeString(val)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}

//	BytesSlice converts 'val' where individual bytes sequences, encoded in URL-safe
//
// base64 without padding, are separated by 'sep' into a slice of bytes slices slice.
func BytesSlice(val, sep string) ([][]byte, error) {
	s := strings.Split(val, sep)
	values := make([][]byte, len(s))
	for i, v := range s {
		value, err := Bytes(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Float64 converts the given string representation into representation of a floating point number into float64.
func Float64(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}

// Float64P ...
func Float64P(val string) (*float64, error) {
	return toPtr(Float64(val))
}

// Float64Slice converts 'val' where individual floating point numbers are separated by
// 'sep' into a float64 slice.
func Float64Slice(val, sep string) ([]float64, error) {
	s := strings.Split(val, sep)
	values := make([]float64, len(s))
	for i, v := range s {
		value, err := Float64(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Float32 converts the given string representation of a floating point number into float32.
func Float32(val string) (float32, error) {
	f, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

// Float32P ...
func Float32P(val string) (*float32, error) {
	return toPtr(Float32(val))
}

// Float32Slice converts 'val' where individual floating point numbers are separated by
// 'sep' into a float32 slice.
func Float32Slice(val, sep string) ([]float32, error) {
	s := strings.Split(val, sep)
	values := make([]float32, len(s))
	for i, v := range s {
		value, err := Float32(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Int ...
func Int(val string) (int, error) {
	i, err := strconv.ParseInt(val, 0, 0)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

// IntP ...
func IntP(val string) (*int, error) {
	return toPtr(Int(val))
}

// IntSlice converts 'val' where individual integers are separated by
// 'sep' into a int slice.
func IntSlice(val, sep string) ([]int, error) {
	s := strings.Split(val, sep)
	values := make([]int, len(s))
	for i, v := range s {
		value, err := Int(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Int8 ...
func Int8(val string) (int8, error) {
	i, err := strconv.ParseInt(val, 0, 8)
	if err != nil {
		return 0, err
	}
	return int8(i), nil
}

// Int8P ...
func Int8P(val string) (*int8, error) {
	return toPtr(Int8(val))
}

// Int8Slice converts 'val' where individual integers are separated by
// 'sep' into a int8 slice.
func Int8Slice(val, sep string) ([]int8, error) {
	s := strings.Split(val, sep)
	values := make([]int8, len(s))
	for i, v := range s {
		value, err := Int8(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Int16 ...
func Int16(val string) (int16, error) {
	i, err := strconv.ParseInt(val, 0, 16)
	if err != nil {
		return 0, err
	}
	return int16(i), nil
}

// Int16P ...
func Int16P(val string) (*int16, error) {
	return toPtr(Int16(val))
}

// Int16Slice converts 'val' where individual integers are separated by
// 'sep' into a int slice.
func Int16Slice(val, sep string) ([]int16, error) {
	s := strings.Split(val, sep)
	values := make([]int16, len(s))
	for i, v := range s {
		value, err := Int16(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Int32 converts the given string representation of an integer into int32.
func Int32(val string) (int32, error) {
	i, err := strconv.ParseInt(val, 0, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

// Int32P ...
func Int32P(val string) (*int32, error) {
	return toPtr(Int32(val))
}

// Int32Slice converts 'val' where individual integers are separated by
// 'sep' into a int32 slice.
func Int32Slice(val, sep string) ([]int32, error) {
	s := strings.Split(val, sep)
	values := make([]int32, len(s))
	for i, v := range s {
		value, err := Int32(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Int64 converts the given string representation of an integer into int64.
func Int64(val string) (int64, error) {
	return strconv.ParseInt(val, 0, 64)
}

// Int64P ...
func Int64P(val string) (*int64, error) {
	return toPtr(Int64(val))
}

// Int64Slice converts 'val' where individual integers are separated by
// 'sep' into a int64 slice.
func Int64Slice(val, sep string) ([]int64, error) {
	s := strings.Split(val, sep)
	values := make([]int64, len(s))
	for i, v := range s {
		value, err := Int64(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Uint ...
func Uint(val string) (uint, error) {
	i, err := strconv.ParseUint(val, 0, 0)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}

// UintP ...
func UintP(val string) (*uint, error) {
	return toPtr(Uint(val))
}

// UintSlice converts 'val' where individual integers are separated by
// 'sep' into a uint slice.
func UintSlice(val, sep string) ([]uint, error) {
	s := strings.Split(val, sep)
	values := make([]uint, len(s))
	for i, v := range s {
		value, err := Uint(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Uint8 ...
func Uint8(val string) (uint8, error) {
	i, err := strconv.ParseUint(val, 0, 8)
	if err != nil {
		return 0, err
	}
	return uint8(i), nil
}

// Uint8P ...
func Uint8P(val string) (*uint8, error) {
	return toPtr(Uint8(val))
}

// Uint8Slice converts 'val' where individual integers are separated by
// 'sep' into a uint8 slice.
func Uint8Slice(val, sep string) ([]uint8, error) {
	s := strings.Split(val, sep)
	values := make([]uint8, len(s))
	for i, v := range s {
		value, err := Uint8(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Uint16 ...
func Uint16(val string) (uint16, error) {
	i, err := strconv.ParseUint(val, 0, 16)
	if err != nil {
		return 0, err
	}
	return uint16(i), nil
}

// Uint16P ...
func Uint16P(val string) (*uint16, error) {
	return toPtr(Uint16(val))
}

// Uint16Slice converts 'val' where individual integers are separated by
// 'sep' into a uint slice.
func Uint16Slice(val, sep string) ([]uint16, error) {
	s := strings.Split(val, sep)
	values := make([]uint16, len(s))
	for i, v := range s {
		value, err := Uint16(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Uint32 converts the given string representation of an integer into uint32.
func Uint32(val string) (uint32, error) {
	i, err := strconv.ParseUint(val, 0, 32)
	if err != nil {
		return 0, err
	}
	return uint32(i), nil
}

// Uint32P ...
func Uint32P(val string) (*uint32, error) {
	return toPtr(Uint32(val))
}

// Uint32Slice converts 'val' where individual integers are separated by
// 'sep' into a uint32 slice.
func Uint32Slice(val, sep string) ([]uint32, error) {
	s := strings.Split(val, sep)
	values := make([]uint32, len(s))
	for i, v := range s {
		value, err := Uint32(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Uint64 converts the given string representation of an integer into uint64.
func Uint64(val string) (uint64, error) {
	return strconv.ParseUint(val, 0, 64)
}

// Uint64P ...
func Uint64P(val string) (*uint64, error) {
	return toPtr(Uint64(val))
}

// Uint64Slice converts 'val' where individual integers are separated by
// 'sep' into a uint64 slice.
func Uint64Slice(val, sep string) ([]uint64, error) {
	s := strings.Split(val, sep)
	values := make([]uint64, len(s))
	for i, v := range s {
		value, err := Uint64(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// toPtr ...
func toPtr[T any](v T, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	return &v, nil
}
