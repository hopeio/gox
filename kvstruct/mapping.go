// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package kvstruct

import (
	"reflect"
	"strings"
)

// Setter tries to set value on a walking by fields of a struct
type Setter interface {
	TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error)
}

type Setters []Setter

func (receiver Setters) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	for _, arg := range receiver {
		if arg != nil {
			isSet, err = arg.TrySet(value, field, key, opt)
			if isSet {
				return
			}
		}
	}
	return
}

func MappingByTag(ptr any, setter Setter, tag string) error {
	_, err := mapping(reflect.ValueOf(ptr), nil, setter, tag)
	return err
}

func mapping(value reflect.Value, field *reflect.StructField, setter Setter, tag string) (bool, error) {
	var tagValue string
	if field != nil {
		tagValue = field.Tag.Get(tag)
	}
	if tagValue == "-" { // just ignoring this field
		return false, nil
	}

	var vKind = value.Kind()

	if vKind == reflect.Ptr {
		var isNew bool
		vPtr := value
		if value.IsNil() {
			isNew = true
			vPtr = reflect.New(value.Type().Elem())
		}
		isSet, err := mapping(vPtr.Elem(), field, setter, tag)
		if err != nil {
			return false, err
		}
		if isNew && isSet {
			value.Set(vPtr)
		}
		return isSet, nil
	}

	if vKind == reflect.Struct {
		tValue := value.Type()

		var isSet bool
		for i := 0; i < value.NumField(); i++ {
			sf := tValue.Field(i)
			if sf.PkgPath != "" && !sf.Anonymous { // unexported
				continue
			}
			ok, err := mapping(value.Field(i), &sf, setter, tag)
			if err != nil {
				return false, err
			}
			isSet = isSet || ok
		}
		return isSet, nil
	}

	if field != nil && !field.Anonymous {
		ok, err := tryToSetValue(value, field, setter, tagValue)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}

type Options struct {
	Default   string
	Required  bool
	Omitempty bool
}

func ParseTag(tagValue string) (string, *Options) {
	if tagValue == "" { // when field is "emptyField" variable
		return "", nil
	}
	alias, opts, _ := strings.Cut(tagValue, ",")
	var opt string
	var setOpt Options
	for len(opts) > 0 {
		opt, opts, _ = strings.Cut(opts, ",")

		if k, v, _ := strings.Cut(opt, "="); k == "default" {
			setOpt.Default = v
			break
		}
		if opt == "required" {
			setOpt.Required = true
			break
		}
		if opt == "omitempty" {
			setOpt.Omitempty = true
			break
		}
	}
	return alias, &setOpt
}

func tryToSetValue(value reflect.Value, field *reflect.StructField, setter Setter, tagValue string) (bool, error) {

	alias, opts := ParseTag(tagValue)
	if alias == "" { // default value is FieldName
		alias = field.Name
	}
	return setter.TrySet(value, field, alias, opts)
}

type CanSetter interface {
	Setter
	Has(key string) bool
}

type CanSetters []CanSetter

func (args CanSetters) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *Options) (isSet bool, err error) {
	for _, arg := range args {
		if arg.Has(key) {
			return arg.TrySet(value, field, key, opt)
		}
	}
	return false, nil
}
