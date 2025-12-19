/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package binding

import (
	"errors"
	"mime/multipart"
	"net/url"
	"reflect"

	"github.com/hopeio/gox/kvstruct"
	stringsx "github.com/hopeio/gox/strings"
)

type MultipartSource multipart.Form

var _ kvstruct.Setter = (*MultipartSource)(nil)

func (ms *MultipartSource) HasValue(key string) bool {
	if _, ok := ms.File[key]; ok {
		return true
	}
	_, ok := ms.Value[key]
	return ok
}

// TrySet tries to set a value by the multipart request with the binding a form file
func (ms *MultipartSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *kvstruct.Options) (isSet bool, err error) {
	if files := ms.File[key]; len(files) != 0 {
		return SetByMultipartFormFile(value, field, files)
	}

	return kvstruct.SetValueByKVsWithStructField(value, field, kvstruct.KVsSource(ms.Value), key, opt)
}

func SetByMultipartFormFile(value reflect.Value, field *reflect.StructField, files []*multipart.FileHeader) (isSet bool, err error) {
	switch value.Kind() {
	case reflect.Ptr:
		switch value.Interface().(type) {
		case *multipart.FileHeader:
			value.Set(reflect.ValueOf(files[0]))
			return true, nil
		}
	case reflect.Struct:
		switch value.Interface().(type) {
		case multipart.FileHeader:
			value.Set(reflect.ValueOf(files[0]).Elem())
			return true, nil
		}
	case reflect.Slice:
		slice := reflect.MakeSlice(value.Type(), len(files), len(files))
		isSet, err = setArrayOfMultipartFormFiles(slice, field, files)
		if err != nil || !isSet {
			return isSet, err
		}
		value.Set(slice)
		return true, nil
	case reflect.Array:
		return setArrayOfMultipartFormFiles(value, field, files)
	}
	return false, errors.New("unsupported field type for multipart.FileHeader")
}

func setArrayOfMultipartFormFiles(value reflect.Value, field *reflect.StructField, files []*multipart.FileHeader) (isSet bool, err error) {
	if value.Len() != len(files) {
		return false, errors.New("unsupported len for []*multipart.FileHeader")
	}
	for i := range files {
		setted, err := SetByMultipartFormFile(value.Index(i), field, files[i:i+1])
		if err != nil || !setted {
			return setted, err
		}
	}
	return true, nil
}

func FormUnmarshal(data []byte, obj any) error {
	vs, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}
	setter := kvstruct.KVsSource(vs)
	value := reflect.ValueOf(obj).Elem()
	typ := value.Type()
	if fields, ok := cache.Load(typ); ok {
		for _, field := range fields.([]Field) {
			_, err = setter.TrySet(value.Field(field.Index), field.Field, field.Name, nil)
			if err != nil {
				return err
			}
		}
	} else {
		var fields []Field
		for i := 0; i < value.NumField(); i++ {
			sf := typ.Field(i)
			if sf.PkgPath != "" && !sf.Anonymous { // unexported
				continue
			}
			var tag, tagValue string
			var isSet bool
			var tags []Tag
			for _, tag = range defaultTags {
				tagValue = sf.Tag.Get(tag)
				if tagValue != "" && tagValue != "-" {
					alias, options := kvstruct.ParseTag(tagValue)
					tags = append(tags, Tag{
						Key:     tag,
						Value:   alias,
						Options: options,
					})
					if setter == nil {
						continue
					}
					if tag == "form" {
						isSet, err = setter.TrySet(value.Field(i), &sf, alias, options)
						if err != nil {
							return err
						}
					}
				}
			}
			field := Field{
				Name:  stringsx.LowerCaseFirst(sf.Name),
				Tags:  tags,
				Index: i,
				Field: &sf,
			}

			if !isSet {
				isSet, err = setter.TrySet(value.Field(i), &sf, field.Name, nil)
				if err != nil {
					return err
				}
			}
			fields = append(fields, field)
			cache.Store(typ, fields)
		}
	}
	return nil
}
