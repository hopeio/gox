package http

const (
	HeaderUserAgent                   = "User-Agent"
	HeaderXForwardedFor               = "X-Forwarded-For"
	HeaderXAccelBuffering             = "X-Accel-Buffering"
	HeaderAuth                        = "Auth"
	HeaderContentType                 = "Content-Type"
	HeaderTrace                       = "Tracing"
	HeaderTraceID                     = "Tracing-ID"
	HeaderTraceBin                    = "Tracing-Bin"
	HeaderAuthorization               = "Authorization"
	HeaderCookie                      = "Cookie"
	HeaderCookieValueToken            = "token"
	HeaderCookieValueDel              = "del"
	HeaderContentDisposition          = "Content-Disposition"
	HeaderContentEncoding             = "Content-Encoding"
	HeaderReferer                     = "Referer"
	HeaderAccept                      = "Accept"
	HeaderAcceptCharset               = "Accept-Charset"
	HeaderAcceptLanguage              = "Accept-Language"
	HeaderAcceptEncoding              = "Accept-Encoding"
	HeaderCacheControl                = "Cache-Control"
	HeaderSetCookie                   = "Set-Cookie"
	HeaderTrailer                     = "Trailer"
	HeaderTransferEncoding            = "Transfer-Encoding"
	HeaderTransferEncodingChunked     = "chunked"
	HeaderTE                          = "TE"
	HeaderLastModified                = "Last-Modified"
	HeaderContentLength               = "Content-Length"
	HeaderAccessControlRequestMethod  = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders = "Access-Control-Request-Headers"
	HeaderOrigin                      = "Origin"
	HeaderConnection                  = "Connection"
	HeaderRange                       = "Range"
	HeaderHost                        = "Host"
	HeaderVia                         = "Via"
	HeaderDate                        = "Date"
	HeaderExpect                      = "Expect"
	HeaderFrom                        = "From"
	HeaderPragma                      = "Pragma"
	HeaderWarning                     = "Warning"
	HeaderContentRange                = "Content-Range"
	HeaderAcceptRanges                = "Accept-Ranges"
	HeaderXForwardedHost              = "X-Forwarded-Host"
)

const (
	HeaderGrpcTraceBin = "Grpc-Trace-Bin"
	HeaderGrpcInternal = "Grpc-Internal"
	HeaderGrpcStatus   = "Grpc-Status"
	HeaderGrpcMsg      = "Grpc-Msg"
	HeaderGrpcMessage  = "Grpc-Message"
	HeaderErrorCode    = "Error-Code"
	HeaderErrorMsg     = "Error-Msg"
	HeaderErrorMessage = "Error-Message"
)

const (
	HeaderDeviceInfo = "Device-Info"
	HeaderAppInfo    = "App-Info"
	HeaderLocation   = "Location"
	HeaderArea       = "Area"
	HeaderInternal   = "Internal"
)

const (
	// ContentTypeHtml is the  string of text/html response header's content type value.
	ContentTypeHtml = "text/html"
	ContentTypeCss  = "text/css"
	// ContentTypeText header value for Text data.
	ContentTypeText = "text/plain"
	// ContentTypeTextXml header value for XML data.
	ContentTypeTextXml = "text/xml"
	// ContentTypeTextMarkdown custom key/content type, the real is the text/html.
	ContentTypeTextMarkdown = "text/markdown"
	// ContentTypeTextYaml header value for YAML plain text.
	ContentTypeTextYaml = "text/yaml"

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
	ContentTypeXmlProblem = "application/problem+xml"
	ContentTypeJavascript = "application/javascript"
	// ContentTypeXml obsolete header value for XML.
	ContentTypeXml = "application/xml"
	// ContentTypeYaml header value for YAML data.
	ContentTypeYaml = "application/yaml"
	// ContentTypeProtobuf header value for Protobuf messages data.
	ContentTypeProtobuf = "application/protobuf"
	// ContentTypeMsgPack header value for MsgPack data.
	ContentTypeMsgPack = "application/msgpack"
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
