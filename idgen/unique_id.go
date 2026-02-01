package idgen

import (
	crand "crypto/rand"
	"io"
)

func UniqueID() string {
	var id ID
	io.ReadFull(crand.Reader, id[:])
	return id.String()
}
