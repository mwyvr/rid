[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/sid?tab=doc)[![Build Status](https://travis-ci.org/solutionroute/sid.svg?branch=master)](https://travis-ci.org/solutionroute/sid)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/sid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# sid

Package sid provides a unique-enough, random-ish ID generator for applications
with modest (read: non-distributed), needs.

String representations of the 8-byte ID, e.g. `05yygjxjehg7y` are compact (13
characters), human-friendly (no i, l, o or u characters), double-clickable (no
'-' or other punctuation) and URL-safe.

```go
    id := sid.New()
    fmt.Printf("%s", id) // 05yygjxjehg7y
```

A sid ID is 8-byte value that can be stored directly as a 64 bit integer; some database
drivers will do just that - if that's not your preference, use id.String().

```go
    id := sid.New()     // 05yygjxjehg7y
    fmt.Println(id[:])  // [1 125 232 75 178 116 96 127]
    // reconstruct an ID from the encoded value
    nid, err := sid.FromString("05yygjxjehg7y") 
    nid == id           // true
```

## Under the covers

Each ID's 8-byte binary representation is comprised of a:

- 6-byte timestamp value representing milliseconds since the Unix epoch
- 2-byte concurrency-safe counter (test included); maxCounter = uint16(65535)

IDs are chronologically sortable to the millisecond.


## Collisions: not through intended use

The 2-byte concurrency-safe counter means a limit of 65,535 unique IDs per
millisecond (65 million a second), which translates to 1 ID every 16
nanoseconds, a limitation unlikely to be problematic in real life as ID
generation alone takes ~55ns on the author's hardware.

There's zero chance of collision in real world, intended, use.

## IDs are kinda randomish

The counter is **randomish** as it is initialized with a random value; where the
counter lands on any given millisecond can't be easily predicted. This allows for
a somewhat faster and definitely more concurrency safe solution, with no dupes being
produced by even 200 go routines.

    [05yyk5b963xka] ms:1640301422896 count:64309 <- counter at previous millisecond
    [05yyk5b967xkc] ms:1640301422897 count:64310
    .
    .
    .
    [05yyk5b967zzy] ms:1640301422897 count:65535
    [05yyk5b964002] ms:1640301422897 count:    1 <- same millisecond, counter safely rolls over
    [05yyk5b964004] ms:1640301422897 count:    2
    [05yyk5b964006] ms:1640301422897 count:    3

## Benchmark

As expected, reality means about 20 million IDs are generated *per second* in
the simplest of use cases, a for loop benchmark generating nothing but new IDs:

    $ go test -benchmem -benchtime 1s  -run=^$ -bench ^BenchmarkIDNew$ github.com/solutionroute/sid
    goos: linux
    goarch: amd64
    pkg: github.com/solutionroute/sid
    cpu: AMD Ryzen 7 3800X 8-Core Processor
    BenchmarkIDNew-16       20454786                59.28 ns/op            0 B/op          0 allocs/op

## Batteries included

ID implements a number of common interfaces including package sql's
driver.Valuer, sql.Scanner, TextMarshaler, TextUnmarshaler, json.Marshaler,
json.Unmarshaler, and Stringer.

Package sid also provides a command line tool `sid` allowing for id generation or inspection:

    $ sid
    05yygjxjehg7y

    $ sid 05yygjxjehg7y
    [05yygjxjehg7y] ms:1640279814772 count:24703 time:2021-12-23 09:16:54.772 -0800 PST id:{1, 125, 232, 75, 178, 116, 96, 127}

    # generate more than 1
    $ sid -c 3
    05yyjfzbmf71p 05yyjfzbmf71r 05yyjfzbmf71t

    # generate and inspect a bunch
    $ sid `sid -c 3`
    [05yyjfzbmf71p] ms:1640295820195 count:52763 time:2021-12-23 13:43:40.195 -0800 PST id:{1, 125, 233, 63, 235, 163, 206, 27}
    [05yyjfzbmf71r] ms:1640295820195 count:52764 time:2021-12-23 13:43:40.195 -0800 PST id:{1, 125, 233, 63, 235, 163, 206, 28}
    [05yyjfzbmf71t] ms:1640295820195 count:52765 time:2021-12-23 13:43:40.195 -0800 PST id:{1, 125, 233, 63, 235, 163, 206, 29}

    # with newlines
    $ sid -c 3 -n
    05yyjga5cra1a
    05yyjga5cra1c
    05yyjga5cra1e

## Source of inspiration

Thanks to the author of this article for giving me inspiration:

https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

Borrowing from that article, here's a comparison with some other ID schemes:

    github.com/solutionroute/sid        05yygjxjehg7y
    github.com/rs/xid:                  9bsv0s091sd002o20hk0
    github.com/segmentio/ksuid:         ZJkWubTm3ZsHZNs7FGt6oFvVVnD
    github.com/kjk/betterguid:          -HIVJnL-rmRZno06mvcV
    github.com/oklog/ulid:              014KG56DC01GG4TEB01ZEX7WFJ
    github.com/chilts/sid:              1257894000000000000-4601851300195147788
    github.com/lithammer/shortuuid:     DWaocVZPEBQB5BRMv6FUsZ
    github.com/google/uuid:             fa931eb3-cdc7-46a1-ae94-eb1b523203be

## Acknowledgement

This package is largely based on the globally-unique capable
[rs/xid](https://github.com/rs/xid) package which itself levers ideas from
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/).

I'll use `xid` if I ever have apps on machines spread around the world working
without central coordinated ID generation.
