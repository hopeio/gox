package idgen

import (
	crand "crypto/rand"
	"math/rand"
	"sync"
)

type randomIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

// NewRandomIDGenerator creates a generator backed by the given math/rand source.
// Prefer NewRandomID for security-sensitive identifiers: math/rand is
// predictable and must not be used where uniqueness/unguessability matters.
func NewRandomIDGenerator(randSource *rand.Rand) *randomIDGenerator {
	return &randomIDGenerator{randSource: randSource}
}

// NewRandomID creates a cryptographically secure random ID using crypto/rand.
func NewRandomID() ID {
	sid := make(ID, 16)
	for {
		_, _ = crand.Read(sid[:])
		if sid.IsValid() {
			break
		}
	}
	return sid
}
