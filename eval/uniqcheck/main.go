package main

// package to determine if the approach used delivers unique IDs in single and
// multi-process applications on a single machine. Runs ranging to 100 million
// generations have been done on a 4-core (8cpu) Intel and 8-core (16cpu) machine,
// with zero duplicates detected.
//
// rid millisecond timestamp reduces collision chance eg assuming 10 ns ID generation:
//
// 		1 millisecond / 10 nanosecond per ID = 100,000 possible IDs per time tick
// 		6 bytes of randomness = 281,474,976,710,656 permutations, per time tick
//
// Multi-machine/site apps:
// 		rid runtime signature byte may introduce enough variability, ymmv
//
// Conclusion:
//
// Almost all applications will be spending much more time in business and storage layer
// than the 4 - 50 nanoseconds it takes to produce an ID and thus the absolute chance
// of a collision is zero.
//
// In addition to this test, using stdout / sort / uniq:
//
// 		rid -c 1000000000 > foo  			 # 1,000,000,000 ids, ctrl-c aborted at 227,322,217
// 		sort foo > foosort 		  			 # takes some time on my laptop ;-)
// 		uniq foosort -d > foonotunique # should be a zero byte file
//
// The above produces:
// 		$ ls -ltr foo*
// 		~ 5683055425 Dec  7 08:05 foo
// 		~ 5683055425 Dec  7 08:12 foosort
// 		~ 		     0 Dec  7 08:15 foonotunique

// TODO - run other compared packages through the gauntlet.

import (
	"sync"

	"github.com/solutionroute/rid"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type check struct {
	keys      map[[15]byte]bool
	lastTick  int64 // millisecond for rid, ksuid, second for xid
	totalKeys int
	mu        sync.RWMutex
}

var (
	genPerRoutine = 1000000 // 64 million (8m * 8) or (1m * 16) takes ~25 seconds
	numRoutines   = 64
	dupes         = 0
	exists        = check{lastTick: 0, keys: make(map[[15]byte]bool)} // arrays can be map keys, not slices
	fmt           = message.NewPrinter(language.English)
)

func main() {
	var wg sync.WaitGroup

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			// fmt.Println("go routine", i)
			defer wg.Done()
			generate()
		}(i)

	}
	wg.Wait()
	fmt.Printf("Total keys: %d. Keys in last time tick: %d. Number of dupes: %d\n", exists.totalKeys, len(exists.keys), dupes)
}

func generate() {
	var id [15]byte
	for i := 0; i < genPerRoutine; i++ {
		tmp := rid.New()
		// TODO - run other compared packages through the gauntlet.
		// tmp := xid.New()
		// tmp := ksuid.New()
		copy(id[:], tmp[:])
		tmpTimestamp := tmp.Time().UnixMilli() // milliseconds will be 000 for pkgs using second resolution, works for this
		exists.mu.Lock()
		// crude - we clear per new millisecond (or per second for pkg like xid)
		if exists.lastTick != tmpTimestamp {
			exists.lastTick = tmpTimestamp
			exists.keys = make(map[[15]byte]bool)
		}
		if !exists.keys[id] {
			exists.keys[id] = true
			exists.totalKeys += 1
		} else {
			dupes += 1
			fmt.Printf("duplicate: %v\n", id)
		}
		exists.mu.Unlock()
	}
}

// Runs

// rid (time resolution is actually milliseconds):
// $ time go run eval/uniqcheck/main.go
// Total keys: 64,000,000. Keys in last time tick: 2,280. Number of dupes: 0
//
// real	0m13.068s
// user	0m20.299s
// sys	0m0.986s

// xid (note time resolution is actually seconds thus more keys in last tick):
// $ time go run eval/uniqcheck/main.go
// Total keys: 64,000,000. Keys in last time tick: 829,009. Number of dupes: 0
//
// real	0m18.838s
// user	0m26.756s
// sys	0m2.029s

// ksuid (note time resolution is actually seconds thus more keys in last tick):
// $ time go run eval/uniqcheck/main.go
// Total keys: 64,000,000. Keys in last time tick: 478,533. Number of dupes: 0
//
// real	0m50.437s
// user	1m9.540s
// sys	0m19.915s
