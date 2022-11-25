package rid

import (
    "fmt"
    "io"
    "sync"
    "crypto/rand"
)

var (
	// BySecond method generates unique random numbers per second clock tick
	rgenerator = &rng{lastUpdated: 0, exists: make(map[uint32]bool)}
    randMutex = sync.Mutex{}
    randBuffer = [randomLen]byte{}
    randBufferLen = len(randBuffer)
)

// randomMachineId generates a fallback machine ID
func randomMachineId() ([]byte, error) {
	b := make([]byte, 2)
	_, err := rand.Reader.Read(b)
    return b, err
}

// randUint32 generates a cryptographically secure random uint32. This function
// is goroutine safe.
func randUint32() uint32 {
    randMutex.Lock()
    _, err := io.ReadAtLeast(rand.Reader, randBuffer[:], randBufferLen)
    if err != nil {
		panic(fmt.Errorf("rid: cannot generate random number: %v;", err))
    }
    randMutex.Unlock()
	return uint32(randBuffer[0])<<24 | uint32(randBuffer[1])<<16 | uint32(randBuffer[2])<<8 | uint32(randBuffer[3])
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
