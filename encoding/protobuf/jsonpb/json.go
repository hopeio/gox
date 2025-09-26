/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package jsonpb

import (
	"github.com/hopeio/gox/encoding/json"
	responsei "github.com/hopeio/gox/net/http"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var JsonPb = &JSONPb{}

type JSONPb struct {
}

func (*JSONPb) ContentType(_ interface{}) string {
	return "application/json"
}

func (j *JSONPb) Marshal(v any) ([]byte, error) {
	if msg, ok := v.(*wrapperspb.StringValue); ok {
		v = msg.Value
	}
	return json.Marshal(&responsei.RespAnyData{
		Data: v,
	})
}

func (j *JSONPb) Name() string {
	return "jsonpb"
}

func (j *JSONPb) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (j *JSONPb) Delimiter() []byte {
	return []byte("\n")
}

func (j *JSONPb) ContentTypeFromMessage(v interface{}) string {
	return j.ContentType(v)
}
