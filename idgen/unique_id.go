package idgen

import (
	crand "crypto/rand"
	"io"
)

func UniqueID() ID {
	var id ID
	io.ReadFull(crand.Reader, id[:])
	return id
}
