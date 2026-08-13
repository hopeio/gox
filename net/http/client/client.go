/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"io"
	"net"
	"net/http"
	stdurl "net/url"
	"time"

	httpx "github.com/hopeio/gox/net/http"
)

var (
	DefaultHttpClient = newHttpClient(ClientTypeApi)
	DefaultLogLevel   = LogLevelError
)

const timeout = time.Minute

type ClientType uint8

const (
	ClientTypeApi ClientType = iota
	ClientTypeDownload
	ClientTypeUpload
)

// apiTransport returns the result.
func apiTransport() *http.Transport {
	return &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		ForceAttemptHTTP2: true,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: timeout,
		// 只限制等响应头阶段，body 流式读取不受影响，长下载安全；
		// 不用 http.Client.Timeout（整请求超时会杀长下载）
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: time.Second,
	}
}

// newHttpClient creates and returns a new instance.
func newHttpClient(typ ClientType) *http.Client {
	switch typ {
	case ClientTypeDownload, ClientTypeUpload:
		return newDownloadHttpClient()
	default:
		return &http.Client{Transport: apiTransport()}
	}
}

// Client defines a type.
type Client struct {
	typ     ClientType
	baseUrl string
	// httpClient settings
	httpClient    *http.Client
	newHttpClient bool
	// request
	httpRequestOptions []HttpRequestOption
	header             http.Header //shared request headers
	reqBodyMarshal     func(v any) ([]byte, error)

	// response
	responseHandler   func(response *http.Response) (retry bool, reader io.ReadCloser, err error)
	respBodyHandler   func(data []byte) ([]byte, error)
	respBodyUnMarshal func(data []byte, v any) error

	// logger
	logger   AccessLog
	logLevel LogLevel

	// retry
	retryTimes    int
	retryInterval time.Duration
	retryHandler  func(*http.Request)
}

// New creates a new instance.
func New() *Client {
	return &Client{httpClient: DefaultHttpClient, logger: DefaultLogger, logLevel: DefaultLogLevel, retryInterval: 200 * time.Millisecond}
}

// BaseUrl returns the result.
func (d *Client) BaseUrl(url string) *Client {
	d.baseUrl = url
	return d
}

// Header returns the result.
func (d *Client) Header(header http.Header) *Client {
	if d.header == nil {
		d.header = make(http.Header)
	}
	httpx.CopyHttpHeader(d.header, header)
	return d
}

// HeaderX returns the result.
func (d *Client) HeaderX(header httpx.Header) *Client {
	if d.header == nil {
		d.header = make(http.Header)
	}
	httpx.HeaderIntoHttpHeader(header, d.header)
	return d
}

// AddHeader updates or inserts a value.
func (d *Client) AddHeader(k, v string) *Client {
	if d.header == nil {
		d.header = make(http.Header)
	}
	d.header.Set(k, v)
	return d
}

// Logger returns the result.
func (d *Client) Logger(logger AccessLog) *Client {
	if logger == nil {
		return d
	}
	d.logger = logger
	return d
}

// DisableLog returns the result.
func (d *Client) DisableLog() *Client {
	d.logLevel = LogLevelSilent
	return d
}

// LogLevel returns the result.
func (d *Client) LogLevel(lvl LogLevel) *Client {
	d.logLevel = lvl
	return d
}

// ResponseHandler returns the result.
func (d *Client) ResponseHandler(handler func(response *http.Response) (retry bool, reader io.ReadCloser, err error)) *Client {
	d.responseHandler = handler
	return d
}

// RespBodyHandler returns the result.
func (d *Client) RespBodyHandler(handler func(data []byte) ([]byte, error)) *Client {
	d.respBodyHandler = handler
	return d
}

// ReqBodyMarshal returns the result.
func (d *Client) ReqBodyMarshal(handler func(v any) ([]byte, error)) *Client {
	d.reqBodyMarshal = handler
	return d
}

// RespBodyUnMarshal returns the result.
func (d *Client) RespBodyUnMarshal(handler func(data []byte, v any) error) *Client {
	d.respBodyUnMarshal = handler
	return d
}

// HttpRequestOptions returns the result.
func (d *Client) HttpRequestOptions(opts ...HttpRequestOption) *Client {
	d.httpRequestOptions = append(d.httpRequestOptions, opts...)
	return d
}

// Timeout returns the result.
func (d *Client) Timeout(timeout time.Duration) *Client {
	if !d.newHttpClient {
		d.httpClient = newHttpClient(d.typ)
		d.newHttpClient = true
	}
	setTimeout(d.httpClient, timeout)
	return d
}

// HttpClient returns the result.
func (d *Client) HttpClient(client *http.Client) *Client {
	d.httpClient = client
	d.newHttpClient = true
	return d
}

