[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/rid?tab=doc)[![Build Status](https://travis-ci.org/solutionroute/rid.svg?branch=master)](https://travis-ci.org/solutionroute/rid)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/rid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# rid

**WORK IN PROGRESS, Nov 23 2022 please check back later**

Package rid provides a random ID generator. The 12 byte binary ID encodes as a
20-character long, URL-friendly/Base32 encoded, mostly k-sortable (to the
second resolution) identifier.

Each ID's 12-byte binary representation is comprised of a:

    - 4-byte timestamp value representing seconds since the Unix epoch
    - 2-byte machine ID
    - 2-byte process ID
    - 4-byte random value with 4,294,967,295 possibilities guaranteed to be
      unique for a given [timestamp|machine ID|process ID].

**Acknowledgement**: This package borrows heavily from the
[rs/xid](https://github.com/rs/xid) package which itself levers ideas from
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/). Where
this package differs is the use of admittedly slower random number generation
as opposed to a trailing counter for the last 4 bytes of the ID.

## Usage


```go
    id := rid.New()
    fmt.Printf("%s", id) //  cdym59rs24a5g86efepg
```

## Batteries included

rid.ID implements a number of common interfaces including package sql's
driver.Valuer, sql.Scanner, TextMarshaler, TextUnmarshaler, json.Marshaler,
json.Unmarshaler, and Stringer.

Package rid also provides a command line tool `rid` allowing for id generation
or inspection:

    $ rid
    cdz2kt8s25hpv44k214g
    $rid `rid`
    [cdz2ktgs25hq5ysrgg0g] seconds:1669212650 random:4214785275 machine:[25 17] pid:25458 time:2022-11-23 06:10:50 -0800 PST ID{99, 126, 41, 234, 25, 17, 99, 114, 251, 56, 132, 1}
    $ rid `rid`
    [cdz2kvgs25hqstfgez00] seconds:1669212654 random:3924850665 machine:[25 17] pid:25468 time:2022-11-23 06:10:54 -0800 PST ID{99, 126, 41, 238, 25, 17, 99, 124, 233, 240, 119, 192}


## See Also

If ~ 400-500ns/op is too slow and/or you don't need the randomness this package
seeks to provide, consider the well tested and highly performant xid package.
See https://github.com/rs/xid.

For a comparison of various unique ID solutions, have a read:

https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

