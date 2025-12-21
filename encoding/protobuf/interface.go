package protobuf

type Unmarshaler interface {
	UnmarshalProto([]byte) error
}

type Marshaler interface {
	MarshalProto() ([]byte, error)
}
