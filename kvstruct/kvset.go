package kvstruct

import (
	"reflect"
	"strings"

	encodingx "github.com/hopeio/gox/strconv"
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
	err = encodingx.ParseReflectSet(value, vs, field)
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
func (form KVsSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
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

func (args Args2) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
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

func (args PeekVsSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	return SetValueByKVsWithStructField(value, field, args, key, opt)
}

func SetValueByKVsWithStructField(value reflect.Value, field *reflect.StructField, kv PeekVs, key string, opt *Options) (isSet bool, err error) {
	vals, _ := kv.Peek(key)
	if len(vals) == 0 {
		if opt == nil || opt.Default == "" {
			return false, nil
		}
		vals = strings.Split(opt.Default, ",")
	}

	err = encodingx.ParseStringsReflectSet(value, vals, field)
	if err != nil {
		return false, err
	}
	return true, nil
}
