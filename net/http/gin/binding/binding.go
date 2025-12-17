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

	"github.com/hopeio/gox/kvstruct"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/binding"

	"github.com/gin-gonic/gin"
)

func Bind(ctx *gin.Context, obj any) error {
	return binding.CommonBind(RequestSource{ctx}, obj)
}

type RequestSource struct {
	*gin.Context
}

func (s RequestSource) Uri() kvstruct.Setter {
	return (uriSource)(s.Params)
}

func (s RequestSource) Query() kvstruct.Setter {
	return (kvstruct.KVsSource)(s.Request.URL.Query())
}

func (s RequestSource) Header() kvstruct.Setter {
	return (binding.HeaderSource)(s.Request.Header)
}

func (s RequestSource) Form() kvstruct.Setter {
	contentType := s.Request.Header.Get(httpx.HeaderContentType)
	if contentType == httpx.ContentTypeForm {
		err := s.Request.ParseForm()
		if err != nil {
			return nil
		}
		return (kvstruct.KVsSource)(s.Request.PostForm)
	}
	if contentType == httpx.ContentTypeMultipart {
		err := s.Request.ParseMultipartForm(binding.DefaultMemory)
		if err != nil {
			return nil
		}
		return (*binding.MultipartSource)(s.Request.MultipartForm)
	}
	return nil
}

func (s RequestSource) BodyBind(obj any) error {
	if s.Request.Method == http.MethodGet {
		return nil
	}
	data, err := io.ReadAll(s.Request.Body)
	if err != nil {
		return fmt.Errorf("read body error: %w", err)
	}
	if len(data) == 0 {
		return nil
	}
	return binding.BodyUnmarshaller(data, obj)
}
