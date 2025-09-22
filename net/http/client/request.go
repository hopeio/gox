/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andybalholm/brotli"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/consts"
	url2 "github.com/hopeio/gox/net/url"
	stringsx "github.com/hopeio/gox/strings"
	"github.com/hopeio/gox/strings/unicode"
	"github.com/klauspost/compress/zstd"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	DefaultClient = New().DisableLog()
	bufPool       = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

type Request struct {
	ctx         context.Context
	Method, Url string
	contentType ContentType
	header      http.Header //请求级请求头
	client      *Client
}

func NewRequest(method, url string, opts ...RequestOption) *Request {
	r := &Request{
		ctx:    context.Background(),
		Method: method,
		Url:    url,
		client: DefaultClient,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (req *Request) Client(c *Client) *Request {
	req.client = c
	return req
}

func (req *Request) Header(header httpx.Header) *Request {
	if req.header == nil {
		req.header = make(http.Header)
	}
	httpx.HeaderIntoHttpHeader(header, req.header)
	return req
}

func (req *Request) AddHeader(k, v string) *Request {
	if req.header == nil {
		req.header = make(http.Header)
	}
	req.header.Add(k, v)
	return req
}

func (req *Request) ContentType(contentType ContentType) *Request {
	req.contentType = contentType
	return req
}

func (req *Request) Context(ctx context.Context) *Request {
	req.ctx = ctx
	return req
}

func (req *Request) DoEmpty() error {
	return req.Do(nil, nil)
}

func (req *Request) DoNoParam(response any) error {
	return req.Do(nil, response)
}

func (req *Request) DoNoResponse(param any) error {
	return req.Do(param, nil)
}

func (req *Request) DoRaw(param any) (RawBytes, error) {
	var raw RawBytes
	err := req.Do(param, &raw)
	if err != nil {
		return raw, err
	}
	return raw, nil
}

func (req *Request) DoStream(param any) (io.ReadCloser, error) {
	var resp *http.Response
	err := req.Do(param, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// Do create a HTTP request
// param: 请求参数 目前只支持编码为json 或 url-encoded
func (req *Request) Do(param, response any) error {
	if req.Method == "" {
		return errors.New("not set method")
	}

	if req.Url == "" {
		return errors.New("not set url")
	}
	if req.ctx == nil {
		req.ctx = context.Background()
	}
	if req.client == nil {
		req.client = DefaultClient
	}
	c := req.client

	var accessLogParam AccessLogParam
	var reqBody, respBody []byte
	var reqTimes int
	var err error
	reqTime := time.Now()
	var request *http.Request
	var resp *http.Response
	// 日志记录
	defer func(now time.Time) {
		if c.logLevel == LogLevelInfo || (err != nil && c.logLevel == LogLevelError) {
			accessLogParam.Duration = time.Since(reqTime)
			c.logger(&AccessLogParam{
				Method:   req.Method,
				Url:      req.Url,
				Duration: time.Since(reqTime),
				ReqBody:  reqBody,
				RespBody: respBody,
				Request:  request,
				Response: resp,
			}, err)
		}
	}(reqTime)
	var body io.Reader
	if req.Method == http.MethodGet {
		req.Url = url2.AppendQueryParam(req.Url, param)
	} else {
		if param != nil {
			switch paramType := param.(type) {
			case string:
				reqBody = stringsx.ToBytes(paramType)
			case []byte:
				reqBody = paramType
			case io.Reader:
				if c.logLevel == LogLevelSilent {
					body = paramType
				} else {
					reqBody, err = io.ReadAll(paramType)
				}
			default:
				if c.customReqMarshal != nil {
					reqBody, err = c.customReqMarshal(param)
					if err != nil {
						return err
					}
				} else {
					switch req.contentType {
					case ContentTypeForm:
						params := url2.QueryParam(param)
						reqBody = stringsx.ToBytes(params)
					default:
						reqBody, err = json.Marshal(param)
						if err != nil {
							return err
						}
					}
				}

			}
		}

		if len(reqBody) > 0 {
			if c.customReqMarshal != nil {
				reqBody, err = c.customReqMarshal(reqBody)
			}
			body = bytes.NewReader(reqBody)
		}
	}

	request, err = http.NewRequestWithContext(req.ctx, req.Method, req.Url, body)
	if err != nil {
		return err
	}
	if req.header != nil {
		request.Header = req.header
	}
	request.Header.Set(consts.HeaderContentType, req.contentType.String())
	httpx.CopyHttpHeader(request.Header, c.header)

Retry:
	if reqTimes > 0 {
		if c.retryInterval != 0 {
			time.Sleep(c.retryInterval)
		}
		reqTime = time.Now()
		if reqBody != nil {
			request.Body = io.NopCloser(bytes.NewReader(reqBody))
		}
		if c.retryHandler != nil {
			c.retryHandler(request)
		}
	}
	resp, err = c.httpClient.Do(request)
	reqTimes++
	if err != nil {
		if c.retryTimes == 0 || reqTimes == c.retryTimes {
			return err
		} else {
			if c.logLevel > LogLevelSilent {
				c.logger(&AccessLogParam{
					Method:   req.Method,
					Url:      req.Url,
					Duration: time.Since(reqTime),
					ReqBody:  reqBody,
					RespBody: respBody,
					Request:  request,
					Response: resp,
				}, errors.New(err.Error()+";will retry"))
			}
			goto Retry
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		if resp.StatusCode == http.StatusNotFound {
			err = ErrNotFound
		} else {
			respBody, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return err
			}
			err = errors.New("status:" + resp.Status + " " + unicode.ToUtf8(respBody))
		}
		return err
	}

	if httpresp, ok := response.(*http.Response); ok {
		*httpresp = *resp
		return err
	}

	if httpresp, ok := response.(**http.Response); ok {
		*httpresp = resp
		return err
	}

	var reader io.Reader
	// net/http会自动处理gzip
	// go1.22 发现没有处理(并不是,是请求时header标明Content-Encoding时不会处理)
	encoding := resp.Header.Get(consts.HeaderContentEncoding)
	var compress bool
	if encoding != "" {
		switch strings.ToLower(encoding) {
		case "gzip":
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				resp.Body.Close()
				return err
			}
			compress = true
		case "br":
			reader = brotli.NewReader(resp.Body)
			compress = true
		case "deflate":
			reader = flate.NewReader(resp.Body)
			compress = true
		case "zstd":
			reader, err = zstd.NewReader(resp.Body)
			if err != nil {
				resp.Body.Close()
				return err
			}
			compress = true
		default:
			reader = resp.Body
		}
	} else {
		reader = resp.Body
	}
	if compress {
		resp.Header.Del(consts.HeaderContentEncoding)
		resp.Header.Del(consts.HeaderContentLength)
		resp.ContentLength = -1
		resp.Uncompressed = true
	}

	if httpresp, ok := response.(*io.Reader); ok {
		*httpresp = reader
		return err
	}

	if c.responseHandler != nil {
		var retry bool
		retry, reader, err = c.responseHandler(resp)
		if retry {
			if c.logLevel > LogLevelSilent {
				c.logger(&AccessLogParam{
					Method:   req.Method,
					Url:      req.Url,
					Duration: time.Since(reqTime),
					ReqBody:  reqBody,
					RespBody: respBody,
					Request:  request,
					Response: resp,
				}, err)
			}
			goto Retry
		} else if err != nil {
			return err
		}
	}

	respBody, err = io.ReadAll(reader)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if len(respBody) > 0 && response != nil {
		//contentType := resp.Header.Get(consts.HeaderContentType)

		if raw, ok := response.(*RawBytes); ok {
			*raw = respBody
			return nil
		}
		if req.client.customResUnMarshal != nil {
			err = req.client.customResUnMarshal(respBody, response)
			if err != nil {
				return fmt.Errorf("json.Unmarshal error: %w", err)
			}
		} else {
			// 默认json
			err = json.Unmarshal(respBody, response)
			if err != nil {
				return fmt.Errorf("json.Unmarshal error: %w", err)
			}
		}
		if v, ok := response.(ResponseBodyCheck); ok {
			err = v.CheckError()
		}
	}

	return err
}
