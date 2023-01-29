// A utility to generate or inspect IDs.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/solutionroute/rid"
)

func main() {

	count := flag.Int("c", 1, "Generate n IDs")
	flag.Parse()
	args := flag.Args()

	if *count > 1 && len(args) > 0 {
		fmt.Fprintf(flag.CommandLine.Output(),
			"error: -c (generate N outputs) and args (inspect inputs) both specified; perform only one at a time.\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(args) > 0 {
		// attempt to decode each as an rid
		for _, arg := range args {
			id, err := rid.FromString(arg)
			if err != nil {
				fmt.Printf("[%s] %s\n", arg, err)
				continue
			}
			fmt.Printf("%s ts:%d rnd:%15d %s ID{%s }\n", arg,
				id.Timestamp(), id.Random(), id.Time(), asHex(id.Bytes()))
		}
	} else {
		// generate one or -c N ids
		for c := 1; c <= *count; c++ {
			fmt.Fprintf(os.Stdout, "%s\n", rid.New())
		}
	}
}

func asHex(b []byte) string {
	s := []string{}
	for _, v := range b {
		s = append(s, fmt.Sprintf(" %#4x", v))
	}
	return strings.Join(s, ",")
}
