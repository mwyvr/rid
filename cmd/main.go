package main

import (
	"fmt"
	"time"

	"github.com/solutionroute/sid"
)

var now = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

func main() {
	t := sid.New()
	show(t)
	b := sid.ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	show(b)
	b = sid.ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	show(b)
	a := sid.NewWithTime(now)
	show(a)

}

func show(id sid.ID) {
	// fmt.Printf("%4v %s %-16d %5d %8v\n", id[:], id.String(), id.Milliseconds(), id.Count(), id.Time().UTC().Format(time.RFC822))
	fmt.Printf("%3v %s %-16d %5d %8v\n", id[:], id.String(), id.Milliseconds(), id.Count(), id.Time().UTC())
}
