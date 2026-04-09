/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"strings"
	"sync"

	iox "github.com/hopeio/gox/io"
	stringsx "github.com/hopeio/gox/strings"
	"github.com/hopeio/gox/strstruct"
	"github.com/hopeio/gox/validator"
)

var (
	DefaultMemory int64 = 32 << 20
	CommonTag           = "json"
	Validate            = validator.ValidateStruct
	defaultTags         = []string{"uri", "path", "query", "header", "form", CommonTag}
)

type Source interface {
	Uri() strstruct.Getter
	Query() strstruct.ValuesGetter
	Header() strstruct.ValuesGetter
	Body() (context.Context, string, io.ReadCloser)
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
	Options *strstruct.Options
}

var cache = sync.Map{}

type Binder interface {
	Bind(r *http.Request, v any) error
}

type CommonBinder interface {
	Bind(r Source, v any) error
}

func Bind(r *http.Request, v any) error {
	if b, ok := v.(Binder); ok {
		return b.Bind(r, v)
	}
	return CommonBind(RequestSource{r}, v)
}

// unhandle multipart form data currently
func CommonBind(s Source, v any) error {
	if b, ok := s.(CommonBinder); ok {
		err := b.Bind(s, v)
		if err != nil {
			return err
		}
		return Validate(v)
	}
	value := reflect.ValueOf(v).Elem()
	typ := value.Type()
	header := s.Header()
	var multipartFormSetter strstruct.Setter
	ctx, contentType, body := s.Body()
	if body != nil {
		if strings.HasPrefix(contentType, ContentTypeForm) {
			data, err := io.ReadAll(body)
			if err != nil {
				return err
			}

			vs, err := url.ParseQuery(stringsx.FromBytes(data))
			if err != nil {
				return nil
			}
			if recorder, ok := body.(RecordBodyer); ok {
				recorder.RecordBody(data, nil)
			}
			multipartFormSetter = strstruct.KVsSource(vs)
		}else if strings.HasPrefix(contentType, ContentTypeMultipart) {
			mr,err := multipartReader(true,contentType,body)
			if err != nil {
				return nil
			}
			multipartForm, err := mr.ReadForm(DefaultMemory)
			if err != nil {
				return err
			}
			multipartFormSetter = (*MultipartSource)(multipartForm)
		}else{
			var data []byte
		if raw, ok := body.(iox.RawByter); ok {
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
		err := DefaultUnmarshal(ctx, contentType, data, v)
		if err != nil {
			return err
		}
		if recorder, ok := body.(RecordBodyer); ok {
			recorder.RecordBody(data, v)
		}
		return DefaultUnmarshal(ctx, contentType, data, v)
		}

	}

	uriSetter, querySetter, headerSetter := strstruct.GetFunc(s.Uri().Get), strstruct.ValuesGetFunc(s.Query().Get), strstruct.ValuesGetFunc(header.Get)
	commonSetter := strstruct.Setters([]strstruct.Setter{uriSetter, querySetter, headerSetter, multipartFormSetter})
	var err error
	if fields, ok := cache.Load(typ); ok {
		var isSet bool
		for _, field := range fields.([]Field) {
			var setter strstruct.Setter
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
		return Validate(v)
	}
	var fields []Field
	for i := 0; i < value.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous { // unexported
			continue
		}
		var tag, tagValue string
		var isSet bool
		var setter strstruct.Setter
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

				alias, options := strstruct.ParseTag(tagValue)
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
	return Validate(v)
}

func multipartReader(allowMixed bool,contentType string,body io.Reader) (*multipart.Reader, error) {
	if contentType == "" {
		return nil, http.ErrNotMultipart
	}
	if body == nil {
		return nil, errors.New("missing form body")
	}
	d, params, err := mime.ParseMediaType(contentType)
	if err != nil || !(d == "multipart/form-data" || allowMixed && d == "multipart/mixed") {
		return nil, http.ErrNotMultipart
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, http.ErrMissingBoundary
	}
	return multipart.NewReader(body, boundary), nil
}

type RequestSource struct {
	*http.Request
}

func (s RequestSource) Uri() strstruct.Getter {
	return (*UriSource)(s.Request)
}

func (s RequestSource) Query() strstruct.ValuesGetter {
	return (strstruct.KVsSource)(s.URL.Query())
}

func (s RequestSource) Header() strstruct.ValuesGetter {
	return (HeaderSource)(s.Request.Header)
}

func (s RequestSource) Body() (context.Context, string, io.ReadCloser) {
	if s.Method == http.MethodGet {
		return s.Context(), "", nil
	}
	contentType := s.Request.Header.Get(HeaderContentType)
	if strings.HasPrefix(contentType, ContentTypeMultipart) || strings.HasPrefix(contentType, ContentTypeForm) {
		return s.Context(), contentType, nil
	}
	return s.Context(), contentType, s.Request.Body
}

type HeaderSource map[string][]string

var _ strstruct.ValuesGetter = HeaderSource(nil)

func (hs HeaderSource) Get(key string) ([]string, bool) {
	v, ok := hs[textproto.CanonicalMIMEHeaderKey(key)]
	for i := range v {
		v[i], _ = url.QueryUnescape(v[i])
	}
	return v, ok
}


type UriSource http.Request

var _ strstruct.Getter = (*UriSource)(nil)

func (req *UriSource) Get(key string) (string, bool) {
	if req.Pattern == "" {
		return "", false
	}
	v := (*http.Request)(req).PathValue(key)
	return v, v != ""
}


type QuerySource map[string][]string

var _ strstruct.Getter = (*UriSource)(nil)

func (req QuerySource) Get(key string) ([]string, bool) {
	v, ok := req[key]
	for i := range v {
		v[i], _ = url.QueryUnescape(v[i])
	}
	return v, ok
}

type MultipartSource multipart.Form

var _ strstruct.Setter = (*MultipartSource)(nil)

// TrySet tries to set a value by the multipart request with the binding a form file
func (ms *MultipartSource) TrySet(value reflect.Value, field *reflect.StructField, key string, opt *strstruct.Options) (isSet bool, err error) {
	if files := ms.File[key]; len(files) != 0 {
		return SetMultipartFrormFile(value, field, files)
	}

	return strstruct.SetValueByValuesGetter(value, field, QuerySource(ms.Value), key, opt)
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
