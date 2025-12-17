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
	"sync"

	jsonx "github.com/hopeio/gox/encoding/json"
	"github.com/hopeio/gox/kvstruct"
	httpx "github.com/hopeio/gox/net/http"
	stringsx "github.com/hopeio/gox/strings"
	"github.com/hopeio/gox/validator"
)

var (
	DefaultMemory    int64 = 32 << 20
	BodyUnmarshaller       = jsonx.Unmarshal
	CommonTag              = "json"
	Validate               = validator.ValidateStruct
	defaultTags            = []string{"uri", "path", "query", "header", "form", CommonTag}
)

type Source interface {
	Uri() kvstruct.Setter
	Query() kvstruct.Setter
	Header() kvstruct.Setter
	Form() kvstruct.Setter
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

func CommonBind(s Source, obj any) error {
	value := reflect.ValueOf(obj).Elem()
	typ := value.Type()
	err := s.BodyBind(obj)
	if err != nil {
		return err
	}
	uriSetter, querySetter, headerSetter, formSetter := s.Uri(), s.Query(), s.Header(), s.Form()
	commonSetter := kvstruct.Setters{Setters: []kvstruct.Setter{uriSetter, querySetter, headerSetter, formSetter}}
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
				case "form":
					setter = formSetter
				case "header":
					setter = headerSetter
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
		for _, tag = range defaultTags {
			tagValue = sf.Tag.Get(tag)
			var tags []Tag
			if tagValue != "" && tagValue != "-" {
				switch tag {
				case "uri", "path":
					setter = uriSetter
				case "query":
					setter = querySetter
				case "form":
					setter = formSetter
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
	contentType := s.Request.Header.Get(httpx.HeaderContentType)
	if contentType == httpx.ContentTypeForm {
		err := s.ParseForm()
		if err != nil {
			return nil
		}
		return (kvstruct.KVsSource)(s.PostForm)
	}
	if contentType == httpx.ContentTypeMultipart {
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
	contentType := s.Request.Header.Get(httpx.HeaderContentType)
	if contentType == httpx.ContentTypeForm || contentType == httpx.ContentTypeMultipart {
		return nil
	}
	data, err := io.ReadAll(s.Body)
	if err != nil {
		return fmt.Errorf("read body error: %w", err)
	}
	return BodyUnmarshaller(data, obj)
}
