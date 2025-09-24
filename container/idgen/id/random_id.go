package id

import (
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"github.com/hopeio/gox/strings"
	"io"
	"math/rand"
	"sync"
)

var defaultIDGenerator randomIDGenerator

func init() {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	defaultIDGenerator.randSource = rand.New(rand.NewSource(rngSeed))
}

type randomIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

func NewRandomID() RandomID {
	defaultIDGenerator.Lock()
	defer defaultIDGenerator.Unlock()
	sid := RandomID{}
	for {
		_, _ = defaultIDGenerator.randSource.Read(sid[:])
		if sid.IsValid() {
			break
		}
	}
	return sid
}

type RandomID [16]byte

var (
	nilTraceID RandomID
	_          json.Marshaler = nilTraceID
)

// IsValid checks whether the trace TraceID is valid. A valid trace ID does
// not consist of zeros only.
func (t RandomID) IsValid() bool {
	return !bytes.Equal(t[:], nilTraceID[:])
}

// MarshalJSON implements a custom marshal function to encode TraceID
// as a hex string.
func (t RandomID) MarshalJSON() ([]byte, error) {
	return strings.ToBytes(`"` + t.String() + `"`), nil
}

// String returns the hex string representation form of a TraceID.
func (t RandomID) String() string {
	return hex.EncodeToString(t[:])
}

func UniqueID() string {
	id := make([]byte, 16)
	io.ReadFull(crand.Reader, id)
	return hex.EncodeToString(id)
}
