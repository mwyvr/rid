[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/sid?tab=doc)[![Build Status](https://travis-ci.org/solutionroute/sid.svg?branch=master)](https://travis-ci.org/solutionroute/sid)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/sid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# sid

Package sid provides a unique ID generator producing URL and human-friendly
(readability and double-click), compact, IDs. They are unique to a single
process, with more than 4 billion possibilities per millisecond.

The String() method produces chronologically
[k-sortable](https://en.wikipedia.org/wiki/Partial_sorting)) encoded IDs that
look like:

    af87cfy46ajbxf40 - 16 characters, and is equivalent to:
    []byte{001, 125, 209, 022, 154, 224, 016, 025, 151, 086}

`sid` base32 encoding utilizes a customized alphabet, popularized by
[Crockford](http://www.crockford.com/base32.html), who replaced the more easily
misread (by humans) i, o, l, and u with the more easily read w, x, y, z.
Additionally, `sid` encoding has digits moved to the tail of the character set
to avoid having a leading zero for a great many years.

Each ID's 10-byte binary representation is comprised of a:

    6-byte timestamp value representing milliseconds since the Unix epoch
    4-byte concurrency-safe counter (test included); maxCounter = uint32(4294967295)

The counter is initialized at a random value at initialization.

ID implements a number of common interfaces including package sql's
driver.Valuer, sql.Scanner, TextMarshaler, TextUnmarshaler, json.Marshaler,
json.Unmarshaler, and Stringer.

## Inspiration

COVID-19 bordom in 2020. Original source of inspiration:
https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

## Acknowledgement

Much of this package was based on the globally-unique capable
[rs/xid](https://github.com/rs/xid) package which itself levers ideas from
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/).

I'd likely use xid if I had apps on machines spread around the world working in
unison on a common datastore.

[Generating good unique IDs in
Go](https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html)
provided additional inspiration.

## Encoded ID comparisons with other packages

    github.com/solutionroute/sid/v3:    af87cfy46ajbxf40
    github.com/rs/xid:                  9bsv0s091sd002o20hk0
    github.com/segmentio/ksuid:         ZJkWubTm3ZsHZNs7FGt6oFvVVnD
    github.com/kjk/betterguid:          -HIVJnL-rmRZno06mvcV
    github.com/oklog/ulid:              014KG56DC01GG4TEB01ZEX7WFJ
    github.com/chilts/sid:              1257894000000000000-4601851300195147788
    github.com/lithammer/shortuuid:     DWaocVZPEBQB5BRMv6FUsZ
    github.com/google/uuid:             fa931eb3-cdc7-46a1-ae94-eb1b523203be

## Batteries included

`cmd/sid` provides a simple tool to generate or inspect SIDs.
    # generate an ID
    $ sid
    af87av3z734qnx8y

    # inspect an ID
    $ sid af87av3z734qnx8y

    # while away your COVID-19 days looking at milliseconds pass by...
    $ sid `sid`

## Example Use

```go
package main

import (
   "fmt"
    "github.com/solutionroute/sid"
)

func main() {
    id := sid.New() // ids are []byte values 6 bytes time, 4 bytes counter
    fmt.Printf(`ID:
        String()       %s   // af87av3z734qnx8y 
        Milliseconds() %d   // 1639876867566
        Count()        %d   // 38307
        Time()         %v   // 2021-12-18 17:21:07.566 -0800 PST
        Bytes():       %3v  // [1 125 208 71 53 238 116 213 207 212]`, 
        id.String(), id.Milliseconds(), id.Count(), id.Time(), id.Bytes())

    if id, err := sid.FromString("af87av3z734qnx8y"); err == nil {
        fmt.Printf("ID: %s Timestamp (ms): %d Count: %d\n", id, id.Milliseconds(), id.Count())
    } 
}
```
