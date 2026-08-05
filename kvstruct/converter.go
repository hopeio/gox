package kvstruct

import (
	"reflect"
	"strconv"

	stringsx "github.com/hopeio/gox/strings"
)

var Sep = ","

type StringConverter func(string) any
type StringConverterE func(string) (any, error)

// IgnoreError returns the result.
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

// GetStringConverter returns the value.
func GetStringConverter(typ reflect.Type) StringConverter {
	return GetStringConverterE(typ).IgnoreError()
}

// GetStringConverterE returns the value.
func GetStringConverterE(typ reflect.Type) StringConverterE {
	kind := typ.Kind()
	if kind == reflect.Slice || kind == reflect.Array {
		return GetStringSliceConverter(typ.Elem())
	}
	return GetStringConverterEByKind(kind)
}

// GetStringSliceConverter returns the value.
func GetStringSliceConverter(elemTyp reflect.Type) func(value string) (any, error) {
	return GetStringSliceConverterByKind(elemTyp.Kind())
}

// GetStringSliceConverterByKind returns the value.
func GetStringSliceConverterByKind(kind reflect.Kind) func(value string) (any, error) {
	if kind == reflect.String {
		return stringConvertString
	}
	if kind > reflect.Uint64 {
		return nil
	}
	return stringConverterSliceArrays[kind]
}

// GetStringConverterByKind returns the value.
func GetStringConverterByKind(kind reflect.Kind) StringConverter {
	return GetStringConverterEByKind(kind).IgnoreError()
}

// GetStringConverterEByKind returns the value.
func GetStringConverterEByKind(kind reflect.Kind) StringConverterE {
	if kind == reflect.String {
		return stringConvertString
	}
	if kind > reflect.Uint64 {
		return nil
	}
	return stringConverterArrays[kind]
}

// stringConvertBool performs the operation.
func stringConvertBool(value string) (any, error) {
	return strconv.ParseBool(value)
}

// stringConvertBoolSlice performs the operation.
func stringConvertBoolSlice(value string) (any, error) {
	return stringsx.BoolSlice(value, Sep)
}

// stringConvertFloat32 performs the operation.
func stringConvertFloat32(value string) (any, error) {
	return stringsx.Float32(value)
}

// stringConvertFloat32Slice performs the operation.
func stringConvertFloat32Slice(value string) (any, error) {
	return stringsx.Float32Slice(value, Sep)
}

// stringConvertFloat64 performs the operation.
func stringConvertFloat64(value string) (any, error) {
	return strconv.ParseFloat(value, 64)
}

// stringConvertFloat64Slice performs the operation.
func stringConvertFloat64Slice(value string) (any, error) {
	return stringsx.Float64Slice(value, Sep)
}

// stringConvertInt performs the operation.
func stringConvertInt(value string) (any, error) {
	return stringsx.Int(value)
}

// stringConvertIntSlice performs the operation.
func stringConvertIntSlice(value string) (any, error) {
	return stringsx.IntSlice(value, Sep)
}

// stringConvertInt8 performs the operation.
func stringConvertInt8(value string) (any, error) {
	return stringsx.Int8(value)
}

// stringConvertInt8Slice performs the operation.
func stringConvertInt8Slice(value string) (any, error) {
	return stringsx.Int8Slice(value, Sep)
}

// stringConvertInt16 performs the operation.
func stringConvertInt16(value string) (any, error) {
	return stringsx.Int16(value)
}

// stringConvertInt16Slice performs the operation.
func stringConvertInt16Slice(value string) (any, error) {
	return stringsx.Int16Slice(value, Sep)
}

// stringConvertInt32 performs the operation.
func stringConvertInt32(value string) (any, error) {
	return stringsx.Int32(value)
}

// stringConvertInt32Slice performs the operation.
func stringConvertInt32Slice(value string) (any, error) {
	return stringsx.Int32Slice(value, Sep)
}

// stringConvertInt64 performs the operation.
func stringConvertInt64(value string) (any, error) {
	return strconv.ParseInt(value, 10, 64)
}

// stringConvertInt64Slice performs the operation.
func stringConvertInt64Slice(value string) (any, error) {
	return stringsx.Int64Slice(value, Sep)
}

// stringConvertString performs the operation.
func stringConvertString(value string) (any, error) {
	return value, nil
}

// stringConvertStringSlice performs the operation.
func stringConvertStringSlice(value string) (any, error) {
	return stringsx.StringSlice(value, Sep)
}

// stringConvertUint performs the operation.
func stringConvertUint(value string) (any, error) {
	return stringsx.Uint(value)
}

// stringConvertUintSlice performs the operation.
func stringConvertUintSlice(value string) (any, error) {
	return stringsx.UintSlice(value, Sep)
}

// stringConvertUint8 performs the operation.
func stringConvertUint8(value string) (any, error) {
	return stringsx.Uint8(value)
}

// stringConvertUint8Slice performs the operation.
func stringConvertUint8Slice(value string) (any, error) {
	return stringsx.Uint8Slice(value, Sep)
}

// stringConvertUint16 performs the operation.
func stringConvertUint16(value string) (any, error) {
	return stringsx.Uint16(value)
}

// stringConvertUint16Slice performs the operation.
func stringConvertUint16Slice(value string) (any, error) {
	return stringsx.Uint16Slice(value, Sep)
}

// stringConvertUint32 performs the operation.
func stringConvertUint32(value string) (any, error) {
	return stringsx.Uint32(value)
}

// stringConvertUint32Slice performs the operation.
func stringConvertUint32Slice(value string) (any, error) {
	return stringsx.Uint32Slice(value, Sep)
}

// stringConvertUint64 performs the operation.
func stringConvertUint64(value string) (any, error) {
	return strconv.ParseUint(value, 10, 64)
}

// stringConvertUint64Slice performs the operation.
func stringConvertUint64Slice(value string) (any, error) {
	return stringsx.Uint64Slice(value, Sep)
}
