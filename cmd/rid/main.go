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
	count := 1
	flag.IntVar(&count, "c", count, "Generate N-count IDs")
	flag.Usage = func() {
		fs := flag.CommandLine
		fcount := fs.Lookup("c")
		fmt.Printf("Usage: rid\n\n")
		fmt.Printf("Options:\n")
		fmt.Printf("  rid dgm3w9sh9f5flv5s\t\tDecode the supplied Base32 ID\n")
		fmt.Printf("  rid -%s N\t\t\t%s default: %s\n\n", fcount.Name, fcount.Usage, fcount.DefValue)
		fmt.Printf("With no parameters, rid generates %s random ID encoded as Base32.\n", fcount.DefValue)
		fmt.Printf("Generate and inspect 4 random IDs using Linux/Unix command substituion:\n")
		fmt.Printf("  rid `rid -c 4`\n")
	}
	flag.Parse()
	args := flag.Args()

	if count > 1 && len(args) > 0 {
		fmt.Fprintf(flag.CommandLine.Output(),
			"rid: Error, cannot generate ID(s) and inspect at the same time. Use command substituion. \n")
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
		for c := 1; c <= count; c++ {
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
