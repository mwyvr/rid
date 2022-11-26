[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/rid?tab=doc)[![Test](https://github.com/solutionroute/rid/actions/workflows/test.yaml/badge.svg)](https://github.com/solutionroute/rid/actions/workflows/test.yaml)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/rid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# rid

Package rid provides a (semi) random ID generator; rid generation is
goroutine-safe. A rid in string form looks like this: `ce0e7egs24nkzkn6egfg`.

The 12 byte binary ID encodes as a 20-character long, URL-friendly/Base32
encoded, mostly k-sortable (to the second resolution) identifier.

Each ID's 12-byte binary representation is comprised of a:

    - 4-byte timestamp value representing seconds since the Unix epoch
    - 2-byte machine ID
    - 2-byte process ID
    - 4-byte random value guaranteed to be unique for a given 
      timestamp+machine ID+process ID.

**Acknowledgement**: This package borrows heavily from the
[rs/xid](https://github.com/rs/xid) package which itself levers ideas from
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/). Where
this package differs is the use of admittedly slower random number generation
as opposed to a trailing counter for the last 4 bytes of the ID.

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

`rid` did not have ultra-high performance as an objective, as using random number
generation is inherently slower than an incrementing counter such as used in
`xid`. That said even my laptop can generate 1 million unique ids in less than a
second, and performance does not degrade significantly as core count increases.

AMD based Desktop with 8 cores/16 cpus:

    goos: linux
    goarch: amd64
    pkg: github.com/solutionroute/rid
    cpu: AMD Ryzen 7 3800X 8-Core Processor             

    $ go test -cpu 1 -benchmem  -run=^$   -bench  ^.*$
    BenchmarkIDNew        	 8099502	       178.6 ns/op	      13 B/op	       0 allocs/op
    BenchmarkIDNewEncoded 	^[ 6485542	       187.1 ns/op	       0 B/op	       0 allocs/op

    $ go test -cpu 8 -benchmem  -run=^$   -bench  ^.*$
    BenchmarkIDNew-8          	 2426989	       505.6 ns/op	      14 B/op	       0 allocs/op
    BenchmarkIDNewEncoded-8   	 2374066	       513.2 ns/op	       0 B/op	       0 allocs/op

    $ go test -cpu 16 -benchmem  -run=^$   -bench  ^.*$
    BenchmarkIDNew-16           	 2061738	       579.1 ns/op	      17 B/op	       0 allocs/op
    BenchmarkIDNewEncoded-16    	 2073908	       591.0 ns/op	       0 B/op	       0 allocs/op

Intel based Laptop with 4 cores/8 cpus:

    goos: linux
    goarch: amd64
    pkg: github.com/solutionroute/rid
    cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz

    $ go test -cpu 1 -benchmem  -run=^$   -bench  ^.*$ 
    BenchmarkIDNew        	 8768462	       178.2 ns/op	      12 B/op	       0 allocs/op
    BenchmarkIDNewEncoded 	 6655179	       179.5 ns/op	       1 B/op	       0 allocs/op

    $ go test -cpu 8 -benchmem  -run=^$   -bench  ^.*$ 
    BenchmarkIDNew-8          	 5793776	       216.9 ns/op	      18 B/op	       0 allocs/op
    BenchmarkIDNewEncoded-8   	 5421951	       214.7 ns/op	       0 B/op	       0 allocs/op

## See Also

If you don't want the randomness this package provides, consider the well
tested and highly performant xid package upon which rid is based. See
https://github.com/rs/xid.

For a comparison of various golang unique ID solutions, have a read:

https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

