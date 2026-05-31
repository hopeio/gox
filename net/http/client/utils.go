/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"fmt"
	urli "github.com/hopeio/gox/net/url"
	"net/http"
	"net/url"
	"os"
	"time"
)

func SetTag(t string) {
	urli.SetTag(t)
}

func SetProxyEnv(url string) {
	os.Setenv("HTTP_PROXY", url)
	os.Setenv("HTTPS_PROXY", url)
}

func setTimeout(client *http.Client, timeout time.Duration) {
	if client == nil {
		client = DefaultHttpClient
	}
	if timeout < time.Second {
		timeout = timeout * time.Second
	}
	client.Timeout = timeout
}

func ensureTransport(client *http.Client) *http.Transport {
	if t, ok := client.Transport.(*http.Transport); ok && t != nil {
		return t
	}
	t := apiTransport()
	client.Transport = t
	return t
}

func setProxy(client *http.Client, proxy func(*http.Request) (*url.URL, error)) {
	ensureTransport(client).Proxy = proxy
}

func closeResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}

func CloseReaderWrap(err error) error {
	return fmt.Errorf("close reader error: %w", err)
}
