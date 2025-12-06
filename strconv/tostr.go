package strconv

import (
	"reflect"
	"strconv"

	"golang.org/x/exp/constraints"
)

func FormatFor[T any](t T) string {
	return ReflectFormat(reflect.ValueOf(t))
}

func FormatAny(t any) string {
	return ReflectFormat(reflect.ValueOf(t))
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
