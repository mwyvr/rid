[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/solutionroute/rid?tab=doc)[![Test](https://github.com/solutionroute/rid/actions/workflows/test.yaml/badge.svg)](https://github.com/solutionroute/rid/actions/workflows/test.yaml)[![Go Coverage](https://img.shields.io/badge/coverage-98.3%25-brightgreen.svg?style=flat)](http://gocover.io/github.com/solutionroute/rid)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# rid

Package rid provides a [k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence),
zero-configuration, unique ID generator.  Binary IDs are encoded as Base32,
producing a 20-character URL-friendly representation like: `ce0e7egs24nkzkn6egfg`.

The 12-byte binary representation of an ID is comprised of a:

- 4-byte timestamp value representing seconds since the Unix epoch
- 2-byte machine identifier
- 2-byte process identifier
- 4-byte cryptographically secure generated random value

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

Benchmarks were purposely left until last in this README as `rid` is plenty fast
enough for many use cases and scales well as CPU/process count increases. Even
my laptop can generate more than 4 million unique `rid.ID` per second on a
single process, scaling to 12.8 million on all cores. See 
[bench/bench_test.go](bench/bench_test.go).

### Intel 4-core Dell Latitude 7420 laptop

    $ go test -cpu 1,2,8 -benchmem  -run=^$   -benchtime 1s -bench  ^.*$ 
    goos: linux
    goarch: amd64
    pkg: github.com/solutionroute/rid/bench
    cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz
    BenchmarkRid               	 4045929	       292.5 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-2             	 6729403	       177.8 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-8             	12820012	        92.63 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid               	31313678	        37.44 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-2             	36996620	        32.17 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-8             	71023970	        16.59 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid             	 3845922	       319.1 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-2           	 2996962	       428.3 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-8           	 2637106	       448.8 ns/op	       0 B/op	       0 allocs/op
    BenchmarkGoogleUuid        	 3605164	       333.5 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-2   	 5092818	       226.6 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-8   	 8914668	       131.7 ns/op	      16 B/op	       1 allocs/op
    BenchmarkUlid           	  149518	      7541 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-2         	  237520	      4812 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-8         	  549673	      2098 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkBetterguid        	14327818	        80.14 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-2      	11664430	       101.1 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-8      	 6717274	       180.1 ns/op	      24 B/op	       1 allocs/op

### AMD 8-core Gigabyte AORUS Master desktop

    $ go test -cpu 1,2,8,16 -benchmem  -run=^$   -bench  ^.*$
    goos: linux
    goarch: amd64
    pkg: foo
    cpu: AMD Ryzen 7 3800X 8-Core Processor             
    BenchmarkRid              	 3195141	       359.2 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-2            	 4658292	       247.4 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-8            	12757447	        94.02 ns/op	       0 B/op	       0 allocs/op
    BenchmarkRid-16           	19939770	        60.63 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid              	22658910	        51.73 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-2            	37240394	        97.57 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-8            	67146513	        32.98 ns/op	       0 B/op	       0 allocs/op
    BenchmarkXid-16           	69354812	        16.97 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid            	 3234849	       365.0 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-2          	 1927971	       766.7 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-8          	 1428528	       828.3 ns/op	       0 B/op	       0 allocs/op
    BenchmarkKsuid-16         	 1417254	       843.6 ns/op	       0 B/op	       0 allocs/op
    BenchmarkGoogleUuid       	 3594734	       324.3 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-2     	 5600349	       210.4 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-8     	20083578	        58.88 ns/op	      16 B/op	       1 allocs/op
    BenchmarkGoogleUuid-16    	29556532	        40.56 ns/op	      16 B/op	       1 allocs/op
    BenchmarkUlid             	  149061	      7771 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-2           	  244858	      4272 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-8           	  793448	      1461 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkUlid-16          	  832110	      1513 ns/op	    5440 B/op	       3 allocs/op
    BenchmarkBetterguid       	14165331	        81.85 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-2     	10588766	       137.5 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-8     	 4297107	       272.1 ns/op	      24 B/op	       1 allocs/op
    BenchmarkBetterguid-16    	 3449325	       347.0 ns/op	      24 B/op	       1 allocs/op

