package idgen

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"
)

var defaultRandomIDGenerator randomIDGenerator

func init() {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	defaultRandomIDGenerator.randSource = rand.New(rand.NewSource(rngSeed))
}

type randomIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

func NewRandomIDGenerator(randSource *rand.Rand) *randomIDGenerator {
	return &randomIDGenerator{randSource: randSource}
}

func NewRandomID() ID {
	defaultRandomIDGenerator.Lock()
	defer defaultRandomIDGenerator.Unlock()
	sid := make(ID, 16)
	for {
		_, _ = defaultRandomIDGenerator.randSource.Read(sid[:])
		if sid.IsValid() {
			break
		}
	}
	return sid
}
