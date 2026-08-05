/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"github.com/hopeio/gox/net/http/client"
)

// GetRequest ...
func GetRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.GetRequest(url))
}

// PostRequest ...
func PostRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.PostRequest(url))
}

// PutRequest ...
func PutRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.PutRequest(url))
}

// DeleteRequest ...
func DeleteRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.DeleteRequest(url))
}

// Get ...
func Get[RESP any](url string, param any) (*RESP, error) {
	return GetRequest[RESP](url).Do(param)
}

// Post ...
func Post[RESP any](url string, param any) (*RESP, error) {
	return PostRequest[RESP](url).Do(param)
}

// Put ...
func Put[RESP any](url string, param any) (*RESP, error) {
	return PutRequest[RESP](url).Do(param)
}

// Delete ...
func Delete[RESP any](url string, param any) (*RESP, error) {
	return DeleteRequest[RESP](url).Do(param)
}
