package strings

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

func Format(t any) string {
	return FormatReflectValue(reflect.ValueOf(t))
}

func FormatReflectValue(value reflect.Value) string {
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
			strs = append(strs, FormatReflectValue(value.Index(i)))
		}
		return strings.Join(strs, ",")
	}
	return ""
}

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

func FormatSigned[T constraints.Signed](v T) string {
	return strconv.FormatInt(int64(v), 10)
}

func FormatUnsigned[T constraints.Unsigned](v T) string {
	return strconv.FormatUint(uint64(v), 10)
}
