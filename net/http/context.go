package http

import (
	"time"

	"github.com/hopeio/scaffold/context"
)

type requestCtxKey struct{}

var RequestCtxKey = requestCtxKey{}

type deviceInfoKey struct{}

var DeviceInfoKey = deviceInfoKey{}

type authInfoKey struct{}

var AuthInfoKey = authInfoKey{}

type RequestMetadata struct {
	Device    *context.DeviceInfo
	Auth      *AuthInfo
	RequestAt time.Time
	TraceId   string
}

type requestMetadataKey struct{}

var RequestMetadataKey = requestMetadataKey{}
