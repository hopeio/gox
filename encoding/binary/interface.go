package binary

type Unmarshaler interface {
	UnmarshalBinary([]byte) error
}

type Marshaler interface {
	MarshalBinary() ([]byte, error)
}
