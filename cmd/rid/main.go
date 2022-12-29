// A utility to generate or inspect rids.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/solutionroute/rid"
)

var (
	count  int  = 1
	altEnc bool = false
)

func init() {
	flag.IntVar(&count, "c", count, "Generate count number of IDs")
}

func main() {
	var (
		id  rid.ID
		err error
	)

	flag.Parse()
	args := flag.Args()

	if count > 1 && len(args) > 0 {
		fmt.Fprintf(flag.CommandLine.Output(),
			"error: -c (output) and args (input) both specified; perform only one at a time.\n")
		flag.Usage()
		os.Exit(1)
	}

	errors := 0
	// If args, attempt to decode as an rid; can't mix Base32 and alt Base64
	for _, arg := range args {
		err = nil
		id, err = rid.FromString(arg)
		if err != nil {
			errors++
			fmt.Printf("[%s] %s\n", arg, err)
			continue
		}
		fmt.Printf("%s ts:%d rnd:%15d %s ID{%s}\n", arg,
			id.Timestamp(), id.Random(), id.Time(), asHex(id.Bytes()))
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
