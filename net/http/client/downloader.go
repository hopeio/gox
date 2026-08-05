/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"github.com/hopeio/gox/os/fs"
	"net/http"
	"time"
)

var DefaultDownloadHttpClient = newDownloadHttpClient()

// newDownloadHttpClient creates and returns a new instance.
func newDownloadHttpClient() *http.Client {
	return &http.Client{Transport: apiTransport()}
}

// TODO: Range Status(206) PartialContent 下载
type Downloader = Client

// NewDownloader creates and returns a new instance.
func NewDownloader(options ...Option) *Downloader {
	downloader := &Downloader{
		typ:           ClientTypeDownload,
		httpClient:    DefaultDownloadHttpClient,
		retryTimes:    3,
		retryInterval: time.Second,
		logger:        nil,
		logLevel:      LogLevelSilent,
	}
	for _, opt := range options {
		opt(downloader)
	}
	return downloader
}

// Download ...
func (d *Downloader) Download(filepath string, r *DownloadReq) error {
	return r.Downloader(d).Download(filepath)
}

// DownloadAttachment ...
func (d *Downloader) DownloadAttachment(dir string, r *DownloadReq) {
	r.Downloader(d).DownloadAttachment(dir)
}

// DownloadReq ...
func (d *Downloader) DownloadReq(url string) *DownloadReq {
	return NewDownloadReq(url).Downloader(d)
}

const DownloadKey = fs.DownloadKey
