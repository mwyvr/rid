package main

// uniqcheck:
// determine if the approach used delivers unique IDs in go applications
// running in more than one go routine.
//
// In addition to this test, using stdout / sort / uniq, a run of 10 million:
//
//  rid -c 10000000 | sort | uniq -d
//  (no output, meaning no duplicates)

import (
	"sync"
	"time"

	"github.com/solutionroute/rid"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const rawLen = 10

var (
	genPerRoutine = int(1 * 10e5)
	numRoutines   = 16
	dupes         = 0
	exists        = check{lastTick: 0, keys: make(map[[rawLen]byte]bool)} // keys can be arrays, not slices
	fmt           = message.NewPrinter(language.English)
)

type check struct {
	keys      map[[rawLen]byte]bool
	lastTick  int64
	totalKeys int
	mu        sync.RWMutex
}

func main() {
	var wg sync.WaitGroup

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			generate()
		}(i)
	}
	wg.Wait()
	fmt.Printf("Total keys: %d. Keys in last time tick: %d. Number of dupes: %d\n", exists.totalKeys, len(exists.keys), dupes)
}

func generate() {
	var id [rawLen]byte
	for i := 0; i < genPerRoutine; i++ {
		tmp := rid.New()
		copy(id[:], tmp[:])
		tmpTimestamp := time.Now().Unix()
		exists.mu.Lock()
		// clear per each new second
		if exists.lastTick != tmpTimestamp {
			exists.lastTick = tmpTimestamp
			exists.keys = make(map[[rawLen]byte]bool)
		}
		if !exists.keys[id] {
			exists.keys[id] = true
			exists.totalKeys += 1
		} else {
			dupes += 1
			exists.totalKeys += 1
			fmt.Printf("Generated: %d, found duplicate: %v\n", exists.totalKeys, id)
		}
		exists.mu.Unlock()
	}
}
