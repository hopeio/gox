/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"io"
	"net/http"

	httpx "github.com/hopeio/gox/net/http"
)

// DefaultHeader ...
func DefaultHeader() http.Header {
	return http.Header{
		httpx.HeaderAcceptLanguage: []string{"zh-CN,zh;q=0.9;charset=utf-8"},
		httpx.HeaderConnection:     []string{"keep-alive"},
		httpx.HeaderUserAgent:      []string{UserAgentChrome117},
		//"Accept", "application/json,text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8", // 将会越来越少用，服务端一般固定格式
	}
}

// DefaultHeaderClient ...
func DefaultHeaderClient() *Client {
	return New().Header(DefaultHeader())
}

// DefaultHeaderRequest ...
func DefaultHeaderRequest() *Request {
	return &Request{client: New().Header(DefaultHeader())}
}

// GetRequest ...
func GetRequest(url string) *Request {
	return NewRequest(http.MethodGet, url)
}

// PostRequest ...
func PostRequest(url string) *Request {
	return NewRequest(http.MethodPost, url)
}

// PutRequest ...
func PutRequest(url string) *Request {
	return NewRequest(http.MethodPut, url)
}

// DeleteRequest ...
func DeleteRequest(url string) *Request {
	return NewRequest(http.MethodDelete, url)
}

// Get ...
func Get(url string, param, response any) error {
	return GetRequest(url).Do(param, response)
}

// GetX ...
func GetX(url string, response any) error {
	return Get(url, nil, response)
}

// GetStream ...
func GetStream(url string, param any) (io.ReadCloser, error) {
	var resp *http.Response
	err := Get(url, param, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetStreamX ...
func GetStreamX(url string) (io.ReadCloser, error) {
	return GetStream(url, nil)
}

// Post ...
func Post(url string, param, response interface{}) error {
	return PostRequest(url).Do(param, response)
}

// Put ...
func Put(url string, param, response interface{}) error {
	return PutRequest(url).Do(param, response)
}

// Delete ...
func Delete(url string, param, response interface{}) error {
	return DeleteRequest(url).Do(param, response)
}
