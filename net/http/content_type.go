/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package http

const (
	// ContentTypeJavascript header value for JSONP & Javascript data.
	ContentTypeJavascript = "text/javascript"
	// ContentTypeHtml is the  string of text/html response header's content type value.
	ContentTypeHtml = "text/html"
	ContentTypeCss  = "text/css"
	// ContentTypeText header value for Text data.
	ContentTypeText = "text/plain"
	// ContentTypeXml header value for XML data.
	ContentTypeXml = "text/xml"
	// ContentTypeMarkdown custom key/content type, the real is the text/html.
	ContentTypeMarkdown = "text/markdown"
	// ContentTypeYamlText header value for YAML plain text.
	ContentTypeYamlText = "text/yaml"

	// ContentTypeMultipart header value for post multipart form data.
	ContentTypeMultipart = "multipart/form-data"

	// ContentTypeOctetStream header value for binary data.
	ContentTypeOctetStream = "application/octet-stream"
	// ContentTypeWebassembly header value for web assembly files.
	ContentTypeWebassembly = "application/wasm"
	// ContentTypeJson header value for JSON data.
	ContentTypeJson = "application/json"
	// ContentTypeJsonProblem header value for JSON API problem error.
	// Read more at: https://tools.ietf.org/html/rfc7807
	ContentTypeJsonProblem = "application/problem+json"
	// ContentTypeXmlProblem header value for XML API problem error.
	// Read more at: https://tools.ietf.org/html/rfc7807
	ContentTypeXmlProblem           = "application/problem+xml"
	ContentTypeJavascriptUnreadable = "application/javascript"
	// ContentTypeXmlUnreadable obsolete header value for XML.
	ContentTypeXmlUnreadable = "application/xml"
	// ContentTypeYaml header value for YAML data.
	ContentTypeYaml = "application/x-yaml"
	// ContentTypeProtobuf header value for Protobuf messages data.
	ContentTypeProtobuf = "application/x-protobuf"
	// ContentTypeMsgPack header value for MsgPack data.
	ContentTypeMsgPack = "application/msgpack"
	// ContentTypeMsgPack2 alternative header value for MsgPack data.
	ContentTypeMsgPack2 = "application/x-msgpack"
	// ContentTypeForm header value for post form data.
	ContentTypeForm = "application/x-www-form-urlencoded"

	// ContentTypeGrpc Content-Type header value for gRPC.
	ContentTypeGrpc      = "application/grpc"
	ContentTypeGrpcWeb   = "application/grpc-web"
	ContentTypePdf       = "application/pdf"
	ContentTypeJsonUtf8  = "application/json;charset=utf-8"
	ContentTypeFormParam = "application/x-www-form-urlencoded;param=value"

	ContentTypeImagePng              = "image/png"
	ContentTypeImageJpeg             = "image/jpeg"
	ContentTypeImageGif              = "image/gif"
	ContentTypeImageBmp              = "image/bmp"
	ContentTypeImageWebp             = "image/webp"
	ContentTypeImageAvif             = "image/avif"
	ContentTypeImageHeif             = "image/heif"
	ContentTypeImageSvg              = "image/svg+xml"
	ContentTypeImageTiff             = "image/tiff"
	ContentTypeImageXIcon            = "image/x-icon"
	ContentTypeImageVndMicrosoftIcon = "image/vnd.microsoft.icon"

	ContentTypeCharsetUtf8 = "charset=UTF-8"
)

const (
	FormDataFieldTmpl = `form-data; name="%s"`
	FormDataFileTmpl  = `form-data; name="%s"; filename="%s"`
	AttachmentTmpl    = `attachment; filename="%s"`
)
