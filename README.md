[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/rid?tab=doc)[![Build Status](https://travis-ci.org/solutionroute/rid.svg?branch=master)](https://travis-ci.org/solutionroute/rid)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/rid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# rid

Package rid provides a random-ish ID generator; the binary representation is 12
bytes long, the Base32-encoded representation is 20 characters long and
URL-friendly. The entropy component is a 4-byte unsigned random number with 4+
billion possibilities per second.

Acknowledgement: This package borrows heavily from the k-sortable rs/xid
package which itself levers ideas from mongodb. See https://github.com/rs/xid.


```go
    id := rid.New()
    fmt.Printf("%s", id) //  cdym59rs24a5g86efepg
```

## Under the covers

Each ID's 12-byte binary representation is comprised of a:

- 4-byte timestamp value representing seconds since the Unix epoch
- 2-byte machine ID
- 2-byte process ID
- 4-byte random value

IDs are chronologically sortable to the second, with a tradeoff in fine-grained
sortability due to the trailing random value component.

The String() representation is Base32 encoded using a modified Crockford
inspired alphabet.

## Batteries included

ID implements a number of common interfaces including package sql's
driver.Valuer, sql.Scanner, TextMarshaler, TextUnmarshaler, json.Marshaler,
json.Unmarshaler, and Stringer.

Package rid also provides a command line tool `rid` allowing for id generation or inspection:

    $ rid
    cdykpdgs2472sz29rxx0

    $ rid cdykpdgs2472sz29rxx0
    [cdykpdgs2472sz29rxx0] seconds:1669151542 entropy:4232693756 machine:[25] pid:4366 time:2022-11-22 13:12:22 -0800 PST \
    ID{99, 125, 59, 54, 25, 17, 14, 44, 252, 73, 199, 122}

    # generate and inspect a bunch
    $ rid `rid -c 2`
    [cdym9v0s24bnhqh7bawg] seconds:1669154028 entropy:3727121118 machine:[25] pid:4375 time:2022-11-22 13:53:48 -0800 PST \
    ID{99, 125, 68, 236, 25, 17, 23, 88, 222, 39, 90, 185}
    [cdym9v0s24bngggvqhw0] seconds:1669154028 entropy:1109113922 machine:[25] pid:4375 time:2022-11-22 13:53:48 -0800 PST \
    ID{99, 125, 68, 236, 25, 17, 23, 88, 66, 27, 188, 120}

## Source of inspiration

Thanks to the author of this article for turning me on to `xid` and other packages:

https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

## Acknowledgement

This package borrows heavily from the [rs/xid](https://github.com/rs/xid)
package which itself levers ideas from
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/).

