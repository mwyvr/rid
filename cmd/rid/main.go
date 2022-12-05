// Package main - the `rid` command - generate or inspect rids.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/solutionroute/rid"
)

var count int = 1

func init() {
	flag.IntVar(&count, "c", count, "Generate count number of IDs")
}

func main() {
	flag.Usage = func() {
		pgm := os.Args[0]
		fmt.Fprintf(flag.CommandLine.Output(),
			"usage: %s -c N                 # generate N rids\n", pgm)
		fmt.Fprintf(flag.CommandLine.Output(),
			"       %s cdym59rs24a5g86efepg # decode one or more rid(s)\n", pgm)
	}
	flag.Parse()
	args := flag.Args()

	if count > 1 && len(args) > 0 {
		fmt.Fprintf(flag.CommandLine.Output(),
			"error: -c (output) and args (input) both specified; perform only one at a time.\n")
		flag.Usage()
		os.Exit(1)
	}

	errors := 0
	// If args, attempt to decode as an rid
	for _, arg := range args {
		id, err := rid.FromString(arg)
		if err != nil {
			errors++
			fmt.Printf("[%s] %s\n", arg, err)
			continue
		}
		fmt.Printf("%s seconds:%d rtsig:[%s] random:%15d | time:%v ID{%s}\n", arg,
			id.Seconds(), asHex(id.RuntimeSignature()), id.Random(), id.Time(), asHex(id.Bytes()))
	}
	if errors > 0 {
		fmt.Printf("%d error(s)\n", errors)
		os.Exit(1)
	}

	// if -c N, generate one rid
	if len(args) == 0 {
		for c := 0; c < count; c++ {
			fmt.Fprintf(os.Stdout, "%s\n", rid.New())
		}
	}
}

func asHex(b []byte) string {
	s := []string{}
	for _, v := range b {
		s = append(s, fmt.Sprintf("%#x", v))
	}
	return strings.Join(s, ",")
}
