package main

// determine if the approach used delivers unique IDs in go applications
// running a single go routine or utilizing multiple cores
//
// In addition to this test, using stdout / sort / uniq, a run of 10 million:
//
//  rid -c 10000000 | sort | uniq -d
//  (no output)

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

// Runs
// 16 go routines * 10e4
// $ time go run main.go
// Total keys: 1,600,000. Keys in last time tick: 702,456. Number of dupes: 0
//
// real	0m1.202s
// user	0m2.643s
// sys	0m0.245s

// 16 * 10e5
// $ time go run main.go
// Total keys: 16,000,000. Keys in last time tick: 247,040. Number of dupes: 0
//
// real	0m11.541s
// user	0m29.377s
// sys	0m1.601s

// 1 routine x 10e7:
// $ time go run main.go
// Total keys: 100,000,000. Keys in last time tick: 776,864. Number of dupes: 0
//
// real	0m29.339s
// user	0m28.195s
// sys	10m1.673s

// 4 routines * 10e7:
// $ time go run main.go
// Generated: 149,958,508, found duplicate: [99 215 1 106 50 126 157 18 97 141]
// Generated: 263,367,074, found duplicate: [99 215 1 175 97 18 247 211 221 213]
// Total keys: 400,000,000. Keys in last time tick: 260,118. Number of dupes: 2
//
// real	4m3.685s
// user	6m34.780s
// sys	0m16.497s
