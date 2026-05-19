/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"time"
)

var DefaultUploader = NewUploader()

type Uploader = Client

func NewUploader() *Uploader {
	return &Uploader{
		typ:           ClientTypeUpload,
		httpClient:    DefaultDownloadHttpClient,
		retryTimes:    3,
		retryInterval: time.Second,
		logger:        nil,
		logLevel:      LogLevelSilent,
	}
}

func (d *Uploader) UploadReq(url string) *UploadReq {
	return NewUploadReq(url).Uploader(d)
}
