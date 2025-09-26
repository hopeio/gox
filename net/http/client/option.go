/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"net/http"

	httpx "github.com/hopeio/gox/net/http"
)

type Option func(req *Client)

type HttpRequestOption func(req *http.Request)

func (o HttpRequestOption) ToOption() Option {
	return func(c *Client) {
		c.HttpRequestOptions(o)
	}
}

func AddHeader(k, v string) HttpRequestOption {
	return func(req *http.Request) {
		req.Header.Add(k, v)
	}
}

func SetRefer(refer string) HttpRequestOption {
	return func(req *http.Request) {
		req.Header.Set(httpx.HeaderReferer, refer)
	}
}

func SetAccept(refer string) HttpRequestOption {
	return func(req *http.Request) {
		req.Header.Set(httpx.HeaderAccept, refer)
	}
}

func SetCookie(cookie string) HttpRequestOption {
	return func(req *http.Request) {
		req.Header.Set(httpx.HeaderCookie, cookie)
	}
}

// TODO
// tag :`request:"uri:xxx;query:xxx;header:xxx;body:xxx"`
func setRequest(p any, req *http.Request) {

}

type HttpClientOption func(client *http.Client)

type RequestOption func(req *Request)
