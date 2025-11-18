/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package grpc_gateway

import (
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hopeio/gox/encoding/json"
	"github.com/hopeio/gox/net/http/grpc/gateway"
)

var JsonPb = &JSONPb{}

type JSONPb struct {
	gateway.JSONPb
}

// NewDecoder returns a runtime.Decoder which reads JSON stream from "r".
func (j *JSONPb) NewDecoder(r io.Reader) runtime.Decoder {
	return json.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *JSONPb) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}
