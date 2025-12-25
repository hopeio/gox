package http

import (
	jsonx "github.com/hopeio/gox/encoding/json"
)

var (
	DefaultMarshal MarshalFunc = func(r any, v any) (data []byte, contentType string) {
		var err error
		switch msg := v.(type) {
		case *CommonAnyResp, *ErrResp:
			data, err = jsonx.Marshal(msg)
		case error:
			data, err = jsonx.Marshal(ErrRespFrom(msg))
		}
		data, err = jsonx.Marshal(&CommonAnyResp{Data: v})
		if err != nil {
			data = []byte(err.Error())
			return data, ContentTypeText
		}
		return data, ContentTypeJson
	}
)

type BindFunc func(r Source, v any) error
type MarshalFunc func(request, v any) (data []byte, contentType string)

type Codec interface {
	Marshaler
	Unmarshaler
}

type Unmarshaler interface {
	Unmarshal(contentType string, data []byte, v any) error
}

// Marshaler defines a conversion between byte sequence and gRPC payloads / fields.
type Marshaler interface {
	// Marshal marshals "v" into byte sequence.
	Marshal(request, v any) (data []byte, contentType string)
}

// Decoder decodes a byte sequence
type Decoder interface {
	Decode(any) error
}

// Encoder encodes gRPC payloads / fields into byte sequence.
type Encoder interface {
	Encode(any) error
}

// DecoderFunc adapts an decoder function into Decoder.
type DecoderFunc func(v any) error

// Decode delegates invocations to the underlying function itself.
func (f DecoderFunc) Decode(v any) error { return f(v) }

// EncoderFunc adapts an encoder function into Encoder
type EncoderFunc func(v any) error

// Encode delegates invocations to the underlying function itself.
func (f EncoderFunc) Encode(v any) error { return f(v) }

// Delimited defines the streaming delimiter.
type Delimited interface {
	// Delimiter returns the record separator for the stream.
	Delimiter() []byte
}

// StreamContentType defines the streaming content type.
type StreamContentType interface {
	// StreamContentType returns the content type for a stream. This shares the
	// same behaviour as for `Marshaler.ContentType`, but is called, if present,
	// in the case of a streamed response.
	StreamContentType(v any) string
}
