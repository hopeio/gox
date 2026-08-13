package strings

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

// Format formats or converts the value.
func Format(t any) string {
	return FormatReflectValue(reflect.ValueOf(t))
}

// FormatReflectValue formats or converts the value.
func FormatReflectValue(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}
	if value.CanInterface() {
		if t, ok := value.Interface().(encoding.TextMarshaler); ok {
			s, _ := t.MarshalText()
			return string(s)
		}
	}

	kind := value.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Pointer:
		// value.Int() 对指针 kind 会 panic；格式化其指向的值
		if value.IsNil() {
			return ""
		}
		return FormatReflectValue(value.Elem())
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(value.Float(), 'g', -1, 32)
	case reflect.Array, reflect.Slice:
		var strs []string
		for i := 0; i < value.Len(); i++ {
			strs = append(strs, FormatReflectValue(value.Index(i)))
		}
		return strings.Join(strs, ",")
	}
	return ""
}

// FormatInteger formats or converts the value.
func FormatInteger(value any) string {
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

// FormatSigned formats or converts the value.
func FormatSigned[T constraints.Signed](v T) string {
	return strconv.FormatInt(int64(v), 10)
}

// FormatUnsigned formats or converts the value.
func FormatUnsigned[T constraints.Unsigned](v T) string {
	return strconv.FormatUint(uint64(v), 10)
}
