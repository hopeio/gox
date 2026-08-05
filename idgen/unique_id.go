package idgen

import (
	crand "crypto/rand"
	"io"
)

// UniqueID returns the result.
func UniqueID() ID {
	id := make(ID, 16)
	io.ReadFull(crand.Reader, id[:])
	return id
}
