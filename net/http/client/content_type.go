/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package client

import (
	"strings"

	"github.com/hopeio/gox/net/http"
)

type ContentType uint8

func (c ContentType) String() string {
	if c < ContentTypeApplication {
		return contentTypes[c] + ";charset=UTF-8"
	}
	return http.ContentTypeOctetStream + ";charset=UTF-8"
}

func (c *ContentType) Decode(contentType string) {
	if strings.HasPrefix(contentType, http.ContentTypeJson) {
		*c = ContentTypeJson
	} else if strings.HasPrefix(contentType, http.ContentTypeForm) {
		*c = ContentTypeForm
	} else if strings.HasPrefix(contentType, "text") {
		*c = ContentTypeText
	} else if strings.HasPrefix(contentType, "image") {
		*c = ContentTypeImage
	} else if strings.HasPrefix(contentType, "video") {
		*c = ContentTypeVideo
	} else if strings.HasPrefix(contentType, "audio") {
		*c = ContentTypeAudio
	} else if strings.HasPrefix(contentType, "application") {
		*c = ContentTypeApplication
	} else {
		*c = ContentTypeJson
	}
}

const (
	ContentTypeUnset ContentType = iota
	ContentTypeJson
	ContentTypeForm
	ContentTypeFormData
	ContentTypeGrpc
	ContentTypeGrpcWeb
	ContentTypeXml
	ContentTypeText
	ContentTypeBinary
	ContentTypeApplication
	ContentTypeImage
	ContentTypeAudio
	ContentTypeVideo
	contentTypeUnSupport
)

var contentTypes = []string{
	http.ContentTypeJson,
	http.ContentTypeForm,
	http.ContentTypeMultipart,
	http.ContentTypeGrpc,
	http.ContentTypeGrpcWeb,
	http.ContentTypeXml,
	http.ContentTypeText,
	http.ContentTypeOctetStream,
	/*	consts.ContentImagePngHeaderValue,
		consts.ContentImageJpegHeaderValue,
		consts.ContentImageGifHeaderValue,
		consts.ContentImageBmpHeaderValue,
		consts.ContentImageWebpHeaderValue,
		consts.ContentImageAvifHeaderValue,
		consts.ContentImageTiffHeaderValue,
		consts.ContentImageXIconHeaderValue,
		consts.ContentImageVndMicrosoftIconHeaderValue,*/
}
