package http

import jsonx "github.com/hopeio/gox/encoding/json"

// Codec defines a conversion between byte sequence and gRPC payloads / fields.
type Codec interface {
	// Marshal marshals "v" into byte sequence.
	Marshal(v any) ([]byte, error)
	// Unmarshal unmarshals "data" into "v".
	// "v" must be a pointer value.
	Unmarshal(data []byte, v any) error
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
	// same behaviour as for `Codec.ContentType`, but is called, if present,
	// in the case of a streamed response.
	StreamContentType(v any) string
}

type Json struct {
}

func (*Json) ContentType(_ any) string {
	return ContentTypeJson
}

func (j *Json) Marshal(v any) ([]byte, error) {
	return jsonx.Marshal(v)
}

func (j *Json) Unmarshal(data []byte, v any) error {
	return jsonx.Unmarshal(data, v)
}
