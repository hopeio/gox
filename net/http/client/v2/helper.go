/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"github.com/hopeio/gox/net/http/client"
)

// GetRequest returns the value.
func GetRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.GetRequest(url))
}

// PostRequest returns the result.
func PostRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.PostRequest(url))
}

// PutRequest updates or inserts a value.
func PutRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.PutRequest(url))
}

// DeleteRequest removes or resets state.
func DeleteRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.DeleteRequest(url))
}

// Get returns the value.
func Get[RESP any](url string, param any) (*RESP, error) {
	return GetRequest[RESP](url).Do(param)
}

// Post performs the operation.
func Post[RESP any](url string, param any) (*RESP, error) {
	return PostRequest[RESP](url).Do(param)
}

// Put updates or inserts a value.
func Put[RESP any](url string, param any) (*RESP, error) {
	return PutRequest[RESP](url).Do(param)
}

// Delete removes or resets state.
func Delete[RESP any](url string, param any) (*RESP, error) {
	return DeleteRequest[RESP](url).Do(param)
}
