package main

// package to determine if the approach is unique.
// todo - run other compared packages through the gauntlet.

import (
	"sync"

	"github.com/segmentio/ksuid"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type check struct {
	keys map[[15]byte]bool
	mu   sync.RWMutex
}

var (
	genPerThread = 20000000
	dupes        = 0
	exists       = check{keys: make(map[[15]byte]bool)}
	fmt          = message.NewPrinter(language.English)
)

func main() {
	var wg sync.WaitGroup
	numThreads := 8

	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func(i int) {
			fmt.Println("go routine", i)
			defer wg.Done()
			generate()
		}(i)

	}
	wg.Wait()
	fmt.Printf("Total keys: %d. Number of dupes: %d\n", len(exists.keys), dupes)
}

func generate() {
	var id [15]byte
	for i := 0; i < genPerThread; i++ {
		// tmp := rid.New()
		// todo - run other compared packages through the gauntlet.
		tmp := ksuid.New()
		copy(id[:], tmp[:])
		exists.mu.Lock()
		if !exists.keys[id] {
			exists.keys[id] = true
		} else {
			dupes += 1
			fmt.Printf("duplicate: %v\n", id)
		}
		exists.mu.Unlock()
	}
}
