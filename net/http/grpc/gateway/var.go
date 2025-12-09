package gateway

import (
	httpx "github.com/hopeio/gox/net/http"
)

var Marshaler httpx.Marshaler = &JsonPb{}

var InComingHeader = []string{
	httpx.HeaderAccept,
	httpx.HeaderAcceptCharset,
	httpx.HeaderAcceptLanguage,
	httpx.HeaderAcceptRanges,
	httpx.HeaderCacheControl,
	httpx.HeaderContentType,
	httpx.HeaderHost,
	httpx.HeaderVia,
	httpx.HeaderDate,
	httpx.HeaderReferer,
	httpx.HeaderOrigin,
	httpx.HeaderUserAgent,
	httpx.HeaderExpect,
	httpx.HeaderFrom,
	httpx.HeaderPragma,
	httpx.HeaderWarning,
	//"Token",
	//"Cookie",
	"If-Match",
	"If-Modified-Since",
	"If-None-Match",
	"If-Schedule-Key-Match",
	"If-Unmodified-Since",
	"Max-Forwards",
}

var OutgoingHeader = []string{httpx.HeaderSetCookie}
