package rid

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync"
)

var (
	randBuffer    = [randomLen]byte{}
	randBufferLen = len(randBuffer)
)

// randomMachineId generates a fallback machine ID
func randomMachineId() ([]byte, error) {
	b := make([]byte, 2)
	_, err := rand.Reader.Read(b)
	return b, err
}

// rng represents a random number generator.
type rng struct {
	lastUpdated int64           // when map was last updated, or 0
	exists      map[uint32]bool //
	mu          sync.RWMutex
}

// Next returns a cryptographically secure random uint32 guaranteed to be
// unique for each timestamp (seconds from Unix epoc). This function is
// goroutine safe.
func (r *rng) Next(ts int64) uint32 {
	if r.lastUpdated != ts {
		// reset the mapping each new second
	    r.mu.Lock()
		for k := range r.exists {
			delete(r.exists, k)
		}
		r.lastUpdated = ts
	    r.mu.Unlock()
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for {
		_, err := io.ReadAtLeast(rand.Reader, randBuffer[:], randBufferLen)
		if err != nil {
			panic(fmt.Errorf("rid: cannot generate random number: %v;", err))
		}
		i := uint32(randBuffer[0])<<24 | uint32(randBuffer[1])<<16 | uint32(randBuffer[2])<<8 | uint32(randBuffer[3])

		if !r.exists[i] {
			r.exists[i] = true
			return i
		}
	}
}
