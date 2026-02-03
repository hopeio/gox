package idgen

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"github.com/hopeio/gox/encoding/multibase"
	"github.com/hopeio/gox/strings"
)

type ID []byte

// IsValid checks whether the trace TraceID is valid. A valid trace ID does
// not consist of zeros only.
func (t ID) IsValid() bool {
	for _, b := range t {
		if b != 0 {
			return true
		}
	}
	return false
}

// MarshalJSON implements a custom marshal function to encode ID
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

func (t ID) Hex() string {
	return hex.EncodeToString(t[:])
}

func (t ID) Bytes() []byte {
	return t[:]
}

func (t ID) Base32() string {
	return base32.StdEncoding.EncodeToString(t)
}

func (t ID) Base58() string {
	return multibase.EncodeBase58(t)
}

func (t ID) Base62() string {
	return multibase.EncodeBase62(t)
}

func (t ID) Base64() string {
	return base64.StdEncoding.EncodeToString(t)
}
