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

func DefaultHeader() httpx.MapHeader {
	return httpx.MapHeader{
		httpx.HeaderAcceptLanguage: "zh-CN,zh;q=0.9;charset=utf-8",
		httpx.HeaderConnection:     "keep-alive",
		httpx.HeaderUserAgent:      UserAgentChrome117,
		//"Accept", "application/json,text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8", // 将会越来越少用，服务端一般固定格式
	}
}

func DefaultHeaderClient() *Client {
	return New().Header(DefaultHeader())
}

func DefaultHeaderRequest() *Request {
	return &Request{client: New().Header(DefaultHeader())}
}

func GetRequest(url string) *Request {
	return NewRequest(http.MethodGet, url)
}

func PostRequest(url string) *Request {
	return NewRequest(http.MethodPost, url)
}

func PutRequest(url string) *Request {
	return NewRequest(http.MethodPut, url)
}

func DeleteRequest(url string) *Request {
	return NewRequest(http.MethodDelete, url)
}

func Get(url string, param, response any) error {
	return GetRequest(url).Do(param, response)
}

func GetX(url string, response any) error {
	return Get(url, nil, response)
}

func GetStream(url string, param any) (io.ReadCloser, error) {
	var resp *http.Response
	err := Get(url, param, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func GetStreamX(url string) (io.ReadCloser, error) {
	return GetStream(url, nil)
}

func Post(url string, param, response interface{}) error {
	return PostRequest(url).Do(param, response)
}

func Put(url string, param, response interface{}) error {
	return PutRequest(url).Do(param, response)
}

func Delete(url string, param, response interface{}) error {
	return DeleteRequest(url).Do(param, response)
}
