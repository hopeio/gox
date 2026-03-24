/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"github.com/hopeio/gox/net/http/client"
)

func GetRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.GetRequest(url))
}

func PostRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.PostRequest(url))
}

func PutRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.PutRequest(url))
}

func DeleteRequest[RESP any](url string) *Request[RESP] {
	return (*Request[RESP])(client.DeleteRequest(url))
}

func Get[RESP any](url string, param any) (*RESP, error) {
	return GetRequest[RESP](url).Do(param)
}

func Post[RESP any](url string, param any) (*RESP, error) {
	return PostRequest[RESP](url).Do(param)
}

func Put[RESP any](url string, param any) (*RESP, error) {
	return PutRequest[RESP](url).Do(param)
}

func Delete[RESP any](url string, param any) (*RESP, error) {
	return DeleteRequest[RESP](url).Do(param)
}
