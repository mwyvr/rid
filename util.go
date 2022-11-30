package rid

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	rander = rand.Reader
	rGen   = &randomGenerator{lastTick: 0, isExists: make(map[[4]byte]bool)}
)

// randomGenerator represents a random number generator providing random numbers guaranteed
// to be unique for each second per machine per process.
type randomGenerator struct {
	lastTick uint32
	isExists map[[4]byte]bool
	rbytes   [4]byte // arrays, not slices, allowed as keys to map
	mu       sync.RWMutex
}

// BySecond returns a random uint32 guaranteed to be unique for each second
// tick of the clock. This function is concurrency-safe.
func (r *randomGenerator) Next(second uint32) []byte {
	// reset the mapping for every new second, when called
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.lastTick != second {
		r.lastTick = second
		for k := range r.isExists {
			delete(r.isExists, k)
			// fmt.Println("was dupe", k)
		}
	}

	for {
		randomBytes(r.rbytes[:]) // pass as a slice
		if !r.isExists[r.rbytes] {
			r.isExists[r.rbytes] = true
			return r.rbytes[:]
		}
	}
}

// randomBytes completely fills slice b with random data via crypto/rand
func randomBytes(b []byte) {
	// as this should *never* fail, panic is appropriate
	if _, err := io.ReadFull(rander, b); err != nil {
		panic(fmt.Errorf("rid: cannot generate random number: %v;", err))
	}
}

// randomMachineId generates a fallback machine ID
func randomMachineId() []byte {
	b := make([]byte, 2)
	randomBytes(b)
	return b
}

// readMachineId generates machine id and puts it into the machineId global
// variable. If this function fails to get the hostname, and the fallback
// fails, it will cause a runtime error.
func readMachineID() []byte {
	id := make([]byte, 2)
	hid, err := readPlatformMachineID()
	if err != nil || len(hid) == 0 {
		hid, err = os.Hostname()
	}
	if err == nil && len(hid) != 0 {
		hw := md5.New()
		hw.Write([]byte(hid))
		copy(id, hw.Sum(nil))
	} else {
		// Fallback to rand number if machine id can't be gathered
		id = randomMachineId()
	}
	return id
}
