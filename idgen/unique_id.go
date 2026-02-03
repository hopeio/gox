package idgen

import (
	crand "crypto/rand"
	"io"
)

func UniqueID() ID {
	id := make(ID, 16)
	io.ReadFull(crand.Reader, id[:])
	return id
}
