/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gateway

import (
	"io"

	"github.com/hopeio/gox/encoding/json"
	httpx "github.com/hopeio/gox/net/http"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var Marshaler httpx.Marshaler = &JsonPb{}

type JsonPb struct {
}

func (*JsonPb) ContentType(_ interface{}) string {
	return httpx.ContentTypeJson
}

func (j *JsonPb) Marshal(v any) ([]byte, error) {
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
	}
	return json.Marshal(&httpx.RespAnyData{
		Data: v,
	})
}

func (j *JsonPb) Name() string {
	return "jsonpb"
}

func (j *JsonPb) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (j *JsonPb) Delimiter() []byte {
	return []byte("\n")
}

// NewDecoder returns a runtime.Decoder which reads JSON stream from "r".
func (j *JsonPb) NewDecoder(r io.Reader) httpx.Decoder {
	return json.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *JsonPb) NewEncoder(w io.Writer) httpx.Encoder {
	return json.NewEncoder(w)
}
