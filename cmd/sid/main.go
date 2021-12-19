// Package main - the `sid` command - provides a very simple facility to generate or inspect sids.
package main

import (
	"flag"
	"fmt"
	"os"

	"oldsid/sid"
)

func main() {

	help := flag.Bool("h", false, "help!")
	flag.Parse()
	args := flag.Args()

	if *help {
		fmt.Println(usage)
		os.Exit(0)
	}

	errors := 0
	for _, arg := range args {
		id, err := sid.FromString(arg)
		if err != nil {
			errors++
			fmt.Printf("[%13s] %s\n", arg, err)
			continue
		}
		fmt.Printf("[%s] ms:%d count:%10d time:%v\n", arg, id.Milliseconds(), id.Count(), id.Time())
	}
	if errors > 0 {
		fmt.Printf("%d error(s)\n", errors)
		os.Exit(1)
	}

	// generate one
	if len(args) == 0 {
		fmt.Println(sid.New())
	}
}

var usage = `Usage:

    sid                     - generates a single short ID
    sid <1 or more SIDs>    - shows details of each SID
    -h                      - this help text

eg:

    sid af3fwszmg7dkp
    [af3fwszmg7dkp] ms:1590635178551 count:18070 time:2020-05-27 20:06:18.551 -0700 PDT
`
