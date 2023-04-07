package trace

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"sync"
)

type randomIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

func newGenerator() *randomIDGenerator {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	return &randomIDGenerator{randSource: rand.New(rand.NewSource(rngSeed))}
}

// NewSpanID returns a non-zero span ID from a randomly-chosen sequence.
func (gen *randomIDGenerator) NewSpanID() string {
	gen.Lock()
	defer gen.Unlock()
	sid := [8]byte{}
	_, _ = gen.randSource.Read(sid[:])
	return hex.EncodeToString(sid[:])
}

// NewTraceID returns a non-zero trace ID from a randomly-chosen sequence.
func (gen *randomIDGenerator) NewTraceID() string {
	gen.Lock()
	defer gen.Unlock()
	tid := [16]byte{}
	_, _ = gen.randSource.Read(tid[:])
	return hex.EncodeToString(tid[:])
}

// NewIDs returns a non-zero trace ID and a non-zero span ID from a
// randomly-chosen sequence.
func (gen *randomIDGenerator) NewIDs() (string, string) {
	gen.Lock()
	defer gen.Unlock()
	tid := [16]byte{}
	_, _ = gen.randSource.Read(tid[:])
	sid := [8]byte{}
	_, _ = gen.randSource.Read(sid[:])
	return hex.EncodeToString(tid[:]), hex.EncodeToString(sid[:])
}
