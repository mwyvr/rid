// Package main - the `sid` command - generate or inspect sids.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/solutionroute/sid"
)

var (
	count   int  = 1
	newline bool = false
)

func init() {
	flag.IntVar(&count, "c", count, "Generate n * IDs")
	flag.BoolVar(&newline, "n", newline, "Print count > 1 IDs each on a new line (default false)")
}

func main() {
	flag.Parse()
	args := flag.Args()

	errors := 0
	for _, arg := range args {
		id, err := sid.FromString(arg)
		if err != nil {
			errors++
			fmt.Printf("[%s] %s\n", arg, err)
			continue
		}
		// pretty print the bytes
		s := "{"
		for _, b := range id.Bytes() {
			if s != "{" {
				s = s + ", "
			}
			s = s + strconv.Itoa(int(b))
		}
		s = s + "}"
		fmt.Printf("[%s] ms:%d count:%5d time:%v ID%s\n", arg, id.Milliseconds(), id.Count(), id.Time(), s)
	}
	if errors > 0 {
		fmt.Printf("%d error(s)\n", errors)
		os.Exit(1)
	}
	// generate one (or -c value) sid
	if len(args) == 0 {
		for c := 0; c < count; c++ {
			fmt.Printf("%s ", sid.New())
			if newline {
				fmt.Println()
			}
		}
	}
	fmt.Println()
}
