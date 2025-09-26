/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hopeio/gox/log"
	httpx "github.com/hopeio/gox/net/http"
	urlx "github.com/hopeio/gox/net/url"
	"github.com/hopeio/gox/os/fs"
)

var DefaultDownloader = NewDownloader()

type DownloadMode uint16

const (
	DModeOverwrite DownloadMode = 1 << iota
	DModeForceContinue
)

const RangeFormat = "bytes=%d-%d/%d"
const RangeFormat2 = "bytes=%d-%d/*"

type DownloadReq struct {
	Url        string
	downloader *Downloader
	ctx        context.Context
	header     http.Header  //请求级请求头
	mode       DownloadMode // 模式，0-强制覆盖，1-不存在下载，2-断续下载
	rangeSize  int64
}

func NewDownloadReq(url string) *DownloadReq {
	return &DownloadReq{
		ctx:        context.Background(),
		Url:        url,
		downloader: DefaultDownloader,
	}
}

func (dReq *DownloadReq) Downloader(c *Downloader) *DownloadReq {
	dReq.downloader = c
	return dReq
}

func (dReq *DownloadReq) SetDownloader(set func(c *Downloader)) *DownloadReq {
	dReq.downloader = NewDownloader()
	set(dReq.downloader)
	return dReq
}
func (req *DownloadReq) Header(header httpx.Header) *DownloadReq {
	if req.header == nil {
		req.header = make(http.Header)
	}
	httpx.HeaderIntoHttpHeader(header, req.header)
	return req
}

func (dReq *DownloadReq) AddHeader(k, v string) *DownloadReq {
	if dReq.header == nil {
		dReq.header = make(http.Header)
	}
	dReq.header.Add(k, v)
	return dReq
}

func (dReq *DownloadReq) Mode(mode DownloadMode) *DownloadReq {
	dReq.mode = mode
	return dReq
}

func (dReq *DownloadReq) GetMode() DownloadMode {
	return dReq.mode
}

// 如果文件已存在，强制覆盖
func (dReq *DownloadReq) OverwriteMode() *DownloadReq {
	dReq.mode |= DModeOverwrite
	return dReq
}

func (dReq *DownloadReq) GetResponse(options ...func(*http.Request)) (*http.Response, error) {
	d := dReq.downloader
	req, err := http.NewRequestWithContext(dReq.ctx, http.MethodGet, dReq.Url, nil)
	if err != nil {
		return nil, err
	}

	// 如果自己设置了接受编码，http库不会自动gzip解压，需要自己处理，不加Accept-Encoding和Range头会自动设置gzip
	//req.Header.Set("Accept-Encoding", "gzip, deflate")
	if dReq.header != nil {
		req.Header = dReq.header
	}
	if _, ok := req.Header[httpx.HeaderAcceptLanguage]; !ok {
		req.Header.Set(httpx.HeaderAcceptLanguage, "zh-CN,zh;q=0.9;charset=utf-8")
	}
	if _, ok := req.Header[httpx.HeaderConnection]; !ok {
		req.Header.Set(httpx.HeaderConnection, "keep-alive")
	}
	if _, ok := req.Header[httpx.HeaderUserAgent]; !ok {
		req.Header.Set(httpx.HeaderUserAgent, UserAgentChrome117)
	}
	httpx.CopyHttpHeader(req.Header, d.header)
	for _, opt := range d.httpRequestOptions {
		opt(req)
	}
	for _, opt := range options {
		opt(req)
	}

	var resp *http.Response
	for i := range d.retryTimes {
		if i > 0 {
			time.Sleep(d.retryInterval)
		}
		resp, err = d.httpClient.Do(req)
		if err != nil {
			log.Warn(err, "url:", req.URL.Path)
			if strings.HasPrefix(err.Error(), "dial tcp: lookup") {
				return nil, err
			}
			continue
		} else {
			return resp, nil
		}
	}
	return nil, err
}

func (dReq *DownloadReq) GetReader() (io.ReadCloser, error) {
	_, reader, err := dReq.getReader()
	return reader, err
}

func (dReq *DownloadReq) getReader() (*http.Response, io.ReadCloser, error) {
Retry:
	resp, err := dReq.GetResponse()
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			return nil, nil, ErrNotFound
		}
		if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
			return nil, nil, ErrRangeNotSatisfiable
		}
		return nil, nil, fmt.Errorf("请求错误,status code:%d,url:%s", resp.StatusCode, dReq.Url)
	}

	d := dReq.downloader
	reader := resp.Body
	if d.responseHandler != nil {
		var retry bool
		retry, reader, err = d.responseHandler(resp)
		if retry {
			goto Retry
		}
		if err != nil {
			return nil, nil, err
		}
	}
	if d.respDataHandler != nil {
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, nil, err
		}
		data, err = d.respDataHandler(data)
		if err != nil {
			return nil, nil, err
		}
		resp.Body.Close()
		reader = io.NopCloser(bytes.NewBuffer(data))
	}
	return resp, reader, nil
}

