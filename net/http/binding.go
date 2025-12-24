/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"strings"
	"sync"

	jsonx "github.com/hopeio/gox/encoding/json"
	iox "github.com/hopeio/gox/io"
	"github.com/hopeio/gox/kvstruct"
	stringsx "github.com/hopeio/gox/strings"
	"github.com/hopeio/gox/validator"
)

var (
	DefaultMemory    int64 = 32 << 20
	DefaultUnmarshal       = bodyUnmarshaler
	CommonTag              = "json"
	Validate               = validator.ValidateStruct
	defaultTags            = []string{"uri", "path", "query", "header", "form", CommonTag}
)

type Source interface {
	Uri() kvstruct.Setter
	Query() kvstruct.Setter
	Header() kvstruct.Setter
	Form() kvstruct.Setter
	Body() (string, io.ReadCloser)
}

type Field struct {
	Name  string
	Tags  []Tag
	Index int
	Field *reflect.StructField
}

type Tag struct {
	Key     string
	Value   string
	Options *kvstruct.Options
}

var cache = sync.Map{}

func Bind(r *http.Request, obj any) error {
	return CommonBind(RequestSource{r}, obj)
}

// unhandle multipart form data currently
func CommonBind(s Source, obj any) error {
	value := reflect.ValueOf(obj).Elem()
	typ := value.Type()
	header := s.Header()
	contentType, body := s.Body()
	if body != nil {
		var data []byte
		if raw, ok := body.(iox.Raw); ok {
			data = raw.Raw()
		} else {
			var err error
			data, err = io.ReadAll(body)
			if err != nil {
				return fmt.Errorf("read body error: %w", err)
			}
		}
		if len(data) == 0 {
			return nil
		}
		err := DefaultUnmarshal(contentType, data, obj)
		if err != nil {
			return err
		}
		if recorder, ok := body.(RecordBody); ok {
			recorder.RecordBody(data, obj)
		}
	}

	uriSetter, querySetter, headerSetter, multipartFormSetter := s.Uri(), s.Query(), header, s.Form()
	commonSetter := kvstruct.Setters([]kvstruct.Setter{uriSetter, querySetter, headerSetter, multipartFormSetter})
	var err error
	if fields, ok := cache.Load(typ); ok {
		var isSet bool
		for _, field := range fields.([]Field) {
			var setter kvstruct.Setter
			for _, tag := range field.Tags {
				switch tag.Key {
				case "uri", "path":
					setter = uriSetter
				case "query":
					setter = querySetter
				case "header":
					setter = headerSetter
				case "form":
					setter = multipartFormSetter
				case CommonTag:
					setter = commonSetter
				}
				if setter == nil {
					continue
				}
				isSet, err = setter.TrySet(value.Field(field.Index), field.Field, tag.Value, tag.Options)
				if err != nil {
					return err
				}
				if isSet {
					break
				}
			}
			if !isSet {
				setter = commonSetter
				isSet, err = setter.TrySet(value.Field(field.Index), field.Field, field.Name, nil)
				if err != nil {
					return err
				}
			}
		}
		return Validate(obj)
	}
	var fields []Field
	for i := 0; i < value.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous { // unexported
			continue
		}
		var tag, tagValue string
		var isSet bool
		var setter kvstruct.Setter
		var tags []Tag
		for _, tag = range defaultTags {
			tagValue = sf.Tag.Get(tag)
			if tagValue != "" && tagValue != "-" {
				switch tag {
				case "uri", "path":
					setter = uriSetter
				case "query":
					setter = querySetter
				case "form":
					setter = multipartFormSetter
				case "header":
					setter = headerSetter
				case CommonTag:
					setter = commonSetter
				}

				alias, options := kvstruct.ParseTag(tagValue)
				tags = append(tags, Tag{
					Key:     tag,
					Value:   alias,
					Options: options,
				})
				if setter == nil {
					continue
				}
				if !isSet {
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
			setter = commonSetter
			isSet, err = setter.TrySet(value.Field(i), &sf, field.Name, nil)
			if err != nil {
				return err
			}
		}
		fields = append(fields, field)
	}
	cache.Store(typ, fields)
	return Validate(obj)
}

type RequestSource struct {
	*http.Request
}

func (s RequestSource) Uri() kvstruct.Setter {
	return (*UriSource)(s.Request)
}

func (s RequestSource) Query() kvstruct.Setter {
	return (kvstruct.KVsSource)(s.URL.Query())
}

func (s RequestSource) Header() kvstruct.Setter {
	return (HeaderSource)(s.Request.Header)
}

func (s RequestSource) Form() kvstruct.Setter {
	contentType := s.Request.Header.Get(HeaderContentType)
	if strings.HasPrefix(contentType, ContentTypeForm) {
		data, err := io.ReadAll(s.Request.Body)
		if err != nil || len(data) == 0 {
			return nil
		}

		vs, err := url.ParseQuery(stringsx.FromBytes(data))
		if err != nil {
			return nil
		}
		if recorder, ok := s.Request.Body.(RecordBody); ok {
			recorder.RecordBody(data, nil)
		}
		return kvstruct.KVsSource(vs)
	}
	if strings.HasPrefix(contentType, ContentTypeMultipart) {
		err := s.ParseMultipartForm(DefaultMemory)
		if err != nil {
			return nil
		}
		return (*MultipartSource)(s.Request.MultipartForm)
	}
	return nil
}

func (s RequestSource) Body() (string, io.ReadCloser) {
	if s.Method == http.MethodGet {
		return "", nil
	}
	contentType := s.Request.Header.Get(HeaderContentType)
	if strings.HasPrefix(contentType, ContentTypeMultipart) || strings.HasPrefix(contentType, ContentTypeForm) {
		return "", nil
	}
	return contentType, s.Request.Body
}

func bodyUnmarshaler(contentType string, data []byte, obj any) error {
	if strings.HasPrefix(contentType, ContentTypeJson) {
		return jsonx.Unmarshal(data, obj)
	}
	return jsonx.Unmarshal(data, obj)
}

type HeaderSource map[string][]string

var _ kvstruct.Setter = HeaderSource(nil)

func (hs HeaderSource) GetVs(key string) ([]string, bool) {
	v, ok := hs[textproto.CanonicalMIMEHeaderKey(key)]
	return v, ok
}

func (hs HeaderSource) Has(key string) bool {
	_, ok := hs[textproto.CanonicalMIMEHeaderKey(key)]
	return ok
}
func (hs HeaderSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *kvstruct.Options) (isSet bool, err error) {
	return kvstruct.SetValueByKVs(value, field, hs, key, opt)
}

type UriSource http.Request

var _ kvstruct.Setter = (*UriSource)(nil)

func (req *UriSource) GetVs(key string) ([]string, bool) {
	if req.Pattern == "" {
		return nil, false
	}
	v := (*http.Request)(req).PathValue(key)
	return []string{v}, v != ""
}

func (req *UriSource) Has(key string) bool {
	v := (*http.Request)(req).PathValue(key)
	return v != ""
}

// TrySet tries to set a value by request's form source (like map[string][]string)
func (req *UriSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *kvstruct.Options) (isSet bool, err error) {
	return kvstruct.SetValueByKVs(value, field, req, key, opt)
}

type MultipartSource multipart.Form

var _ kvstruct.Setter = (*MultipartSource)(nil)

func (ms *MultipartSource) Has(key string) bool {
	if _, ok := ms.File[key]; ok {
		return true
	}
	_, ok := ms.Value[key]
	return ok
}

// TrySet tries to set a value by the multipart request with the binding a form file
func (ms *MultipartSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *kvstruct.Options) (isSet bool, err error) {
	if files := ms.File[key]; len(files) != 0 {
		return SetMultipartFrormFile(value, field, files)
	}

	return kvstruct.SetValueByKVs(value, field, kvstruct.KVsSource(ms.Value), key, opt)
}

func SetMultipartFrormFile(value reflect.Value, field *reflect.StructField, files []*multipart.FileHeader) (isSet bool, err error) {
	if len(files) == 0 {
		return false, nil
	}
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
		isSet, err = setArrayOfMultipartFormFile(slice, field, files)
		if err != nil || !isSet {
			return isSet, err
		}
		value.Set(slice)
		return true, nil
	case reflect.Array:
		return setArrayOfMultipartFormFile(value, field, files)
	}
	return false, errors.New("unsupported field type for multipart.FileHeader")
}

func setArrayOfMultipartFormFile(value reflect.Value, field *reflect.StructField, files []*multipart.FileHeader) (isSet bool, err error) {
	if value.Len() != len(files) {
		return false, errors.New("unsupported len for []*multipart.FileHeader")
	}
	for i := range files {
		setted, err := SetMultipartFrormFile(value.Index(i), field, files[i:i+1])
		if err != nil || !setted {
			return setted, err
		}
	}
	return true, nil
}
