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

// github.com/go-resty/resty 是个不错的选择,但是缺少一些我需要的功能，例如brotli解码，以及自定义处理body data，用于解决一些参数和返回body的AES加密或其他
// 不是并发安全的

var (
	DefaultHttpClient = newHttpClient(ClientTypeApi)
	DefaultLogLevel   = LogLevelError
)

const timeout = time.Minute

type ClientType uint8

const (
	ClientTypeApi      ClientType = iota
	ClientTypeDownload ClientType = iota
	ClientTypeUpload   ClientType = iota
)

func newHttpClient(typ ClientType) *http.Client {
	if typ == ClientTypeApi {
		return &http.Client{
			//Timeout: timeout * 2,
			Transport: &http.Transport{
				Proxy:             http.ProxyFromEnvironment, // 代理使用
				ForceAttemptHTTP2: true,
				DialContext: (&net.Dialer{
					Timeout:   timeout,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				//DisableKeepAlives: true,
				TLSHandshakeTimeout: timeout,
			},
		}
	}
	return newDownloadHttpClient()
}

// Client ...
type Client struct {
	typ ClientType
	// httpClient settings
	httpClient    *http.Client
	newHttpClient bool
	// request
	httpRequestOptions []HttpRequestOption
	header             http.Header //公共请求头
	reqDataMarshal     func(v any) ([]byte, error)

	// response
	responseHandler   func(response *http.Response) (retry bool, reader io.ReadCloser, err error)
	respDataHandler   func(data []byte) ([]byte, error)
	respDataUnMarshal func(data []byte, v any) error

	// logger
	logger   AccessLog
	logLevel LogLevel

	// retry
	retryTimes    int
	retryInterval time.Duration
	retryHandler  func(*http.Request)
}

func New() *Client {
	return &Client{httpClient: DefaultHttpClient, logger: DefaultLogger, logLevel: DefaultLogLevel, retryInterval: 200 * time.Millisecond}
}

func (d *Client) Header(header http.Header) *Client {
	if d.header == nil {
		d.header = make(http.Header)
	}
	httpx.CopyHttpHeader(d.header, header)
	return d
}

func (d *Client) AddHeader(k, v string) *Client {
	if d.header == nil {
		d.header = make(http.Header)
	}
	d.header.Add(k, v)
	return d
}

func (d *Client) Logger(logger AccessLog) *Client {
	if logger == nil {
		return d
	}
	d.logger = logger
	return d
}

func (d *Client) DisableLog() *Client {
	d.logLevel = LogLevelSilent
	return d
}

func (d *Client) LogLevel(lvl LogLevel) *Client {
	d.logLevel = lvl
	return d
}

// handler 返回值:是否重试,返回数据,错误
func (d *Client) ResponseHandler(handler func(response *http.Response) (retry bool, reader io.ReadCloser, err error)) *Client {
	d.responseHandler = handler
	return d
}

func (d *Client) RespDataHandler(handler func(data []byte) ([]byte, error)) *Client {
	d.respDataHandler = handler
	return d
}

func (d *Client) ReqDataMarshal(handler func(v any) ([]byte, error)) *Client {
	d.reqDataMarshal = handler
	return d
}

func (d *Client) RespDataUnMarshal(handler func(data []byte, v any) error) *Client {
	d.respDataUnMarshal = handler
	return d
}

func (d *Client) HttpRequestOptions(opts ...HttpRequestOption) *Downloader {
	d.httpRequestOptions = append(d.httpRequestOptions, opts...)
	return d
}

// 设置过期时间,仅对单次请求有效
func (d *Client) Timeout(timeout time.Duration) *Client {
	if !d.newHttpClient {
		d.httpClient = newHttpClient(d.typ)
		d.newHttpClient = true
	}
	setTimeout(d.httpClient, timeout)
	return d
}

func (d *Client) HttpClient(client *http.Client) *Client {
	d.httpClient = client
	d.newHttpClient = true
	return d
}

func (d *Client) SetHttpClient(opt HttpClientOption) *Client {
	if !d.newHttpClient {
		d.httpClient = newHttpClient(d.typ)
		d.newHttpClient = true
	}
	opt(d.httpClient)
	return d
}

func (d *Client) RetryTimes(retryTimes int) *Client {
	d.retryTimes = retryTimes
	return d
}

func (d *Client) RetryTimesWithInterval(retryTimes int, retryInterval time.Duration) *Client {
	d.retryTimes = retryTimes
	d.retryInterval = retryInterval
	return d
}

func (d *Client) RetryHandler(handle func(r *http.Request)) *Client {
	d.retryHandler = handle
	return d
}

func (d *Client) Proxy(proxyUrl string) *Client {
	if !d.newHttpClient {
		d.httpClient = newHttpClient(d.typ)
		d.newHttpClient = true
	}
	if proxyUrl != "" {
		purl, _ := stdurl.Parse(proxyUrl)
		setProxy(d.httpClient, http.ProxyURL(purl))
	}
	return d
}

func (d *Client) ResetProxy() *Client {
	if !d.newHttpClient {
		return d
	}
	d.httpClient.Transport.(*http.Transport).Proxy = http.ProxyFromEnvironment
	return d
}

func (d *Client) BasicAuth(authUser, authPass string) *Client {
	d.httpRequestOptions = append(d.httpRequestOptions, func(request *http.Request) {
		request.SetBasicAuth(authUser, authPass)
	})
	return d
}

func (d *Client) Clone() *Client {
	return &(*d)
}

func (d *Client) Request(method, url string) *Request {
	r := &Request{
		Method: method, Url: url, client: d,
	}
	return r
}

func (d *Client) Do(r *Request, param, response any) error {
	return r.Client(d).Do(param, response)
}

func (d *Client) Get(url string, param, response any) error {
	return NewRequest(http.MethodGet, url).Client(d).Do(param, response)
}

func (d *Client) GetRequest(url string) *Request {
	return NewRequest(http.MethodGet, url).Client(d)
}

func (d *Client) Post(url string, param, response any) error {
	return NewRequest(http.MethodPost, url).Client(d).Do(param, response)
}

func (d *Client) PostRequest(url string) *Request {
	return NewRequest(http.MethodPost, url).Client(d)
}

func (d *Client) Put(url string, param, response any) error {
	return NewRequest(http.MethodPut, url).Client(d).Do(param, response)
}

func (d *Client) PutRequest(url string) *Request {
	return NewRequest(http.MethodPut, url).Client(d)
}

func (d *Client) Delete(url string, param, response any) error {
	return NewRequest(http.MethodDelete, url).Client(d).Do(param, response)
}

func (d *Client) DeleteRequest(url string) *Request {
	return NewRequest(http.MethodDelete, url).Client(d)
}

func (d *Client) GetX(url string, response any) error {
	return NewRequest(http.MethodGet, url).Client(d).Do(nil, response)
}

func (d *Client) GetRaw(url string, param any) (RawBytes, error) {
	return NewRequest(http.MethodGet, url).Client(d).DoRaw(param)
}

func (d *Client) GetRawX(url string) (RawBytes, error) {
	return NewRequest(http.MethodGet, url).Client(d).DoRaw(nil)
}

func (d *Client) GetStream(url string, param any) (io.ReadCloser, error) {
	return NewRequest(http.MethodGet, url).Client(d).DoStream(param)
}

func (d *Client) GetStreamX(url string) (io.ReadCloser, error) {
	return NewRequest(http.MethodGet, url).Client(d).DoStream(nil)
}

type RawBytes = []byte
