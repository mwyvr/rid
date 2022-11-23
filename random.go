package rid

import (
    "fmt"
    "sync"
    "crypto/rand"
)

// randUint32 generates a cryptographically secure random uint32
func randUint32() uint32 {
	b := make([]byte, 4)
	if _, err := rand.Reader.Read(b); err != nil {
		panic(fmt.Errorf("rid: cannot generate random number: %v;", err))
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

// rng represents a random number generator providing random numbers guaranteed
// to be unique for each second. rid further ensures randomness by
// timestamp+machine ID+process ID.
type rng struct {
    lastUpdated int64 // when map was last updated, or 0
    exists map[uint32]bool // 
    mu sync.Mutex
}

// BySecond returns a random uint32 guaranteed to be unique for each ts
// (timestamp or second from Unix epoc) tick of the clock. Concurrency-safe.
func (r *rng) BySecond (ts int64) uint32 {
    r.mu.Lock()
    defer r.mu.Unlock()
    // reset the mapping each new second
    if r.lastUpdated != ts {
        r.lastUpdated = ts
        r.exists = make(map[uint32]bool)
    }

    for {
        i := randUint32()
        if !r.exists[i] {
            r.exists[i] = true
            return i
        } 
    }
}
