package rid

import (
    "fmt"
    "sync"
    "encoding/binary"
    crypto_rand "crypto/rand"
    math_rand "math/rand"
)

func init() {
    var b [8]byte
    _, err := crypto_rand.Read(b[:])
    if err != nil {
        panic("cannot seed math/rand package with cryptographically secure random number generator")
    }
    math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

// randUint generates a random uint32 for use as one component of a unique rid.
func XrandUint() uint32 {
	b := make([]byte, 4)
	if _, err := crypto_rand.Reader.Read(b); err != nil {
		panic(fmt.Errorf("rid: cannot generate random number: %v;", err))
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func randUint32() uint32 {
    return math_rand.Uint32()
}

// rng represents a random number generator providing random numbers guaranteed
// to be unique for each second per machine per process.
type rng struct {
    lastSecond int64
    wasGenerated map[uint32]bool
    mu sync.Mutex
}

// newRng returns an initialized Random number generator
func newRng() *rng {
    return &rng {lastSecond: 0, wasGenerated: make(map[uint32]bool)}
}

// BySecond returns a random uint32 guaranteed to be unique for each second
// tick of the clock. This function is concurrency-safe.
func (r *rng) BySecond (second int64) uint32 {
    r.mu.Lock()
    defer r.mu.Unlock()
    // reset the mapping for every new second
    if r.lastSecond != second {
        r.lastSecond = second
        r.wasGenerated = make(map[uint32]bool)
    }

    for {
        i := randUint32()
        if !r.wasGenerated[i] {
            r.wasGenerated[i] = true
            return i
        } 
    }
}