func (dReq *DownloadReq) Download(filepath string) error {
	if dReq.mode&DModeOverwrite == 0 && fs.Exist(filepath) {
		return nil
	}
	if dReq.downloader.retryTimes == 0 {
		dReq.downloader.retryTimes = 1
	}
	if dReq.mode&DModeForceContinue != 0 {
		return dReq.continuationDownload(filepath)
	}
	var reader io.ReadCloser
	var err error
	var resp *http.Response
	var notContinuation bool
	for range dReq.downloader.retryTimes {
		resp, reader, err = dReq.getReader()
		if err != nil {
			return err
		}
		if !notContinuation && resp.Header.Get(httpx.HeaderAcceptRanges) == "bytes" {
			length := httpx.GetContentLength(resp.Header)
			if length > defaultSize {
				reader.Close()
				return dReq.continuationDownload(filepath)
			}
			notContinuation = true
		}
		err = fs.Download(filepath, reader)
		reader.Close()
		if err == nil {
			return nil
		}
		log.Warn(err, dReq.Url, filepath)
	}
	return err
}

func (dReq *DownloadReq) DownloadAttachment(dir string) error {

	if dReq.downloader.retryTimes == 0 {
		dReq.downloader.retryTimes = 1
	}
	var reader io.ReadCloser
	var err error
	var resp *http.Response
	filepath := dir + fs.PathSeparator + path.Base(dReq.Url)
	first := true
	for range dReq.downloader.retryTimes {
		resp, reader, err = dReq.getReader()
		if err != nil {
			return err
		}

		if first {
			disposition, err := httpx.ParseContentDisposition(resp.Header.Get(httpx.HeaderContentDisposition))
			if err != nil {
				return err
			}
			filepath = dir + fs.PathSeparator + disposition
			if dReq.mode&DModeOverwrite == 0 && fs.Exist(filepath) {
				return nil
			}
			if resp.Header.Get(httpx.HeaderAcceptRanges) == "bytes" {
				length := httpx.GetContentLength(resp.Header)
				if length > defaultSize {
					reader.Close()
					return dReq.continuationDownload(filepath)
				}
			}
			first = false
		}
		if dReq.mode&DModeForceContinue != 0 {
			reader.Close()
			return dReq.continuationDownload(filepath)
		}
		err = fs.Download(filepath, reader)
		reader.Close()
		if err == nil {
			return nil
		}
		log.Warn(err, dReq.Url, filepath)
	}
	return err
}

func (dReq *DownloadReq) continuationDownload(filepath string) error {
	f, err := fs.OpenFile(filepath+DownloadKey, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	fileinfo, err := f.Stat()
	if err != nil {
		return err
	}

	offset := fileinfo.Size()
	var reader io.ReadCloser
	for range dReq.downloader.retryTimes {
		dReq.header.Set(httpx.HeaderRange, httpx.FormatRange(offset, 0))

		reader, err = dReq.GetReader()
		if err != nil {
			if errors.Is(err, ErrRangeNotSatisfiable) {
				f.Close()
				return os.Rename(filepath+DownloadKey, filepath)
			}
			continue
		}

		var written int64
		written, err = io.Copy(f, reader)
		reader.Close()

		if err == nil || err == io.EOF {
			f.Close()
			return os.Rename(filepath+DownloadKey, filepath)
		}

		offset += written
	}
	f.Close()
	return err
}

const defaultRange = "bytes=0-"
const defaultSize = 30 * 1024 * 1024

// TODO: 利用简单任务调度实现
func (dReq *DownloadReq) ConcurrencyDownload(filepath string, url string, concurrencyNum int) error {
	if dReq.mode&DModeOverwrite == 0 && fs.Exist(filepath) {
		return nil
	}
	panic("TODO")
	return nil
}

func GetReader(url string) (io.ReadCloser, error) {
	return GetReaderWithHttpRequestOptions(url)
}

func GetReaderWithHttpRequestOptions(url string, opts ...HttpRequestOption) (io.ReadCloser, error) {
	resp, err := NewDownloader().HttpRequestOptions(opts...).DownloadReq(url).GetResponse()
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func Download(filepath, url string) error {
	return NewDownloadReq(url).Download(filepath)
}

func GetImage(url string) (io.ReadCloser, error) {
	return GetReaderWithHttpRequestOptions(url, ImageOption)
}

func DownloadImage(filepath, url string) error {
	reader, err := GetReaderWithHttpRequestOptions(url, ImageOption)
	if err != nil {
		return err
	}
	return fs.Download(filepath, reader)
}

func ImageOption(req *http.Request) {
	req.Header.Set(httpx.HeaderAccept, "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
}

func DownloadToDir(dir, url string) error {
	return NewDownloadReq(url).Download(dir + fs.PathSeparator + urlx.URIBase(url))
}
