[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/rid?tab=doc)[![Test](https://github.com/solutionroute/rid/actions/workflows/test.yaml/badge.svg)](https://github.com/solutionroute/rid/actions/workflows/test.yaml)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/rid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# rid

Package rid provides a [k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence),
zero-configuration, unique ID generator.  Binary IDs Base-32 encode as a
20-character URL-friendly representation like: `ce0e7egs24nkzkn6egfg`.

The 12-byte binary representation of an ID is comprised of a:

    - 4-byte timestamp value representing seconds since the Unix epoch
    - 2-byte machine identifier
    - 2-byte process identifier
    - 4-byte cryptographically secure generated random value

Including the machine and process info into an ID makes `rid` potentially
suitable, without need for configuration or coordination, for distributed
applications. Your use case may vary.

rid implements a number of well-known interfaces to make interacting with json
and databases more convenient.

**Acknowledgement**: This package borrows _heavily_ from the at-scale capable
[rs/xid](https://github.com/rs/xid) package which itself levers ideas from
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/).

Where this package primarily differs is the use of cryptographically secure
random numbers as opposed to xid's use of a monotonic counter for the last 4
bytes of the ID.

## Usage

```go
    id := rid.New()
    fmt.Printf("%s", id) //  ce0e7egs24nkzkn6egfg
```

## Batteries included

`rid.ID` implements a number of common interfaces including package sql's
driver.Valuer, sql.Scanner, TextMarshaler, TextUnmarshaler, json.Marshaler,
json.Unmarshaler, and Stringer.

Package `rid` also provides a command line tool `rid` allowing for id generation
and inspection. To install: `go install github.com/solutionroute/rid/...`

    $ rid
    ce0e7ygs24nw4zebrz10

    # produce 4
    $ rid -c 4
    ce0e8n0s24p7329f3gfg
    ce0e8n0s24p73q9hazp0
    ce0e8n0s24p73qjbffz0
    ce0e8n0s24p72rxdr64g

    # produce one (or more, with -c) and inspect
    $rid `rid -c 4`

    [ce3qht8s24mjwma64r8g] seconds:1669822697 random:1363551825 machine:[0x19,0x11] pid:10542 time:2022-11-30 07:38:17 -0800 PST ID{0x63,0x87,0x78,0xe9,0x19,0x11,0x29,0x2e,0x51,0x46,0x26,0x11}
    [ce3qht8s24mjxybpys70] seconds:1669822697 random:4185323257 machine:[0x19,0x11] pid:10542 time:2022-11-30 07:38:17 -0800 PST ID{0x63,0x87,0x78,0xe9,0x19,0x11,0x29,0x2e,0xf9,0x76,0xf6,0x4e}
    [ce3qht8s24mjw0280sz0] seconds:1669822697 random:   4720128 machine:[0x19,0x11] pid:10542 time:2022-11-30 07:38:17 -0800 PST ID{0x63,0x87,0x78,0xe9,0x19,0x11,0x29,0x2e,0x0,0x48,0x6,0x7e}
    [ce3qht8s24mjwwecpntg] seconds:1669822697 random:1909241201 machine:[0x19,0x11] pid:10542 time:2022-11-30 07:38:17 -0800 PST ID{0x63,0x87,0x78,0xe9,0x19,0x11,0x29,0x2e,0x71,0xcc,0xb5,0x75}

## Benchmark

Using a random number for one segment of the ID is inherently slower than
an incrementing counter such as used in `github.com/rs/xid`. That said, my 4-core laptop
can generate 4 million unique IDs in 1 second using a single process and scales
up from there.

    $ go test -cpu 1,4,8 -benchmem  -run=^$   -bench  ^.*$
    goos: linux goarch: amd64
    cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz
    BenchmarkNew            	 4041667	       295.5 ns/op	       0 B/op	       0 allocs/op
    BenchmarkNew-4          	 9856077	       121.5 ns/op	       0 B/op	       0 allocs/op
    BenchmarkNew-8          	12913791	        92.11 ns/op	       0 B/op	       0 allocs/op

On an 8-core AMD desktop:

    $ go test -cpu 1,4,8,16 -benchmem  -run=^$   -bench  ^.*$
    goos: linux goarch: amd64
    cpu: AMD Ryzen 7 3800X 8-Core Processor
    BenchmarkNew              	 2934926	       348.3 ns/op	       0 B/op	       0 allocs/op
    BenchmarkNew-4            	 6130580	       175.4 ns/op	       0 B/op	       0 allocs/op
    BenchmarkNew-8            	11195751	        93.88 ns/op	       0 B/op	       0 allocs/op
    BenchmarkNew-16           	20034466	        60.26 ns/op	       0 B/op	       0 allocs/op

## See Also

If you don't need the sortable semi-randomness this package provides, consider
the well tested and performant k-sortable `rs/xid` package upon which `rid` is
based. See https://github.com/rs/xid.

For a comparison of various golang unique ID solutions, including `rs/xid`, see:
https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html
