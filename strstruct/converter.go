package strstruct

import (
	"reflect"
	"strconv"

	stringsx "github.com/hopeio/gox/strings"
)

var Sep = ","

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

var stringConverterSliceArrays = [...]StringConverterE{
	reflect.Invalid: nil,
	reflect.Bool:    stringConvertBoolSlice,
	reflect.Int:     stringConvertIntSlice,
	reflect.Int8:    stringConvertInt8Slice,
	reflect.Int16:   stringConvertInt16Slice,
	reflect.Int32:   stringConvertInt32Slice,
	reflect.Int64:   stringConvertInt64Slice,
	reflect.Uint:    stringConvertUintSlice,
	reflect.Uint8:   stringConvertUint8Slice,
	reflect.Uint16:  stringConvertUint16Slice,
	reflect.Uint32:  stringConvertUint32Slice,
	reflect.Uint64:  stringConvertUint64Slice,
	reflect.Uintptr: stringConvertUintSlice,
	reflect.Float32: stringConvertFloat32Slice,
	reflect.Float64: stringConvertFloat64Slice,
}

func GetStringConverter(typ reflect.Type) StringConverter {
	return GetStringConverterE(typ).IgnoreError()
}

func GetStringConverterE(typ reflect.Type) StringConverterE {
	kind := typ.Kind()
	if kind == reflect.Slice || kind == reflect.Array {
		return GetStringSliceConverter(typ.Elem())
	}
	return GetStringConverterEByKind(kind)
}

func GetStringSliceConverter(elemTyp reflect.Type) func(value string) (any, error) {
	return GetStringSliceConverterByKind(elemTyp.Kind())
}

func GetStringSliceConverterByKind(kind reflect.Kind) func(value string) (any, error) {
	if kind == reflect.String {
		return stringConvertString
	}
	if kind > reflect.Uint64 {
		return nil
	}
	return stringConverterSliceArrays[kind]
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
	return strconv.ParseBool(value)
}
func stringConvertBoolSlice(value string) (any, error) {
	return stringsx.BoolSlice(value, Sep)
}
func stringConvertFloat32(value string) (any, error) {
	return stringsx.Float32(value)
}
func stringConvertFloat32Slice(value string) (any, error) {
	return stringsx.Float32Slice(value, Sep)
}
func stringConvertFloat64(value string) (any, error) {
	return strconv.ParseFloat(value, 64)
}
func stringConvertFloat64Slice(value string) (any, error) {
	return stringsx.Float64Slice(value, Sep)
}
func stringConvertInt(value string) (any, error) {
	return stringsx.Int(value)
}
func stringConvertIntSlice(value string) (any, error) {
	return stringsx.IntSlice(value, Sep)
}
func stringConvertInt8(value string) (any, error) {
	return stringsx.Int8(value)
}
func stringConvertInt8Slice(value string) (any, error) {
	return stringsx.Int8Slice(value, Sep)
}
func stringConvertInt16(value string) (any, error) {
	return stringsx.Int16(value)
}
func stringConvertInt16Slice(value string) (any, error) {
	return stringsx.Int16Slice(value, Sep)
}
func stringConvertInt32(value string) (any, error) {
	return stringsx.Int32(value)
}
func stringConvertInt32Slice(value string) (any, error) {
	return stringsx.Int32Slice(value, Sep)
}
func stringConvertInt64(value string) (any, error) {
	return strconv.ParseInt(value, 10, 64)
}
func stringConvertInt64Slice(value string) (any, error) {
	return stringsx.Int64Slice(value, Sep)
}
func stringConvertString(value string) (any, error) {
	return value, nil
}

func stringConvertStringSlice(value string) (any, error) {
	return stringsx.StringSlice(value, Sep)
}
func stringConvertUint(value string) (any, error) {
	return stringsx.Uint(value)
}
func stringConvertUintSlice(value string) (any, error) {
	return stringsx.UintSlice(value, Sep)
}
func stringConvertUint8(value string) (any, error) {
	return stringsx.Uint8(value)
}
func stringConvertUint8Slice(value string) (any, error) {
	return stringsx.Uint8Slice(value, Sep)
}
func stringConvertUint16(value string) (any, error) {
	return stringsx.Uint16(value)
}
func stringConvertUint16Slice(value string) (any, error) {
	return stringsx.Uint16Slice(value, Sep)
}
func stringConvertUint32(value string) (any, error) {
	return stringsx.Uint32(value)
}
func stringConvertUint32Slice(value string) (any, error) {
	return stringsx.Uint32Slice(value, Sep)
}
func stringConvertUint64(value string) (any, error) {
	return strconv.ParseUint(value, 10, 64)
}
func stringConvertUint64Slice(value string) (any, error) {
	return stringsx.Uint64Slice(value, Sep)
}
