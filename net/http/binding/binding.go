/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package binding

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"

	jsonx "github.com/hopeio/gox/encoding/json"
	"github.com/hopeio/gox/kvstruct"
	httpx "github.com/hopeio/gox/net/http"
	stringsx "github.com/hopeio/gox/strings"
	"github.com/hopeio/gox/validator"
)

var (
	DefaultMemory    int64 = 32 << 20
	BodyUnmarshaller       = bodyUnmarshaller
	CommonTag              = "json"
	Validate               = validator.ValidateStruct
	defaultTags            = []string{"uri", "path", "query", "header", "form", CommonTag}
)

type Source interface {
	Uri() kvstruct.Setter
	Query() kvstruct.Setter
	Header() kvstruct.Setter
	MultipartForm() kvstruct.Setter
	BodyBind(obj any) error
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
	err := s.BodyBind(obj)
	if err != nil {
		return err
	}
	uriSetter, querySetter, headerSetter, multipartFormSetter := s.Uri(), s.Query(), s.Header(), s.MultipartForm()
	commonSetter := kvstruct.Setters{Setters: []kvstruct.Setter{uriSetter, querySetter, headerSetter, multipartFormSetter}}
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

func (s RequestSource) MultipartForm() kvstruct.Setter {
	contentType := s.Request.Header.Get(httpx.HeaderContentType)
	if strings.HasPrefix(contentType, httpx.ContentTypeMultipart) {
		err := s.ParseMultipartForm(DefaultMemory)
		if err != nil {
			return nil
		}
		return (*MultipartSource)(s.Request.MultipartForm)
	}
	return nil
}

func (s RequestSource) BodyBind(obj any) error {
	if s.Method == http.MethodGet {
		return nil
	}
	contentType := s.Request.Header.Get(httpx.HeaderContentType)
	if strings.HasPrefix(contentType, httpx.ContentTypeMultipart) {
		return nil
	}
	data, err := io.ReadAll(s.Body)
	if err != nil {
		return fmt.Errorf("read body error: %w", err)
	}
	if recorder, ok := s.Body.(httpx.RequestRecorder); ok {
		recorder.RecordRequest(contentType, data, obj)
	}
	if len(data) == 0 {
		return nil
	}
	return BodyUnmarshaller(contentType, data, obj)
}

func bodyUnmarshaller(contentType string, data []byte, obj any) error {
	if strings.HasPrefix(contentType, httpx.ContentTypeForm) {
		return FormUnmarshal(data, obj)
	}
	if strings.HasPrefix(contentType, httpx.ContentTypeJson) {
		return jsonx.Unmarshal(data, obj)
	}
	return jsonx.Unmarshal(data, obj)
}
