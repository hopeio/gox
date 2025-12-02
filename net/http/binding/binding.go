/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package binding

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"

	http2 "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/reflect/mtos"
	"github.com/hopeio/gox/validation/validator"
)

var (
	DefaultMemory    int64                   = 32 << 20
	BodyUnmarshaller func([]byte, any) error = json.Unmarshal
)

const commonTag = "json"

var Validate = validator.ValidateStruct

var defaultTags = []string{"uri", "path", "query", "header", "form", commonTag}

func CommonTag(tag string) {
	defaultTags[5] = tag
}

type Source interface {
	Uri() mtos.Setter
	Query() mtos.Setter
	Header() mtos.Setter
	Form() mtos.Setter
	BodyBind(obj any) error
}

type Field struct {
	Tags  []Tag
	Index int
	Field *reflect.StructField
}

type Tag struct {
	Key     string
	Value   string
	Options mtos.SetOptions
}

var cache = sync.Map{}

func Bind(r *http.Request, obj any) error {
	return CommonBind(RequestSource{r}, obj)
}

func CommonBind(s Source, obj any) error {
	value := reflect.ValueOf(obj).Elem()
	typ := value.Type()
	err := s.BodyBind(obj)
	if err != nil {
		return err
	}
	uriSetter, querySetter, headerSetter, formSetter := s.Uri(), s.Query(), s.Header(), s.Form()
	commonSetter := mtos.Setters{Setters: []mtos.Setter{uriSetter, querySetter, headerSetter}}
	if fields, ok := cache.Load(typ); ok {
		var isSet bool
		for _, field := range fields.([]Field) {
			var setter mtos.Setter
			for _, tag := range field.Tags {
				switch tag.Key {
				case "uri", "path":
					setter = uriSetter
				case "query":
					setter = querySetter
				case "header":
					setter = headerSetter
				case "form":
					setter = formSetter
				case commonTag:
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
		}
		return Validate(obj)
	}
	var fields []Field
	for i := 0; i < value.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous { // unexported
			continue
		}
		var tagValue string
		var tag string
		var isSet bool
		var setter mtos.Setter
		for _, tag = range defaultTags {
			tagValue = sf.Tag.Get(tag)
			var tags []Tag
			if tagValue != "" && tagValue != "-" {
				switch tag {
				case "uri", "path":
					setter = uriSetter
				case "query":
					setter = querySetter
				case "header":
					setter = headerSetter
				case "form":
					setter = formSetter
				case commonTag:
					setter = commonSetter
				}
				tagValues := strings.Split(tagValue, ",")
				tagValue = tagValues[0]
				options := mtos.SetOptions{}
				tags = append(tags, Tag{
					Key:     tag,
					Value:   tagValue,
					Options: options,
				})
				if setter == nil {
					continue
				}
				isSet, err = setter.TrySet(value.Field(i), &sf, tagValue, options)
				if err != nil {
					return err
				}
				field := Field{
					Tags:  tags,
					Index: i,
					Field: &sf,
				}
				fields = append(fields, field)
				if isSet {
					break
				}
			}
		}
	}
	cache.Store(typ, fields)
	return Validate(obj)
}

type RequestSource struct {
	*http.Request
}

func (s RequestSource) Uri() mtos.Setter {
	return (*UriSource)(s.Request)
}

func (s RequestSource) Query() mtos.Setter {
	return (mtos.KVsSource)(s.URL.Query())
}

func (s RequestSource) Header() mtos.Setter {
	return (HeaderSource)(s.Request.Header)
}

func (s RequestSource) Form() mtos.Setter {
	contentType := s.Request.Header.Get(http2.HeaderContentType)
	if contentType == http2.ContentTypeForm {
		err := s.ParseForm()
		if err != nil {
			return nil
		}
		return (mtos.KVsSource)(s.PostForm)
	}
	if contentType == http2.ContentTypeMultipart {
		err := s.ParseMultipartForm(DefaultMemory)
		if err != nil {
			return nil
		}
		return (*MultipartSource)(s.MultipartForm)
	}
	return nil
}

func (s RequestSource) BodyBind(obj any) error {
	if s.Method == http.MethodGet {
		return nil
	}
	data, err := io.ReadAll(s.Body)
	if err != nil {
		return fmt.Errorf("read body error: %w", err)
	}
	return BodyUnmarshaller(data, obj)
}
