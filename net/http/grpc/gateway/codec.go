package gateway

import (
	"context"
	"strings"

	jsonx "github.com/hopeio/gox/encoding/json"
	httpx "github.com/hopeio/gox/net/http"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func init() {
	httpx.DefaultMarshal = DefaultMarshal
	httpx.DefaultUnmarshal = Unmarshaller
}

var DefaultMarshal httpx.MarshalFunc = func(ctx context.Context, v any) (data []byte, contentType string) {
	switch msg := v.(type) {
	case *wrapperspb.StringValue:
		v = msg.Value
	case *wrapperspb.BoolValue:
		v = msg.Value
	case *wrapperspb.Int32Value:
		v = msg.Value
	case *wrapperspb.Int64Value:
		v = msg.Value
	case *wrapperspb.UInt32Value:
		v = msg.Value
	case *wrapperspb.UInt64Value:
		v = msg.Value
	case *wrapperspb.FloatValue:
		v = msg.Value
	case *wrapperspb.DoubleValue:
		v = msg.Value
	case *wrapperspb.BytesValue:
		v = msg.Value
	case *httpx.CommonAnyResp, *httpx.ErrResp:
		data, err := jsonx.Marshal(msg)
		if err != nil {
			data = []byte(err.Error())
			return data, httpx.ContentTypeText
		}
		return data, httpx.ContentTypeJson
	case error:
		data, err := jsonx.Marshal(httpx.ErrRespFrom(msg))
		if err != nil {
			data = []byte(err.Error())
			return data, httpx.ContentTypeText
		}
		return data, httpx.ContentTypeJson
	}
	data, err := jsonx.Marshal(&httpx.CommonAnyResp{Data: v})
	if err != nil {
		data = []byte(err.Error())
		return data, httpx.ContentTypeText
	}
	return data, httpx.ContentTypeJson
}

var Unmarshaller = func(ctx context.Context, contentType string, data []byte, v any) error {
	if strings.HasPrefix(contentType, httpx.ContentTypeProtobuf) {
		return proto.Unmarshal(data, v.(proto.Message))
	}
	var wrapped httpx.CommonAnyResp
	if err := jsonx.Unmarshal(data, &wrapped); err == nil && wrapped.Data != nil {
		inner, err := jsonx.Marshal(wrapped.Data)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "marshal frame data: %v", err)
		}
		if err := unmarshalInner(inner, v); err != nil {
			return status.Errorf(codes.InvalidArgument, "unmarshal frame: %v", err)
		}
		return nil
	}
	if err := jsonx.Unmarshal(data, v); err != nil {
		return status.Errorf(codes.InvalidArgument, "unmarshal frame: %v", err)
	}
	return nil
}

func unmarshalInner(inner []byte, v any) error {
	if err := jsonx.Unmarshal(inner, v); err == nil {
		return nil
	}
	switch msg := v.(type) {
	case *wrapperspb.StringValue:
		var s string
		if err := jsonx.Unmarshal(inner, &s); err == nil {
			msg.Value = s
			return nil
		}
	case *wrapperspb.BoolValue:
		var b bool
		if err := jsonx.Unmarshal(inner, &b); err == nil {
			msg.Value = b
			return nil
		}
	case *wrapperspb.Int32Value:
		var n int32
		if err := jsonx.Unmarshal(inner, &n); err == nil {
			msg.Value = n
			return nil
		}
	case *wrapperspb.Int64Value:
		var n int64
		if err := jsonx.Unmarshal(inner, &n); err == nil {
			msg.Value = n
			return nil
		}
	case *wrapperspb.UInt32Value:
		var n uint32
		if err := jsonx.Unmarshal(inner, &n); err == nil {
			msg.Value = n
			return nil
		}
	case *wrapperspb.UInt64Value:
		var n uint64
		if err := jsonx.Unmarshal(inner, &n); err == nil {
			msg.Value = n
			return nil
		}
	case *wrapperspb.FloatValue:
		var f float32
		if err := jsonx.Unmarshal(inner, &f); err == nil {
			msg.Value = f
			return nil
		}
	case *wrapperspb.DoubleValue:
		var f float64
		if err := jsonx.Unmarshal(inner, &f); err == nil {
			msg.Value = f
			return nil
		}
	case *wrapperspb.BytesValue:
		var b []byte
		if err := jsonx.Unmarshal(inner, &b); err == nil {
			msg.Value = b
			return nil
		}
	}
	return jsonx.Unmarshal(inner, v)
}