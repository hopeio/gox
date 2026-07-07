package http

import (
	"context"

	jsonx "github.com/hopeio/gox/encoding/json"

)

var (
	DefaultUnmarshal UnmarshalFunc = func(ctx context.Context, contentType string, data []byte, v any) error {
		return jsonx.Unmarshal(data, v)
	}

	DefaultMarshal MarshalFunc = func(ctx context.Context, v any) (data []byte, contentType string, err error) {
		switch msg := v.(type) {
		case *CommonAnyResp, *ErrResp:
			data, err = jsonx.Marshal(msg)
		case error:
			data, err = jsonx.Marshal(ErrRespFrom(msg))
		}
		data, err = jsonx.Marshal(&CommonAnyResp{Data: v})
		if err != nil {
			return data, ContentTypeText, err
		}
		return data, ContentTypeJson, nil
	}
)


func JsonMarshal(ctx context.Context, v any) ([]byte, string, error) {
	data, err := jsonx.Marshal(v)
	if err != nil {
		return data, ContentTypeText, err
	}
	return data, ContentTypeJson, nil
}


type BindFunc func(r Source, v any) error
type MarshalFunc func(ctx context.Context, v any) (data []byte, contentType string, err error)
type UnmarshalFunc func(ctx context.Context, contentType string, data []byte, v any) error

type Codec interface {
	Marshaler
	Unmarshaler
}

type Unmarshaler interface {
	Unmarshal(ctx context.Context, contentType string, data []byte, v any) error
}

// Marshaler defines a conversion between byte sequence and gRPC payloads / fields.
type Marshaler interface {
	// Marshal marshals "v" into byte sequence.
	Marshal(ctx context.Context, v any) (data []byte, contentType string)
}

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