// SetHttpClient updates or inserts a value.
func (d *Client) SetHttpClient(opt HttpClientOption) *Client {
	if !d.newHttpClient {
		d.httpClient = newHttpClient(d.typ)
		d.newHttpClient = true
	}
	opt(d.httpClient)
	return d
}

// RetryTimes returns the result.
func (d *Client) RetryTimes(retryTimes int) *Client {
	d.retryTimes = retryTimes
	return d
}

// RetryTimesWithInterval returns the result.
func (d *Client) RetryTimesWithInterval(retryTimes int, retryInterval time.Duration) *Client {
	d.retryTimes = retryTimes
	d.retryInterval = retryInterval
	return d
}

// RetryHandler returns the result.
func (d *Client) RetryHandler(handle func(r *http.Request)) *Client {
	d.retryHandler = handle
	return d
}

// ensureOwnHttpClient performs the operation.
func (d *Client) ensureOwnHttpClient() {
	if !d.newHttpClient {
		d.httpClient = newHttpClient(d.typ)
		d.newHttpClient = true
	}
}

// Proxy returns the result.
func (d *Client) Proxy(proxyUrl string) *Client {
	d.ensureOwnHttpClient()
	if proxyUrl == "" {
		return d
	}
	purl, err := stdurl.Parse(proxyUrl)
	if err != nil {
		return d
	}
	setProxy(d.httpClient, http.ProxyURL(purl))
	return d
}

// NoProxy returns the result.
func (d *Client) NoProxy() *Client {
	d.ensureOwnHttpClient()
	setProxy(d.httpClient, nil)
	return d
}

// ResetProxy removes or resets state.
func (d *Client) ResetProxy() *Client {
	d.ensureOwnHttpClient()
	setProxy(d.httpClient, http.ProxyFromEnvironment)
	return d
}

// BasicAuth returns the result.
func (d *Client) BasicAuth(authUser, authPass string) *Client {
	d.httpRequestOptions = append(d.httpRequestOptions, func(request *http.Request) {
		request.SetBasicAuth(authUser, authPass)
	})
	return d
}

// Clone returns the result.
func (d *Client) Clone() *Client {
	c := *d
	if d.header != nil {
		c.header = d.header.Clone()
	}
	if d.httpRequestOptions != nil {
		c.httpRequestOptions = make([]HttpRequestOption, len(d.httpRequestOptions))
		copy(c.httpRequestOptions, d.httpRequestOptions)
	}
	return &c
}

// Request returns the result.
func (d *Client) Request(method, url string) *Request {
	r := &Request{
		Method: method, Url: url, client: d,
	}
	return r
}

// Do executes the operation.
func (d *Client) Do(r *Request, param, response any) error {
	return r.Client(d).Do(param, response)
}

// Get returns the value.
func (d *Client) Get(url string, param, response any) error {
	return NewRequest(http.MethodGet, url).Client(d).Do(param, response)
}

// GetRequest returns the value.
func (d *Client) GetRequest(url string) *Request {
	return NewRequest(http.MethodGet, url).Client(d)
}

// Post performs the operation.
func (d *Client) Post(url string, param, response any) error {
	return NewRequest(http.MethodPost, url).Client(d).Do(param, response)
}

// PostRequest returns the result.
func (d *Client) PostRequest(url string) *Request {
	return NewRequest(http.MethodPost, url).Client(d)
}

// Put updates or inserts a value.
func (d *Client) Put(url string, param, response any) error {
	return NewRequest(http.MethodPut, url).Client(d).Do(param, response)
}

// PutRequest updates or inserts a value.
func (d *Client) PutRequest(url string) *Request {
	return NewRequest(http.MethodPut, url).Client(d)
}

// Delete removes or resets state.
func (d *Client) Delete(url string, param, response any) error {
	return NewRequest(http.MethodDelete, url).Client(d).Do(param, response)
}

// DeleteRequest removes or resets state.
func (d *Client) DeleteRequest(url string) *Request {
	return NewRequest(http.MethodDelete, url).Client(d)
}

// GetX returns the value.
func (d *Client) GetX(url string, response any) error {
	return NewRequest(http.MethodGet, url).Client(d).Do(nil, response)
}

// GetRaw returns the value.
func (d *Client) GetRaw(url string, param any) (RawBytes, error) {
	return NewRequest(http.MethodGet, url).Client(d).DoRaw(param)
}

// GetRawX returns the value.
func (d *Client) GetRawX(url string) (RawBytes, error) {
	return NewRequest(http.MethodGet, url).Client(d).DoRaw(nil)
}

// GetStream returns the value.
func (d *Client) GetStream(url string, param any) (io.ReadCloser, error) {
	return NewRequest(http.MethodGet, url).Client(d).DoStream(param)
}

// GetStreamX returns the value.
func (d *Client) GetStreamX(url string) (io.ReadCloser, error) {
	return NewRequest(http.MethodGet, url).Client(d).DoStream(nil)
}

type RawBytes = []byte
