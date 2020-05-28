# sid

Package sid provides a short ID generator producing relatively compact, unique
enough (65535 per millisecond), URL and human-friendly IDs.

The eight-byte ID itself is composed of:

- 6-byte timestamp value representing milliseconds since the Unix epoch
- 2-byte counter 0-65535 that rolls over when it hits maximum.

If for some reason your application needs to produce more than 65,535 new IDs
per _millisecond_ in any situation other than tests and benchmarks, this ID generator
is not the one you are looking for. May the force be with you!

String representations look like:

    af1zwtepacw38 // 13 characters long
    
For readability purposes, the Base32 encoding of ID byte values uses the
[Crockford character set](https://www.crockford.com/base32.html) rather than
the standard.

ID generation is concurrency safe.

The package provides implementations of some well-known interfaces for encoding and SQL.

Under construction, May 2020, it's pandemic time.

## Example Use

    ```go
    package main

	import (
		"fmt"
		"github.com/solutionroute/sid"
	)

	func main(){
		id := sid.New()
		fmt.Printf("ID: %s Timestamp (ms): %d Count: %5d Bytes: %3v\n",
			id.String(), id.Milliseconds(), id.Count(), id[:])
	}
    ```

## Acknowldegment

Much of this package was based on [rs/xid](https://github.com/rs/xid), 
which itself was inspired by [MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/).

So why this? I had an itch to scratch, an interest in looking at how ID
generation was being tackled for distributed applications, but much less grand
needs for myself. Mostly I wanted a shorter string representation - sid.IDs are
13 characters as opposed to 20, or 24, respectively. and was just interested in
looking at the problem.

Other inspriration was found in [Generating good unique IDs in
Go](https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html),
and reading the source of various packages offering more than this one does. I
also looked at [HashIDs](https://github.com/speps/go-hashids) but opted for the
fixed length nature of Base32 instead.
