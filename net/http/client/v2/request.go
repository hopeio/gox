/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"context"
	"net/http"

	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/client"
)

// Client ...

type Request[RESP any] client.Request

// NewRequest creates and returns a new instance.
func NewRequest[RESP any](method, url string) *Request[RESP] {
	return &Request[RESP]{Method: method, Url: url}
}

// NewRequestFromV1 creates and returns a new instance.
func NewRequestFromV1[RESP any](req *client.Request) *Request[RESP] {
	return (*Request[RESP])(req)
}

// Client ...
func (req *Request[RESP]) Client(client2 *client.Client) *Request[RESP] {
	(*client.Request)(req).Client(client2)
	return req
}

// Origin ...
func (req *Request[RESP]) Origin() *client.Request {
	return (*client.Request)(req)
}

// Header ...
func (req *Request[RESP]) Header(header http.Header) *Request[RESP] {
	(*client.Request)(req).Header(header)
	return req
}

// HeaderX ...
func (req *Request[RESP]) HeaderX(header httpx.Header) *Request[RESP] {
	(*client.Request)(req).HeaderX(header)
	return req
}

// AddHeader ...
func (req *Request[RESP]) AddHeader(k, v string) *Request[RESP] {
	(*client.Request)(req).AddHeader(k, v)
	return req
}

// ContentType ...
func (req *Request[RESP]) ContentType(contentType client.ContentType) *Request[RESP] {
	(*client.Request)(req).ContentType(contentType)
	return req
}

// Context ...
func (req *Request[RESP]) Context(ctx context.Context) *Request[RESP] {
	(*client.Request)(req).Context(ctx)
	return req
}

// Do create a HTTP request
func (req *Request[RESP]) Do(param any) (*RESP, error) {
	response := new(RESP)
	err := (*client.Request)(req).Do(param, response)
	return response, err
}
