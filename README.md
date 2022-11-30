[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/rid?tab=doc)[![Test](https://github.com/solutionroute/rid/actions/workflows/test.yaml/badge.svg)](https://github.com/solutionroute/rid/actions/workflows/test.yaml)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/rid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# rid

Package rid provides a [k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence),
zero-configuration, unique ID generator.  Binary IDs are encoded as Base32,
producing a 20-character URL-friendly representation like: `ce0e7egs24nkzkn6egfg`.

The 12-byte binary representation of an ID is comprised of a:

- 4-byte timestamp value representing seconds ticked since the Unix epoch
- 2-byte machine identifier
- 2-byte process identifier
- 4-byte crypto/rand generated random value verified per tick for uniqueness

Including the machine and process info into an ID makes `rid` potentially
suitable, without need for configuration or coordination, for distributed
applications. Your use case may vary.

rid implements a number of well-known interfaces to make use with json
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
    fmt.Printf("%s", id) // ce0e7egs24nkzkn6egfg
```

## Batteries included

`rid.ID` implements a number of common interfaces including:

- database/sql: driver.Valuer, sql.Scanner
- encoding: TextMarshaler, TextUnmarshaler
- encoding/json: json.Marshaler, json.Unmarshaler
- Stringer

Package `rid` also provides a command line tool `rid` allowing for id generation
and inspection. To install: `go install github.com/solutionroute/rid/...`

    $ rid
    ce0e7ygs24nw4zebrz10

    $ rid -c 2
    ce0e8n0s24p7329f3gfg
    ce0e8n0s24p73q9hazp0

    # produce 4 and inspect
    $rid `rid -c 4`
    ce3r1ars24zj69nxftj0 seconds:1669824683 machine:[0x19,0x11] pid:16163 random: 649952806 | time:2022-11-30 08:11:23 -0800 PST ID{0x63,0x87,0x80,0xab,0x19,0x11,0x3f,0x23,0x26,0xbd,0x7e,0xa4}
    ce3r1ars24zj7xwwc3k0 seconds:1669824683 machine:[0x19,0x11] pid:16163 random:4154220791 | time:2022-11-30 08:11:23 -0800 PST ID{0x63,0x87,0x80,0xab,0x19,0x11,0x3f,0x23,0xf7,0x9c,0x60,0xe6}
    ce3r1ars24zj6016apjg seconds:1669824683 machine:[0x19,0x11] pid:16163 random:   2512128 | time:2022-11-30 08:11:23 -0800 PST ID{0x63,0x87,0x80,0xab,0x19,0x11,0x3f,0x23,0x0,0x26,0x55,0xa5}
    ce3r1ars24zj7dkzbd80 seconds:1669824683 machine:[0x19,0x11] pid:16163 random:3061799862 | time:2022-11-30 08:11:23 -0800 PST ID{0x63,0x87,0x80,0xab,0x19,0x11,0x3f,0x23,0xb6,0x7f,0x5b,0x50}

## Package Comparisons

| Package                                                   |BLen|ELen| K-Sort| 0-Cfg | Encoded ID                           | Method     | Components |
|-----------------------------------------------------------|----|----|-------|-------|--------------------------------------|------------|------------|
| [solutionroute/rid](https://github.com/solutionroute/rid) | 12 | 20 |  true |  true | ce3vsz0s24fn979qfjpg                 | crypt/rand | ts(seconds) : machine ID : process ID : random |
| [rs/xid](https://github.com/rs/xid)                       | 12 | 20 |  true |  true | ce3rpv0p26gdpm40gbv0                 | counter    | ts(seconds) : machine ID : process ID : counter |
| [segmentio/ksuid](https://github.com/segmentio/ksuid)     | 20 | 27 |  true |  true | 2IHYlFPNznxhMcMpdi4ppCtwJWZ          | random     | ts(seconds) : random |
| [google/uuid](https://github.com/google/uuid)             | 16 | 36 | false |  true | db5507af-6a9c-40ea-899b-0fe3c547086e | crypt/rand | (v4) version + variant + 122 bits random |
| [oklog/ulid](https://github.com/oklog/ulid)               | 16 | 26 |  true |  true | 01GK53ME5694KZW2NS79RK70BT           | crypt/rand | ts(ms) : choice of random |
| [kjk/betterguid](https://github.com/kjk/betterguid)       | 20 | 20 |  true |  true | -NI9DYXaHaA4RFWy_R1l                 | counter    | ts(ms) + per-ms math/rand initialized counter |

If you don't need the k-sortable randomness this and other packages provide,
consider the well-tested and performant k-sortable `rs/xid` package
upon which `rid` is based. See https://github.com/rs/xid.

For a detailed comparison of various golang unique ID solutions, including `rs/xid`, see:
https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

## Package Benchmarks

Benchmarks were purposely left until last in this README. `rid` was developed to
provide a more random unique ID than `xid` produces; to do so it relies on
`crypto/rand` plus a uniqueness checking algorithm which involves a map and
mutex locks, with a predictable hit on performance.

Nevertheless, `rid` remains performant enough for many use cases. See 
[bench/bench_test.go](bench/bench_test.go).

### Intel 4-core Dell Latitude 7420 laptop

    $ go test -cpu 1,2,8 -benchmem  -run=^$   -benchtime 1s -bench  ^.*$ 
    goos: linux
    goarch: amd64
    pkg: github.com/solutionroute/rid/bench
    cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz
    BenchmarkRid            	 3132163	       440.2 ns/op	      17 B/op	       0 allocs/op
    BenchmarkRid-2          	 2332372	       501.2 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-8          	 2434858	       489.9 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid            	30396781	        37.74 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-2          	35541860	        34.44 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-8          	74413412	        16.47 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid          	 3705994	       315.3 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-2        	 3211878	       368.7 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-8        	 3229306	       364.9 ns/op	       0 B/op	       0 allocs/op
    BenchmarkGoogleUuid     	 4200759	       285.5 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-2   	 6372727	       186.4 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-8   	11384917	       124.4 ns/op	      16 B/op	       1 allocs/op
    BenchmarkUlid           	  159339	      7436 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-2         	  229455	      4741 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-8         	  639322	      1806 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkBetterguid     	13666479	        80.71 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-2   	11158867	       103.4 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-8   	 6611118	       173.9 ns/op	      24 B/op	       1 allocs/op
