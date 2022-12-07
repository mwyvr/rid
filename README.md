[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/rid?tab=doc)[![Test](https://github.com/solutionroute/rid/actions/workflows/test.yaml/badge.svg)](https://github.com/solutionroute/rid/actions/workflows/test.yaml)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/rid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# rid

Package `rid` provides a [k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence),
zero-configuration, unique ID generator. Binary IDs are Base32-encoded,
producing a 24-character case-insensitive URL-friendly representation like:
`062ekgz5k5f23ejagw2n7c9f`.

Base32 encoding evenly aligns with 15 byte / 120 bit binary data. The 15-byte
binary representation of an ID is comprised of a:

- 6-byte timestamp value representing milliseconds since the Unix epoch
- 1-byte machine+process signature, derived from md5(machine ID + process ID)
- 6-byte random number using Go's runtime `fastrand` function. [1]

`rid` also implements a number of well-known interfaces to make use with JSON
and databases more convenient.

**Acknowledgement**: This package borrows _heavily_ from the at-scale capable
[rs/xid](https://github.com/rs/xid) package which itself levers ideas from
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/).

Where this package differs, rid (15 bytes) | xid (12 bytes):

- 6-bytes of time, millisecond resolution | 4 bytes, second resolution
- 1-byte machine+process signature | 3 bytes machine ID, 2 bytes process ID
- 6-byte random number | 3-byte monotonic counter randomly initialized once 

## Usage

```go
	i := rid.New()
	fmt.Printf("%s\n", i)           // 062ekkxhmp31522vfjt7jv9t 
```

## Batteries included

`rid.ID` implements a number of common interfaces including:

- database/sql: driver.Valuer, sql.Scanner
- encoding: TextMarshaler, TextUnmarshaler
- encoding/json: json.Marshaler, json.Unmarshaler
- Stringer

Package `rid` also provides a command line tool `rid` allowing for id generation
and inspection. To install: `go install github.com/solutionroute/rid/cmd/...`

    $ rid 
    062ekjasgt18j0xgabq5zw45

    $ rid -c 2
    062ekjdxbc4yr0v0zyhv19zb
    062ekjdxbc4pesrn45jfz89k

    # produce 4 and inspect
    $rid `rid -c 4`
    062ekjn39b2g7mvzwsxk2mx9 ts:1670369682250 rtsig:[0xc5] random:  4206918794033 | time:2022-12-06 15:34:42.25 -0800 PST ID{0x1,0x84,0xe9,0xca,0xa3,0x4a,0xc5,0x3,0xd3,0x7f,0xe6,0x7b,0x31,0x53,0xa9}
    062ekjn39b2tex8f39ht2vxk ts:1670369682250 rtsig:[0xc5] random:184121206399905 | time:2022-12-06 15:34:42.25 -0800 PST ID{0x1,0x84,0xe9,0xca,0xa3,0x4a,0xc5,0xa7,0x75,0xf,0x1a,0x63,0xa1,0x6f,0xb3}
    062ekjn39b2n2km1wn6qzaty ts:1670369682250 rtsig:[0xc5] random: 89397628587391 | time:2022-12-06 15:34:42.25 -0800 PST ID{0x1,0x84,0xe9,0xca,0xa3,0x4a,0xc5,0x51,0x4e,0x81,0xe5,0x4d,0x7f,0xab,0x5e}
    062ekjn39b2vxg1h326m5z9w ts:1670369682250 rtsig:[0xc5] random:209732666690882 | time:2022-12-06 15:34:42.25 -0800 PST ID{0x1,0x84,0xe9,0xca,0xa3,0x4a,0xc5,0xbe,0xc0,0x31,0x18,0x8d,0x42,0xfd,0x3c}

## Random Source

For random number generation `rid` uses a Go runtime `fastrand64` [1],
available in Go versions released post-spring 2022; it's non-deterministic,
goroutine safe, and fast.  For the purpose of *this* package, `fastrand64`
seems ideal.

Use of `fastrand` makes `rid` performant and scales well as cores/parallel
processes are added. While more testing will be done, no ID collisions have
been observed over numerous runs producing upwards of 300 million ID using
single and multiple goroutines.

[1] For more information on fastrand (wyrand) see: https://github.com/wangyi-fudan/wyhash
 and [Go's sources for runtime/stubs.go](https://cs.opensource.google/go/go/+/master:src/runtime/stubs.go;bpv=1;bpt=1?q=fastrand&ss=go%2Fgo:src%2Fruntime%2F).

## Package Comparisons

| Package                                                   |BLen|ELen| K-Sort| 0-Cfg | Encoded ID and Next | Method | Components |
|-----------------------------------------------------------|----|----|-------|-------|---------------------|--------|------------|
| [solutionroute/rid](https://github.com/solutionroute/rid) | 15 | 24 |  true |  true | `062ejz2nn8sm19eqhaj4h97w`<br>`062ejz2nn8sm7c72aywz5gas` | fastrand | ts(seconds) : runtime signature : random |
| [rs/xid](https://github.com/rs/xid)                       | 12 | 20 |  true |  true | `ce7rr1gp26gbkqlf7kp0`<br>`ce7rr1gp26gbkqlf7kpg` | counter | ts(seconds) : machine ID : process ID : counter |
| [segmentio/ksuid](https://github.com/segmentio/ksuid)     | 20 | 27 |  true |  true | `2IYhiP9ndQPRZYmMnvQ5sq9JXw8`<br>`2IYhiNdG6sj9qvaUO7unG1UsgZ2` | random | ts(seconds) : random |
| [google/uuid](https://github.com/google/uuid)             | 16 | 36 | false |  true | `75eb752d-d1d0-4000-994e-fbce08743687`<br>`d6873ae4-2240-46d6-8b83-42b21a55125f` | crypt/rand | (v4) version + variant + 122 bits random |
| [oklog/ulid](https://github.com/oklog/ulid)               | 16 | 26 |  true |  true | `01GKMQRNDAQCWFR7WJA8QX8V1R`<br>`01GKMQRNDA9Q99KRYY2K8F827Q` | crypt/rand | ts(ms) : choice of random |
| [kjk/betterguid](https://github.com/kjk/betterguid)       | 20 | 20 |  true |  true | `-NIdU4Le0mBzlgH0A87M`<br>`-NIdU4Le0mBzlgH0A87N` | counter | ts(ms) + per-ms math/rand initialized counter |

If you don't need the k-sortable randomness this and other packages provide,
consider the well-tested and performant k-sortable `rs/xid` package
upon which `rid` is based. See https://github.com/rs/xid.

For a detailed comparison of various golang unique ID solutions, including `rs/xid`, see:
https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

## Package Benchmarks

A comparison with the above noted packages can be found in [bench/bench_test.go](bench/bench_test.go). Output:

### Intel 4-core Dell Latitude 7420 laptop

    $ go test -cpu 1,2,8 -benchmem  -run=^$   -benchtime 1s -bench  ^.*$ 
    goos: linux
    goarch: amd64
    pkg: github.com/solutionroute/rid/bench
    cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz
    BenchmarkRid            	27389251	        41.95 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-2          	55504586	        22.52 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-8          	131805276	         9.059 ns/op	   0 B/op	       0 allocs/op
    BenchmarkXid            	31823905	        36.99 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-2          	37957798	        31.98 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-8          	70593487	        16.82 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid          	 3749377	       324.4 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-2        	 3287676	       367.3 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-8        	 3296826	       365.2 ns/op	       0 B/op	       0 allocs/op
    BenchmarkGoogleUuid     	 4289882	       284.5 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-2   	 6151603	       217.7 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-8   	 8814963	       131.8 ns/op	      16 B/op	       1 allocs/op
    BenchmarkUlid           	  150135	      7539 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-2         	  235570	      4785 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-8         	  558735	      2098 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkBetterguid     	14361985	        82.00 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-2   	11374424	       101.6 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-8   	 7073932	       167.7 ns/op	      24 B/op	       1 allocs/op

### AMD 8-core desktop

    $ go test -cpu 1,2,4,8,16 -benchmem  -run=^$   -bench  ^.*$
    goos: linux
    goarch: amd64
    pkg: github.com/solutionroute/rid/eval/bench
    cpu: AMD Ryzen 7 3800X 8-Core Processor             
    BenchmarkRid              	19820259	        59.45 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-2            	39708972	        29.60 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-4            	68249139	        15.26 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-8            	153133635	         7.773 ns/op	   0 B/op	       0 allocs/op
    BenchmarkRid-16           	277969574	         4.317 ns/op	   0 B/op	       0 allocs/op
    BenchmarkXid              	21991341	        52.61 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-2            	11893554	       101.4 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-4            	22855626	        52.89 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-8            	40688620	        33.60 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-16           	69289405	        16.92 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid            	 3220936	       367.3 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-2          	 1549900	       778.4 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-4          	 1397349	       856.2 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-8          	 1447863	       840.7 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-16         	 1406568	       862.3 ns/op	       0 B/op	       0 allocs/op
    BenchmarkGoogleUuid       	 3203955	       340.4 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-2     	 5828282	       203.0 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-4     	 9102511	       110.8 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-8     	20184568	        58.50 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-16    	29069440	        40.88 ns/op	      16 B/op	       1 allocs/op
    BenchmarkUlid             	  146658	      7878 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-2           	  272575	      4315 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-4           	  509660	      2363 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-8           	  720045	      1475 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-16          	  832911	      1504 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkBetterguid       	14368564	        80.30 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-2     	 7962180	       156.1 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-4     	 4272111	       277.1 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-8     	 3742044	       327.4 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-16    	 3163327	       384.9 ns/op	      24 B/op	       1 allocs/op