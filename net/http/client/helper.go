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

// DefaultHeader returns the result.
func DefaultHeader() http.Header {
	return http.Header{
		httpx.HeaderAcceptLanguage: []string{"zh-CN,zh;q=0.9;charset=utf-8"},
		httpx.HeaderConnection:     []string{"keep-alive"},
		httpx.HeaderUserAgent:      []string{UserAgentChrome117},
		//"Accept", "application/json,text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8", // Less common over time; servers usually use a fixed format
	}
}

// DefaultHeaderClient returns the result.
func DefaultHeaderClient() *Client {
	return New().Header(DefaultHeader())
}

// DefaultHeaderRequest returns the result.
func DefaultHeaderRequest() *Request {
	return &Request{client: New().Header(DefaultHeader())}
}

// GetRequest returns the value.
func GetRequest(url string) *Request {
	return NewRequest(http.MethodGet, url)
}

// PostRequest returns the result.
func PostRequest(url string) *Request {
	return NewRequest(http.MethodPost, url)
}

// PutRequest updates or inserts a value.
func PutRequest(url string) *Request {
	return NewRequest(http.MethodPut, url)
}

// DeleteRequest removes or resets state.
func DeleteRequest(url string) *Request {
	return NewRequest(http.MethodDelete, url)
}

// Get returns the value.
func Get(url string, param, response any) error {
	return GetRequest(url).Do(param, response)
}

// GetX returns the value.
func GetX(url string, response any) error {
	return Get(url, nil, response)
}

// GetStream returns the value.
func GetStream(url string, param any) (io.ReadCloser, error) {
	var resp *http.Response
	err := Get(url, param, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetStreamX returns the value.
func GetStreamX(url string) (io.ReadCloser, error) {
	return GetStream(url, nil)
}

// Post performs the operation.
func Post(url string, param, response interface{}) error {
	return PostRequest(url).Do(param, response)
}

// Put updates or inserts a value.
func Put(url string, param, response interface{}) error {
	return PutRequest(url).Do(param, response)
}

// Delete removes or resets state.
func Delete(url string, param, response interface{}) error {
	return DeleteRequest(url).Do(param, response)
}
