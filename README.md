[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/rid?tab=doc)[![Test](https://github.com/solutionroute/rid/actions/workflows/test.yaml/badge.svg)](https://github.com/solutionroute/rid/actions/workflows/test.yaml)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/rid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# rid

Package `rid` provides a [k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence),
zero-configuration, unique ID generator.  Binary IDs are encoded as Base32,
producing a 20-character URL-friendly representation like: `ce7cjjn0vwjj89nerxag`.

The 12-byte binary representation of an ID is comprised of a:

- 4-byte timestamp value representing seconds ticked since the Unix epoch
- 2-byte machine+process signature, derived from a md5 hash of the machine ID + process ID
- 6-byte random number using Go's runtime `fastrand` function. [1]

`rid` also implements a number of well-known interfaces to make use with json
and databases more convenient.

**Acknowledgement**: This package borrows _heavily_ from the at-scale capable
[rs/xid](https://github.com/rs/xid) package which itself levers ideas from
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/).

Where this package primarily differs is the use of 6-byte random numbers as 
opposed to xid's use of a monotonic counter for the last 4 bytes of the ID.

[1] For more information on fastrand (wyrand) see: https://github.com/wangyi-fudan/wyhash
 and [Go's sources for runtime/stubs.go](https://cs.opensource.google/go/go/+/master:src/runtime/stubs.go;bpv=1;bpt=1?q=fastrand&ss=go%2Fgo:src%2Fruntime%2F).

## Usage

```go
	i := rid.New()
	fmt.Printf("%s\n", i)           // ce7cq2h7m59ymhyny5f0
	fmt.Printf("%v\n", i[:])        // [99 142 203 138 39 161 83 234 71 213 241 94]
	res, _ := i.MarshalJSON()
	fmt.Printf("%s\n", res)         // "ce7cq2h7m59ymhyny5f0"
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
    ce7cjjn0vwjj89nerxag

    $ rid -c 2
    ce7cjm5zmaza1s8ef8g0
    ce7cjm5zma56p6t0pjm0

    # produce 4 and inspect
    $rid `rid -c 4`
    ce7cczpvebcrq17kydhg seconds:1670301310 rtsig:[0xdb,0x72] random:239193254261603 | time:2022-12-05 20:35:10 -0800 PST ID{0x63,0x8e,0xc6,0x7e,0xdb,0x72,0xd9,0x8b,0x84,0xf3,0xf3,0x63}
    ce7cczpve8fnkc0d68wg seconds:1670301310 rtsig:[0xdb,0x72] random: 34470066205241 | time:2022-12-05 20:35:10 -0800 PST ID{0x63,0x8e,0xc6,0x7e,0xdb,0x72,0x1f,0x59,0xb0,0xd,0x32,0x39}
    ce7cczyveas70ayma7g0 seconds:1670301311 rtsig:[0xdb,0x72] random:196194841416160 | time:2022-12-05 20:35:11 -0800 PST ID{0x63,0x8e,0xc6,0x7f,0xdb,0x72,0xb2,0x70,0x2b,0xd4,0x51,0xe0}
    ce7cczyve8jp1x5d13e0 seconds:1670301311 rtsig:[0xdb,0x72] random: 41098352068828 | time:2022-12-05 20:35:11 -0800 PST ID{0x63,0x8e,0xc6,0x7f,0xdb,0x72,0x25,0x60,0xf4,0xad,0x8,0xdc}

## Random Source

For random number generation `rid` uses a Go runtime `fastrand64`, available in
Go versions released post-spring 2022; it's non-deterministic, goroutine safe, 
and fast.  For the purpose of *this* package, `fastrand64` seems ideal.

## Package Comparisons

| Package                                                   |BLen|ELen| K-Sort| 0-Cfg | Encoded ID                           | Method     | Components |
|-----------------------------------------------------------|----|----|-------|-------|--------------------------------------|------------|------------|
| [solutionroute/rid](https://github.com/solutionroute/rid) | 12 | 20 |  true |  true | ce3vsz0s24fn979qfjpg                 | fastrand   | ts(seconds) : runtime signature : random |
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

A comparison with the above noted packages can be found in [bench/bench_test.go](bench/bench_test.go). Output:

### Intel 4-core Dell Latitude 7420 laptop


### AMD 8-core desktop

    $ go test -cpu 1,2,4,8,16 -benchmem  -run=^$   -bench  ^.*$
    goos: linux
    goarch: amd64
    pkg: github.com/solutionroute/rid/bench
    cpu: AMD Ryzen 7 3800X 8-Core Processor             
    BenchmarkRid              	22546425	        52.57 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-2            	44619606	        26.36 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-4            	76766934	        13.51 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-8            	171874088	         6.869 ns/op	   0 B/op	       0 allocs/op
    BenchmarkRid-16           	305219312	         3.963 ns/op	   0 B/op	       0 allocs/op
    BenchmarkXid              	22564863	        51.45 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-2            	11812347	       102.3 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-4            	24562400	        52.75 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-8            	50628301	        33.06 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-16           	69468259	        17.08 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid            	 3238129	       363.5 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-2          	 1558274	       811.4 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-4          	 1453086	       836.2 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-8          	 1413405	       837.1 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-16         	 1371385	       861.7 ns/op	       0 B/op	       0 allocs/op
    BenchmarkGoogleUuid       	 3394983	       385.8 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-2     	 4834682	       209.6 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-4     	 9113331	       110.7 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-8     	17528270	        59.86 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-16    	29063694	        40.57 ns/op	      16 B/op	       1 allocs/op
    BenchmarkUlid             	  144672	      7925 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-2           	  277130	      4259 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-4           	  473964	      2330 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-8           	  798924	      1445 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-16          	  792290	      1479 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkBetterguid       	14279642	        81.82 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-2     	 7232544	       141.4 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-4     	 4828852	       274.6 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-8     	 4040710	       305.8 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-16    	 3563704	       366.4 ns/op	      24 B/op	       1 allocs/op

