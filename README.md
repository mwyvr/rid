[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/sid?tab=doc)[![Build Status](https://travis-ci.org/solutionroute/sid.svg?branch=master)](https://travis-ci.org/solutionroute/sid)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/sid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# sid

Package sid provides a unique-enough, random-ish ID generator for applications
with modest (read: non-distributed), needs. String representations of the ID are
compact (13 characters), human-friendly (no i,l,o or u characters),
double-clickable (no '-' or other punctuation) and URL-safe.

IDs are mostly chronologically sortable with a minor tradeoff made for improved
randomness in the trailing counter value.

```go
    id := sid.New()
    fmt.Printf("%s", id) // af1zwtepacw38
```

A sid ID is 8-byte value that can be stored directly as a 64 bit integer; some database
drivers will do just that - if that's not your preference, use id.String().

```go
    id := sid.New()     // af1zwtepacw38
    fmt.Println(id[:])  // [1 125 227 253 59 110 47 62]
    // reconstruct an ID from the encoded value
    nid, err := sid.FromString("af1zwtepacw38") 
    nid == id           // true
```

## Motivation: modest needs

sid was intended for single process, single machine apps such as might use Go
friendly datastores like BoltDB, Badger or abstractions on top of either like
Genji | bolthold | badgerhold or other single-connection-only document or
key-value datastores.

## Under the covers

Each ID's 8-byte binary representation: id{1, 111, 89, 64, 140, 0, 165, 159} is
comprised of a:

- 6-byte timestamp value representing milliseconds since the Unix epoch
- 2-byte concurrency-safe counter (test included); maxCounter = uint16(65535)

## Collisions: not through intended use

The 2-byte concurrency-safe counter means up to 65,535 unique IDs can
theoretically be produced per millisecond - that's 1 ID every 16 nanoseconds.

We say theoretically because on the author's hardware it takes ~50ns to produce
an ID, another 50-90ns to encode it depending on the encoder, and longer yet to
shove the associated data into a datastore. This means there's zero chance of
collision in real world, intended, use.

## IDs are randomish

The counter is **randomish** as it is initialized with a random value and
thereafter at any new millisecond an ID is requested. This is intended to
dissuade URL parameter hackers... but it's random-ish, so don't use sid.ID for a
secure token (**that's not an intended use**)! Still, that's 65 million
*potential* IDs per second, but more likely **only** several million randomish
IDs per second in the real world.

    af88je3v03f7p
    af88je3v03f7r
    af88je3v03f7t
    af88je3v08n1r <- new millisecond, counter re-initialized with a random number
    af88je3v08n1t
    af88je3v08n1w
    af88je3v08n1y

## Batteries included

ID implements a number of common interfaces including package sql's
driver.Valuer, sql.Scanner, TextMarshaler, TextUnmarshaler, json.Marshaler,
json.Unmarshaler, and Stringer.

Package sid also provides a command line tool `sid` allowing for id generation or inspection:

    $ sid
    af88jfz84bx5j

    $ sid af1zwtepacw38
    [af1zwtepacw38] ms:1577750400000 count:42399 time:2019-12-30 16:00:00 -0800 PST id:{1, 111, 89, 64, 140, 0, 165, 159}

    # generate more than 1
    $ sid -c 3
    af88jgn52y7c0 af88jgn52y7c2 af88jgn52y7c4

    # generate and inspect a bunch
    $ sid `sid -c 3`
    [af88jgpv71728] ms:1640209420781 count:64399 time:2021-12-22 13:43:40.781 -0800 PST id:{1, 125, 228, 25, 145, 237, 251, 143}
    [af88jgpv7173a] ms:1640209420781 count:64400 time:2021-12-22 13:43:40.781 -0800 PST id:{1, 125, 228, 25, 145, 237, 251, 144}
    [af88jgpv7173c] ms:1640209420781 count:64401 time:2021-12-22 13:43:40.781 -0800 PST id:{1, 125, 228, 25, 145, 237, 251, 145}

    # with newlines
    $ sid -c 3 -n
    af88jgqnp4nkp
    af88jgqnp4nkr
    af88jgqnp4nkt

## Source of inspiration

Acknowledgement: Much of this package is based on the globally-unique capable
[rs/xid](https://github.com/rs/xid) package which itself levers ideas from mongodb. See
https://github.com/rs/xid. I'd use xid if I had a fleet of apps on machines
spread around the world working in unison on a common datastore.

Thanks to the author of this article for giving me inspiration:

https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

Borrowing data from that article, here's a comparison with some other ID schemes:

    github.com/solutionroute/sid        af1zwtepacw38
    github.com/rs/xid:                  9bsv0s091sd002o20hk0
    github.com/segmentio/ksuid:         ZJkWubTm3ZsHZNs7FGt6oFvVVnD
    github.com/kjk/betterguid:          -HIVJnL-rmRZno06mvcV
    github.com/oklog/ulid:              014KG56DC01GG4TEB01ZEX7WFJ
    github.com/chilts/sid:              1257894000000000000-4601851300195147788
    github.com/lithammer/shortuuid:     DWaocVZPEBQB5BRMv6FUsZ
    github.com/google/uuid:             fa931eb3-cdc7-46a1-ae94-eb1b523203be

`sid` base32 encoding utilizes a customized alphabet, popularized by
[Crockford](http://www.crockford.com/base32.html), who replaced the more easily
misread (by humans) i, o, l and u with the more easily read w, x, y and z.
Additionally, `sid` customized encoding has digits moved to the tail of the
character set to avoid having a leading zero for a great many years.

Each ID's 10-byte binary representation is comprised of a:

    6-byte timestamp value representing milliseconds since the Unix epoch
    4-byte concurrency-safe counter (test included); maxCounter = uint32(4294967295)

The counter is initialized at a random value at initialization.

## Acknowledgement

Much of this package was based on the globally-unique capable
[rs/xid](https://github.com/rs/xid) package which itself levers ideas from
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/).

Having borrowed heavily from it, I'd likely use `xid` if I had apps on machines
spread around the world working in unison on a common datastore.
