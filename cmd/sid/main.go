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
)

func init() {
	flag.IntVar(&count, "c", count, "Generate n * IDs")
}

func main() {
    flag.Usage = func() {
        pgm := os.Args[0]
        fmt.Fprintf(flag.CommandLine.Output(), "usage: %s -c N          # print N sid(s)\n", pgm)
        fmt.Fprintf(flag.CommandLine.Output(), "       %s 0629p0rqdrw8p # decode one or more sid(s)\n", pgm)
        // flag.PrintDefaults()
    }
	flag.Parse()
	args := flag.Args()
    
    if count > 1 && len(args) > 0 {
        fmt.Fprintf(flag.CommandLine.Output(), "error: -c (output) and args (input) both specified; perform only one at a time.\n")
        flag.Usage()
        os.Exit(1)
    }

	errors := 0
    fmt.Fprintf(os.Stdout, "args: %s %d", args, len(args))

    // If no valid flag, Either attempt to decode string as a sid
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
        fmt.Printf("[%s] seconds:%d entropy:%d machine:%v pid:%v time:%v ID%s\n", arg, id.Seconds(), id.Entropy(), id.Machine(), id.Pid(), id.Time(), s)
	}
	if errors > 0 {
		fmt.Printf("%d error(s)\n", errors)
		os.Exit(1)
	}

	// OR... generate one (or -c value) sid
    if len(args) == 0 {
        for c := 0; c < count; c++ {
            fmt.Fprintf(os.Stdout, "%s", sid.New())
            if count > 1 {
                fmt.Println()
            }
        }

    }
}
