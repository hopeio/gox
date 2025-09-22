/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package text

import (
	"encoding"
	reflectx "github.com/hopeio/gox/reflect/converter"
	stringsx "github.com/hopeio/gox/strings"
	"golang.org/x/exp/constraints"
	"golang.org/x/net/html/charset"
	tencoding "golang.org/x/text/encoding"
	"strconv"
)

func DetermineEncoding(content []byte, contentType string) (e tencoding.Encoding, name string, certain bool) {
	return charset.DetermineEncoding(content, contentType)
}

func Unmarshal[T any](str string) error {
	var t T
	v, vp := any(t), any(&t)
	itv, ok := v.(encoding.TextUnmarshaler)
	if !ok {
		itv, ok = vp.(encoding.TextUnmarshaler)
	}
	if ok {
		err := itv.UnmarshalText([]byte(str))
		if err != nil {
			return err
		}
	}
	return nil
}

func StringConvertFor[T any](str string) (T, error) {
	var t T
	a, ap := any(t), any(&t)
	itv, ok := a.(encoding.TextUnmarshaler)
	if !ok {
		itv, ok = ap.(encoding.TextUnmarshaler)
	}
	if ok {
		err := itv.UnmarshalText(stringsx.ToBytes(str))
		if err != nil {
			return t, err
		}
		return t, nil
	}

	v, err := reflectx.StringConvertFor[T](str)
	if err != nil {
		return t, err
	}
	return v, nil
}

func IntToString[T constraints.Integer](v T) string {
	return strconv.FormatInt(int64(v), 10)
}

func UintToString[T constraints.Integer](v T) string {
	return strconv.FormatUint(uint64(v), 10)
}

func AnyIntToString(value interface{}) string {
	switch v := value.(type) {
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	}
	return ""
}
