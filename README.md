
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/solutionroute/sid)[![Build Status](https://travis-ci.org/solutionroute/sid.svg?branch=master)](https://travis-ci.org/solutionroute/sid)[![Go Coverage](http://gocover.io/_badge/github.com/solutionroute/sid)](http://gocover.io/github.com/solutionroute/sid)

# sid

Package sid provides a short ID generator producing relatively compact, unique
enough (65535 per millisecond), URL and human-friendly IDs.

The 8-byte ID itself is composed of:

- 6-byte timestamp value representing milliseconds since the Unix epoch
- 2-byte concurrency-safe counter (concurrency test included)

If for some reason your application needs to produce more than 65,535 new IDs
per _millisecond_ in any situation other than tests and benchmarks, this ID generator
is not the one you are looking for. May the force be with you!

String representations are chronologically [k-sortable](https://en.wikipedia.org/wiki/Partial_sorting), 
13 characters long and look like: `af1zwtepacw38`.

For readability purposes, the Base32 encoding of ID byte values uses a variant of the
[Crockford character set](https://www.crockford.com/base32.html) (omits i, l, o, u) rather than
the standard.

`cmd/sid` provides a simple tool to generate or inspect SIDs.

## Example Use

```go
package main

import (
    "fmt"
    "github.com/solutionroute/sid"
)

func main(){
    id := sid.New()
    fmt.Printf("ID: %s Timestamp (ms): %d Count: %5d \nBytes: %3v\n",
        id.String(), id.Milliseconds(), id.Count(), id[:])
}
// ID: af3fwdh337xx6 Timestamp (ms): 1590631922127 Count: 26430
// Bytes: [  1 114  89  12 249 207 103  62]
```

## Motivation

So why this? I had an itch to scratch, and an interest in looking at how ID
generation was being tackled for distributed applications. Having much less grand
needs, and wanting a shorter string representation (13 chars vs 20 or more), the
original-sounding "sid" was born.

## Acknowledgments

Much of this package was based on [rs/xid](https://github.com/rs/xid), which
itself was inspired by
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/).

[oklog/ulid](https://github.com/oklog/ulid)'s use of millisecond-resolution
timestamps was a good fit; independently also came to choose [Crockford's
Base32 character set](https://en.wikipedia.org/wiki/Base32#Crockford's_Base32)
over unsortable schemes like [Z-Base32](https://en.wikipedia.org/wiki/Base32#z-base-32) or
[HashIDs](https://github.com/speps/go-hashids).

Other inspriration was found in [Generating good unique IDs in
Go](https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html),
and reading the source of various packages offering more than this one does.
