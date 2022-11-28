package rid

import (
	"crypto/rand"
	"hash/maphash"
	"sync"
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

// Next returns a psuedo random uint32 guaranteed to be unique for each
// timestamp (seconds from Unix epoch) | machineID | pid. This implementation
// uses hash/maphash to access a fast runtime generated seed as the random
// number.  Why not math/rand or crypto/rand? This approach levers a
// random-enough fast runtime generator providing a 2 - 5 times performance
// increase; even more importantly, it scales better as cores increase.
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

		// Sum64 initializes Seed{}; since there's no bytes in the buffer to hash,
		// what is returned is the Seed itself, i.e.
		// seed {17011520470102362949} -> Sum64: 17011520470102362949
    // from maphash/hash.go:
		// "A Hash is not safe for concurrent use by multiple goroutines, but a Seed is."
		i := uint32(new(maphash.Hash).Sum64() >> 32)

    // but map access requires the lock
		if !r.exists[i] {
			r.exists[i] = true
			return i
		}
	}
}
