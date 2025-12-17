package http

import (
	"io"

	jsonx "github.com/hopeio/gox/encoding/json"
	"github.com/hopeio/gox/errors"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Codec interface {
	Marshaler
	Unmarshaler
}

type Unmarshaler interface {
	Unmarshal(data []byte, v any) error
}

// Marshaler defines a conversion between byte sequence and gRPC payloads / fields.
type Marshaler interface {
	// Marshal marshals "v" into byte sequence.
	Marshal(v any) ([]byte, error)
	// ContentType returns the Content-Type which this marshaler is responsible for.
	// The parameter describes the type which is being marshalled, which can sometimes
	// affect the content type returned.
	ContentType(v any) string
}

// Decoder decodes a byte sequence
type Decoder interface {
	Decode(v any) error
}

// Encoder encodes gRPC payloads / fields into byte sequence.
type Encoder interface {
	Encode(v any) error
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

type Json struct {
}

func (*Json) ContentType(_ any) string {
	return ContentTypeJson
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
	case error:
		return jsonx.Marshal(errors.ErrRespFrom(msg))
	}
	return jsonx.Marshal(&RespAnyData{Data: v})
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

// NewDecoder returns a runtime.Decoder which reads JSON stream from "r".
func (j *Json) NewDecoder(r io.Reader) Decoder {
	return jsonx.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *Json) NewEncoder(w io.Writer) Encoder {
	return jsonx.NewEncoder(w)
}
