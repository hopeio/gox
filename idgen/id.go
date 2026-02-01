package idgen

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/hopeio/gox/strings"
)

type ID [16]byte

var (
	nilTraceID ID
	_          json.Marshaler = nilTraceID
)

// IsValid checks whether the trace TraceID is valid. A valid trace ID does
// not consist of zeros only.
func (t ID) IsValid() bool {
	return !bytes.Equal(t[:], nilTraceID[:])
}

// MarshalJSON implements a custom marshal function to encode TraceID
// as a hex string.
func (t ID) MarshalJSON() ([]byte, error) {
	return strings.ToBytes(`"` + t.String() + `"`), nil
}

func (t ID) UnmarshalJSON(data []byte) error {
	if len(data) != 18 {
		return errors.New("invalid ID")
	}
	_, err := hex.Decode(t[:], data[1:17])
	return err
}

// String returns the hex string representation form of a TraceID.
func (t ID) String() string {
	return hex.EncodeToString(t[:])
}
