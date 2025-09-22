/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"context"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/client"
)

// Client ...

type Request[RES any] client.Request

func NewRequest[RES any](method, url string) *Request[RES] {
	return &Request[RES]{Method: method, Url: url}
}

func NewRequestFromV1[RES any](req *client.Request) *Request[RES] {
	return (*Request[RES])(req)
}

func (req *Request[RES]) Client(client2 *client.Client) *Request[RES] {
	(*client.Request)(req).Client(client2)
	return req
}

func (req *Request[RES]) Origin() *client.Request {
	return (*client.Request)(req)
}

func (req *Request[RES]) Header(header httpx.Header) *Request[RES] {
	(*client.Request)(req).Header(header)
	return req
}
func (req *Request[RES]) AddHeader(k, v string) *Request[RES] {
	(*client.Request)(req).AddHeader(k, v)
	return req
}

func (req *Request[RES]) ContentType(contentType client.ContentType) *Request[RES] {
	(*client.Request)(req).ContentType(contentType)
	return req
}

func (req *Request[RES]) Context(ctx context.Context) *Request[RES] {
	(*client.Request)(req).Context(ctx)
	return req
}

func (req *Request[RES]) DoNoParam() (*RES, error) {
	response := new(RES)
	return response, (*client.Request)(req).Do(nil, response)
}

// Do create a HTTP request
func (req *Request[RES]) Do(param any) (*RES, error) {
	response := new(RES)
	err := (*client.Request)(req).Do(param, response)
	return response, err
}
