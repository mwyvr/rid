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
this package differs is the use of a slightly optimized but admittedly slower
random number generation as opposed to xid's use of a simple counter for the
last 4 bytes of the ID.

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

`rid` using random number
generation is inherently slower than an incrementing counter such as used in
`xid`. That said even my laptop can generate 1 million unique ids in less than a
second, and performance does not degrade significantly as core count increases.

See [benchmark.md](./benchmark.md) for more.

## See Also

If you don't want the randomness this package provides, consider the well
tested and highly performant xid package upon which rid is based. See
https://github.com/rs/xid.

For a comparison of various golang unique ID solutions, have a read:

https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

