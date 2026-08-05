/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	urli "github.com/hopeio/gox/net/url"
)

// SetTag updates or inserts a value.
func SetTag(t string) {
	urli.SetTag(t)
}

// SetProxyEnv updates or inserts a value.
func SetProxyEnv(url string) {
	os.Setenv("HTTP_PROXY", url)
	os.Setenv("HTTPS_PROXY", url)
}

// setTimeout performs the operation.
func setTimeout(client *http.Client, timeout time.Duration) {
	if client == nil {
		client = DefaultHttpClient
	}
	client.Timeout = timeout
}

// ensureTransport returns the result.
func ensureTransport(client *http.Client) *http.Transport {
	if t, ok := client.Transport.(*http.Transport); ok && t != nil {
		return t
	}
	t := apiTransport()
	client.Transport = t
	return t
}

// setProxy performs the operation.
func setProxy(client *http.Client, proxy func(*http.Request) (*url.URL, error)) {
	ensureTransport(client).Proxy = proxy
}

// closeResponse closes and releases resources.
func closeResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}

// CloseReaderWrap closes and releases resources.
func CloseReaderWrap(err error) error {
	return fmt.Errorf("close reader error: %w", err)
}
