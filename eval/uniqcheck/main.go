// Package main seeks to determine if the approach used delivers sufficiently
// unique IDs in go applications potentially running multiple goroutines.
//
// Considerations:
//
// - objective: keep IDs and their encoded representation short
// - you can generate a lot of random numbers in 1 second
// - is 48 bits of randomness per second enough
// - using a faster, scalable, random generator raises the bar
//
// In addition to this test, a single-threaded test using stdout / sort / uniq,
// a run of 10 million or more results in no duplicates on various test machines:
//
//	rid -c 10000000 | sort | uniq -d
//	(no output, meaning no duplicates)
//
// Running this test results in output like:
// Total keys: 40,000,000. Keys in last time tick: 1,825,240. Number of dupes: 0
package main

import (
	"sync"
	"time"

	"github.com/solutionroute/rid"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const rawLen = 10

var (
	genPerRoutine = int(5 * 10e5)
	numRoutines   = 8
	dupes         = 0
	// since the underlying structure of ID is an array, not a slice, rid.ID can be a key
	exists = check{lastTick: 0, keys: make(map[rid.ID]bool)}
	fmt    = message.NewPrinter(language.English)
)

type check struct {
	keys      map[rid.ID]bool
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
	var id rid.ID
	for i := 0; i < genPerRoutine; i++ {
		id = rid.New()
		tmpTimestamp := time.Now().Unix()
		exists.mu.Lock()
		if exists.lastTick != tmpTimestamp {
			exists.lastTick = tmpTimestamp
			// reset each new second
			exists.keys = make(map[rid.ID]bool)
		}
		if !exists.keys[id] {
			exists.keys[id] = true
			exists.totalKeys++
		} else {
			dupes++
			exists.totalKeys++
			fmt.Printf("Generated: %d, found duplicate: %v\n", exists.totalKeys, id)
		}
		exists.mu.Unlock()
	}
}
