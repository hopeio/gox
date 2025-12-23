package gateway

import (
	jsonx "github.com/hopeio/gox/encoding/json"
	httpx "github.com/hopeio/gox/net/http"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Json struct {
}

func (*Json) ContentType(_ any) string {
	return httpx.ContentTypeJson
}

func (j *Json) Marshal(v any) ([]byte, error) {
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
		return jsonx.Marshal(msg)
	case error:
		return jsonx.Marshal(httpx.ErrRespFrom(msg))
	}
	return jsonx.Marshal(&httpx.CommonAnyResp{Data: v})
}

func (j *Json) Name() string {
	return "json"
}

func (j *Json) Unmarshal(data []byte, v interface{}) error {
	return jsonx.Unmarshal(data, v)
}

func (j *Json) Delimiter() []byte {
	return []byte("\n")
}
