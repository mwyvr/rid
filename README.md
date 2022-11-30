[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/rid?tab=doc)[![Test](https://github.com/solutionroute/rid/actions/workflows/test.yaml/badge.svg)](https://github.com/solutionroute/rid/actions/workflows/test.yaml)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/rid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# rid

Package rid provides a [k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence),
configuration-free, unique ID generator.  Binary IDs Base-32 encode as a
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

rid.ID implements a number of common interfaces including package sql's
driver.Valuer, sql.Scanner, TextMarshaler, TextUnmarshaler, json.Marshaler,
json.Unmarshaler, and Stringer.

Package rid also provides a command line tool `rid` allowing for id generation
or inspection:

    $ rid
    ce0e7ygs24nw4zebrz10

    # produce 4
    $ rid -c 4
    ce0e8n0s24p7329f3gfg
    ce0e8n0s24p73q9hazp0
    ce0e8n0s24p73qjbffz0
    ce0e8n0s24p72rxdr64g

    # produce one (or more, with -c) and inspect
    $rid `rid -c 2`
    [ce0e960s24phh0qrnz7g] seconds:1669391512 random:2197335938 machine:[0x19, 0x11] pid:11544 time:2022-11-25 07:51:52 -0800 PST ID{0x63, 0x80, 0xe4, 0x98, 0x19, 0x11, 0x2d, 0x18, 0x82, 0xf8, 0xaf, 0xcf}
    [ce0e960s24phh39seya0] seconds:1669391512 random:2369353613 machine:[0x19, 0x11] pid:11544 time:2022-11-25 07:51:52 -0800 PST ID{0x63, 0x80, 0xe4, 0x98, 0x19, 0x11, 0x2d, 0x18, 0x8d, 0x39, 0x77, 0x94}

## Benchmark

`rid` using a random number for one segment of the ID is inherently slower than
an incrementing counter such as used in `xid`. That said, my 4-core laptop
can generate 4 million unique IDs in 1 second using a single process.

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

If you don't want the sortable semi-randomness this package provides, consider
the well tested and highly performant k-sortable `rs/xid` package upon which
`rid` is based. See https://github.com/rs/xid.

For a comparison of various golang unique ID solutions, including `rs/xid`, see:
https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html
