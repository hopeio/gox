package kvstruct

import (
	"reflect"
	"strings"

	encodingx "github.com/hopeio/gox/encoding/text"
)

type Get interface {
	Get(key string) (string, bool)
}

type Gets []Get

func (args Gets) Get(key string) (v string, ok bool) {
	for i := range args {
		if v, ok = args[i].Get(key); ok {
			return
		}
	}
	return
}
func (args Gets) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	return SetValueByKV(value, field, args, key, opt)
}

type KVSource map[string]string

func (form KVSource) Get(key string) (string, bool) {
	v, ok := form[key]
	return v, ok
}

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form KVSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	return SetValueByKV(value, field, form, key, opt)
}

type KVsSource map[string][]string

var _ Setter = KVsSource(nil)

func (form KVsSource) GetVs(key string) ([]string, bool) {
	v, ok := form[key]
	return v, ok
}

func (form KVsSource) Has(key string) bool {
	_, ok := form[key]
	return ok
}

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form KVsSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	return SetValueByKVs(value, field, form, key, opt)
}

type GetVs interface {
	GetVs(key string) ([]string, bool)
}

type GetVss []GetVs

func (args GetVss) GetVs(key string) (v []string, ok bool) {
	for i := range args {
		if v, ok = args[i].GetVs(key); ok {
			return
		}
	}
	return
}

func (args GetVss) Has(key string) bool {
	for i := range args {
		if _, ok := args[i].GetVs(key); ok {
			return ok
		}
	}
	return false
}

func (args GetVss) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	return SetValueByKVs(value, field, args, key, opt)
}

func SetValueByKV(value reflect.Value, field *reflect.StructField, kv Get, key string, opt *Options) (isSet bool, err error) {
	vs, _ := kv.Get(key)
	if vs == "" {
		if opt == nil || opt.Default == "" {
			return false, nil
		}
		vs = opt.Default
	}
	err = encodingx.ParseSetReflectValue(value, vs, field)
	if err != nil {
		return false, err
	}
	return true, nil
}

func SetValueByKVs(value reflect.Value, field *reflect.StructField, kv GetVs, key string, opt *Options) (isSet bool, err error) {
	vals, _ := kv.GetVs(key)
	if len(vals) == 0 {
		if opt == nil || opt.Default == "" {
			return false, nil
		}
		vals = strings.Split(opt.Default, ",")
	}

	err = encodingx.ParseStringsSetReflectValue(value, vals, field)
	if err != nil {
		return false, err
	}
	return true, nil
}
