package mtos

import (
	"encoding"
	"encoding/json"
	"fmt"
	stringsx "github.com/hopeio/gox/strings"
	"reflect"
	"strings"
	"time"
)

type PeekV interface {
	Peek(key string) (string, bool)
}

type Args []PeekV

func (args Args) Peek(key string) (v string, ok bool) {
	for i := range args {
		if v, ok = args[i].Peek(key); ok {
			return
		}
	}
	return
}
func (args Args) TrySet(value reflect.Value, field *reflect.StructField, key string) (isSet bool, err error) {
	return SetByKV(value, field, args, key)
}

func SetByKV(value reflect.Value, field *reflect.StructField, kv PeekV, key string) (isSet bool, err error) {
	vs, ok := kv.Peek(key)
	if !ok {
		return false, nil
	}
	err = SetValueByStringWithStructField(value, field, vs)
	if err != nil {
		return false, err
	}
	return true, nil
}

type KVSource map[string]string

func (form KVSource) Peek(key string) (string, bool) {
	v, ok := form[key]
	return v, ok
}

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form KVSource) TrySet(value reflect.Value, field *reflect.StructField, key string) (isSet bool, err error) {
	return SetByKV(value, field, form, key)
}

type KVsSource map[string][]string

var _ Setter = KVsSource(nil)

func (form KVsSource) Peek(key string) ([]string, bool) {
	v, ok := form[key]
	return v, ok
}

func (form KVsSource) HasValue(key string) bool {
	_, ok := form[key]
	return ok
}

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form KVsSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt SetOptions) (isSet bool, err error) {
	return SetValueByKVsWithStructField(value, field, form, key, opt)
}

type PeekVs interface {
	Peek(key string) ([]string, bool)
}

type Args2 []PeekVs

func (args Args2) Peek(key string) (v []string, ok bool) {
	for i := range args {
		if v, ok = args[i].Peek(key); ok {
			return
		}
	}
	return
}

func (args Args2) HasValue(key string) bool {
	for i := range args {
		if _, ok := args[i].Peek(key); ok {
			return ok
		}
	}
	return false
}

func (args Args2) TrySet(value reflect.Value, field *reflect.StructField, key string, opt SetOptions) (isSet bool, err error) {
	return SetValueByKVsWithStructField(value, field, args, key, opt)
}

type PeekVsSource []PeekVs

func (args PeekVsSource) Peek(key string) (v []string, ok bool) {
	for i := range args {
		if v, ok = args[i].Peek(key); ok {
			return
		}
	}
	return
}

func (args PeekVsSource) HasValue(key string) bool {
	for i := range args {
		if _, ok := args[i].Peek(key); ok {
			return ok
		}
	}
	return false
}

func (args PeekVsSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt SetOptions) (isSet bool, err error) {
	return SetValueByKVsWithStructField(value, field, args, key, opt)
}

func SetValueByKVsWithStructField(value reflect.Value, field *reflect.StructField, kv PeekVs, key string, opt SetOptions) (isSet bool, err error) {
	vs, ok := kv.Peek(key)
	if !ok && opt.defaultValue == "" {
		return false, nil
	}

	switch value.Kind() {
	case reflect.Slice:
		if !ok {
			vs = []string{opt.defaultValue}
		}
		return true, setSlice(vs, value, field)
	case reflect.Array:
		if !ok {
			vs = []string{opt.defaultValue}
		}
		if len(vs) != value.Len() {
			return false, fmt.Errorf("%q is not valid value for %s", vs, value.Type().String())
		}
		return true, setArray(vs, value, field)
	default:
		var val string
		if !ok {
			val = opt.defaultValue
		}

		if len(vs) > 0 {
			val = vs[0]
		}
		return true, SetValueByStringWithStructField(value, field, val)
	}
}

func SetValueByStringWithStructField(value reflect.Value, field *reflect.StructField, val string) error {
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
		if kind == reflect.Slice {
			value.Set(reflect.MakeSlice(typ, len(strs), len(strs)))
		}
		for i := 0; i < value.Len(); i++ {
			if err := SetValueByString(value.Index(i), strs[i]); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		switch anyV.(type) {
		case time.Time:
			return setTimeField(val, field, value)
		}
		return json.Unmarshal(stringsx.ToBytes(val), value.Addr().Interface())
	case reflect.Map:
		return json.Unmarshal(stringsx.ToBytes(val), value.Addr().Interface())
	default:
		return errUnknownType
	}
	return nil
}

func SetFieldByString(dst any, field, value string) error {
	if value == "" {
		return nil
	}

	fieldValue := reflect.ValueOf(dst).Elem().FieldByName(field)
	return SetValueByString(fieldValue, value)
}

func SetValueByString(value reflect.Value, val string) error {
	return SetValueByStringWithStructField(value, nil, val)
}
