![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/mwyvr/rid)[![godoc](http://img.shields.io/badge/godev-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/mwyvr/rid?tab=doc)[![Test](https://github.com/mwyvr/rid/actions/workflows/test.yaml/badge.svg)](https://github.com/mwyvr/rid/actions/workflows/test.yaml)[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)![Coverage](https://img.shields.io/badge/coverage-92.6%25-brightgreen)

# rid

Package rid provides a performant, goroutine-safe generator of short
[k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence) unique IDs
suitable for use where inter-process ID generation coordination is not
required.

Using a non-standard character set (fewer vowels), IDs Base-32 encode as a
16-character URL-friendly, case-insensitive representation like
`dfp7qt0v2pwt0v2x`.

An ID is a:

  - 4-byte timestamp value representing seconds since the Unix epoch, plus a
  - 6-byte random value; see the [Random Source](#random-source) discussion.

Built-in (de)serialization simplifies interacting with SQL databases and JSON.
`cmd/rid` provides the `rid` utility to generate or inspect IDs. Thanks to
`internal/fastrand` introduced in Go 1.19 and made the default `math/rand` source in Go
1.20, ID generation starts fast and scales well as cores are added. De-serialization
has also been optimized. See [Package Benchmarks](#package-benchmarks).

Why `rid` instead of [alternatives](#package-comparisons)?

  - At 10 bytes binary, 16 bytes Base32 encoded, rid.IDs are case-insensitive
    and short, yet with 48 bits of uniqueness *per second*, are unique
    enough for many use cases.
  - IDs have a random component rather than a potentially guessable
    monotonic counter found in some libraries.

_**Acknowledgement**: This package borrows heavily from rs/xid
(https://github.com/rs/xid), a zero-configuration globally-unique
high-performance ID generator that leverages ideas from MongoDB
(https://docs.mongodb.com/manual/reference/method/ObjectId/)._

## Example:

```go
id := rid.New()
fmt.Printf("%s\n", id.String())
// Output: dfp7qt97menfv8ll

id2, err := rid.FromString("dfp7qt97menfv8ll")
if err != nil {
	fmt.Println(err)
}
fmt.Printf("%s %d %v\n", id2.Time(), id2.Random(), id2.Bytes())
// Output: 2022-12-28 09:24:57 -0800 PST 43582827111027 [99 172 123 233 39 163 106 237 162 115]
```

## CLI

Package `rid` also provides the `rid` tool for id generation and inspection. 

    $ rid 
	dfpb18y8dg90hc74

 	$ rid -c 2
	dfp9l9cgs05blztq
	dfp9l9d80yxdf804

    # produce 4 and inspect
	$ rid `rid -c 4`
	dfp9lmz9ksw87w48 ts:1672255955 rnd:256798116540552 2022-12-28 11:32:35 -0800 PST ID{ 0x63, 0xac, 0x99, 0xd3, 0xe9, 0x8e, 0x78, 0x83, 0xf0, 0x88 }
	dfp9lmxefym2ht2f ts:1672255955 rnd:190729433933902 2022-12-28 11:32:35 -0800 PST ID{ 0x63, 0xac, 0x99, 0xd3, 0xad, 0x77, 0xa8, 0x28, 0x68, 0x4e }
	dfp9lmt5zjy7km9n ts:1672255955 rnd: 76951796109621 2022-12-28 11:32:35 -0800 PST ID{ 0x63, 0xac, 0x99, 0xd3, 0x45, 0xfc, 0xbc, 0x78, 0xd1, 0x35 }
	dfp9lmxt5sms80m7 ts:1672255955 rnd:204708502569607 2022-12-28 11:32:35 -0800 PST ID{ 0x63, 0xac, 0x99, 0xd3, 0xba, 0x2e, 0x69, 0x94,  0x2, 0x87 }

## Uniqueness
 
To satisfy whether rid.IDs are unique enough for your use case, run
[eval/uniqcheck/main.go](eval/uniqcheck/main.go) with various values for number
of go routines and iterations, or, at the command line, produce IDs
and use OS utilities to check:

    rid -c 2000000 | sort | uniq -d
    // None output

## Change Log

- 2025-03-03 head: Now utilizing crypto/rand; performance remains acceptable. Require Go 1.24+.
- 2025-02-28 head: Updated benchmarks, included google/uuid V7 as well as more output for visual comparison.
- 2023-03-02 v1.1.6: Package depends on math/rand/v2 and now requires Go 1.22+.
- 2023-01-23 Replaced the stdlib Base32 encoding/decoding with an unrolled version for decoding performance.
- 2022-12-28 The "10byte" branch was merged to master; the "15byte-historical" branch will be left dormant.

## Contributing

Contributions are welcome.

## Package Comparisons

| Package                                                   |BLen|ELen| K-Sort| Encoded ID and Next | Method | Components |
|-----------------------------------------------------------|----|----|-------|---------------------|--------|------------|
| [solutionroute/rid](https://github.com/solutionroute/rid) | 10 | 16 |  true | `dz369n6d6377s87d`<br>`dz369n0xkc5yy0qk`<br>`dz369n1f66fp20wq`<br>`dz369n7c9wtgp944`  | crypt/rand | 4 byte ts(sec) : 6 byte random |
| [rs/xid](https://github.com/rs/xid)                       | 12 | 20 |  true | `cv369l5q9fa4kd8oi9a0`<br>`cv369l5q9fa4kd8oi9ag`<br>`cv369l5q9fa4kd8oi9b0`<br>`cv369l5q9fa4kd8oi9bg`  | counter | 4 byte ts(sec) : 2 byte mach ID : 2 byte pid : 3 byte monotonic counter |
| [segmentio/ksuid](https://github.com/segmentio/ksuid)     | 20 | 27 |  true | `2tphzpKfP9Tj0r14XtsXw48vvGU`<br>`2tphzrH8yIhsWyMIRpSMKjLI36k`<br>`2tphzvMCgAaNd7mpYtkAUzwahdx`<br>`2tphztFxf0sFFm6fXU10v5X9J3v`  | math/rand | 4 byte ts(sec) : 16 byte random |
| [google/uuid](https://github.com/google/uuid)             | 16 | 36 | false | `cb978d85-a710-488b-bebd-8dbaf55db670`<br>`98128f69-2190-4e73-8631-25cc56e9eb00`<br>`1a662558-e3bb-4bfb-a775-a39da8177f01`<br>`f9e320c4-86b9-4055-b387-160a8456a049`  | crypt/rand | v4: 16 bytes random with version & variant embedded |
| [google/uuid](https://github.com/google/uuid)V7           | 16 | 36 |  true | `01955ef9-df51-7d46-9d2b-f01195e3fa80`<br>`01955ef9-df51-7d47-9282-15a86d461159`<br>`01955ef9-df51-7d48-bfc8-147557df171a`<br>`01955ef9-df51-7d49-8cc2-5637be115e7b`  | crypt/rand | v7: 16 bytes : 8 bytes time+sequence, random with version & variant embedded |
| [oklog/ulid](https://github.com/oklog/ulid)               | 16 | 26 |  true | `01JNFFKQTH2VD414Z3C1HRX676`<br>`01JNFFKQTHH1SBBT2NG0XC0XJH`<br>`01JNFFKQTHD1MYD3RXVTVGS8P3`<br>`01JNFFKQTHRK8M66FP3DGY31C3`  | crypt/rand | 6 byte ts(ms) : 10 byte counter random init per ts(ms) |
| [kjk/betterguid](https://github.com/kjk/betterguid)       | 17 | 20 |  true | `-OKTySxG2zFlWEvK7zUy`<br>`-OKTySxG2zFlWEvK7zUz`<br>`-OKTySxG2zFlWEvK7zV-`<br>`-OKTySxG2zFlWEvK7zV0`  | counter | 8 byte ts(ms) : 9 byte counter random init per ts(ms) |

| Package                                                   |BLen|ELen| K-Sort| Encoded ID and Next | Method | Components |
|-----------------------------------------------------------|----|----|-------|---------------------|--------|------------|
| [solutionroute/rid](https://github.com/solutionroute/rid) | 10 | 16 |  true | `dz13fh4t76bfkq9r`<br>`dz13fh5t4rsy3tb7`<br>`dz13fh445w8s044g`<br>`dz13fh45dwl22j6z`  | math/rand/v2 | 4 byte ts(sec) : 6 byte random |
| [rs/xid](https://github.com/rs/xid)                       | 12 | 20 |  true | `cv13eg5q9fafhigle550`<br>`cv13eg5q9fafhigle55g`<br>`cv13eg5q9fafhigle560`<br>`cv13eg5q9fafhigle56g`  | counter | 4 byte ts(sec) : 2 byte mach ID : 2 byte pid : 3 byte monotonic counter |
| [segmentio/ksuid](https://github.com/segmentio/ksuid)     | 20 | 27 |  true | `2tgl0f9wgEOOphdVlVRlU21RUJ2`<br>`2tgl0d9ZVFNDRMvlvJ1UNW5MkO5`<br>`2tgl0gGaE7OJHvRoqNvUXVmztxG`<br>`2tgl0d310frXcqu7ppDvIfkna6f`  | math/rand | 4 byte ts(sec) : 16 byte random |
| [google/uuid](https://github.com/google/uuid)             | 16 | 36 | false | `340a0264-d35a-4c3c-b68c-755778c36050`<br>`16efa807-d9c0-4613-9c4e-d192e0404bc4`<br>`379c0c45-b1f0-4a21-ae3e-43bc590a6c96`<br>`df20e8b8-f4e7-45f1-91f0-b9d5d4874ff7`  | crypt/rand | v4: 16 bytes random with version & variant embedded |
| [google/uuid](https://github.com/google/uuid)V7           | 16 | 36 |  true | `01954ea7-d25c-71fb-97d7-81556c5b5728`<br>`01954ea7-d25c-71fc-970b-c617b4a260b4`<br>`01954ea7-d25c-71fd-9a80-a4951289adfe`<br>`01954ea7-d25c-71fe-8dd2-933a4a2a5574`  | crypt/rand | v7: 16 bytes : 6 bytes time, random with version & variant embedded |b
| [oklog/ulid](https://github.com/oklog/ulid)               | 16 | 26 |  true | `01JN7AFMJWJN7B3W6FCQW3BNDP`<br>`01JN7AFMJWDDAF6K2FDGJR7ZX2`<br>`01JN7AFMJW2NS2YCGD16EZ4GHY`<br>`01JN7AFMJWA04FB0EE3Q6GKNP3`  | crypt/rand | 6 byte ts(ms) : 10 byte counter random init per ts(ms) |
| [kjk/betterguid](https://github.com/kjk/betterguid)       | 17 | 20 |  true | `-OKDdx8RLoXfnXo8EfYU`<br>`-OKDdx8RLoXfnXo8EfYV`<br>`-OKDdx8RLoXfnXo8EfYW`<br>`-OKDdx8RLoXfnXo8EfYX`  | counter | 8 byte ts(ms) : 9 byte counter random init per ts(ms) |


With only 48 bits of randomness per second, `rid` does not attempt to weigh in
as a globally unique ID generator. If that is your requirement, `rs/xid` is a
solid  feature comparable alternative for such needs.

For a comparison of various Go-based unique ID solutions, see:
https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

## Package Benchmarks

A benchmark suite for the above-noted packages can be found in
[eval/bench/bench_test.go](eval/bench/bench_test.go). All runs were done with
scaling_governor set to `performance`:

    echo "performance" | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

```
❯ go test -cpu 1,2,4,8,16,32 -test.benchmem -bench .
goos: linux
goarch: amd64
pkg: github.com/mwyvr/rid/eval/bench
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkRid                	25817522	       42.16 ns/op	      0 B/op	      0 allocs/op
BenchmarkRid-2              	32066263	       32.23 ns/op	      0 B/op	      0 allocs/op
BenchmarkRid-4              	36123939	       31.55 ns/op	      0 B/op	      0 allocs/op
BenchmarkRid-8              	36245113	       32.63 ns/op	      0 B/op	      0 allocs/op
BenchmarkRid-16             	36280018	       32.84 ns/op	      0 B/op	      0 allocs/op
BenchmarkRid-32             	55567880	       22.81 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid                	39963730	       28.66 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid-2              	38590256	       29.45 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid-4              	37522834	       31.74 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid-8              	39866997	       32.51 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid-16             	43631900	       32.68 ns/op	      0 B/op	      0 allocs/op
BenchmarkXid-32             	49602223	       22.44 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid              	16018834	       74.81 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid-2            	14673158	       81.78 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid-4            	11699431	      104.5 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid-8            	10058565	      117.8 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid-16           	8374312	      146.6 ns/op	      0 B/op	      0 allocs/op
BenchmarkKsuid-32           	7095456	      171.0 ns/op	      0 B/op	      0 allocs/op
BenchmarkGoogleUuid         	23303894	       48.78 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuid-2       	29063116	       39.37 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuid-4       	34529593	       33.95 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuid-8       	36596574	       33.07 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuid-16      	46557832	       33.95 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuid-32      	46031656	       25.79 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7       	14000478	       84.06 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7-2     	13783161	       88.46 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7-4     	10600320	      113.3 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7-8     	9574357	      127.0 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7-16    	8125311	      151.0 ns/op	     16 B/op	      1 allocs/op
BenchmarkGoogleUuidV7-32    	7138972	      173.9 ns/op	     16 B/op	      1 allocs/op
BenchmarkUlid               	 200292	     5709 ns/op	   5440 B/op	      3 allocs/op
BenchmarkUlid-2             	 381554	     3104 ns/op	   5440 B/op	      3 allocs/op
BenchmarkUlid-4             	 700375	     1752 ns/op	   5440 B/op	      3 allocs/op
BenchmarkUlid-8             	1000000	     1084 ns/op	   5440 B/op	      3 allocs/op
BenchmarkUlid-16            	 974892	     1223 ns/op	   5440 B/op	      3 allocs/op
BenchmarkUlid-32            	 900584	     1313 ns/op	   5440 B/op	      3 allocs/op
BenchmarkBetterguid         	25144016	       46.35 ns/op	     24 B/op	      1 allocs/op
BenchmarkBetterguid-2       	24168476	       49.03 ns/op	     24 B/op	      1 allocs/op
BenchmarkBetterguid-4       	18565868	       65.24 ns/op	     24 B/op	      1 allocs/op
BenchmarkBetterguid-8       	14666895	       79.84 ns/op	     24 B/op	      1 allocs/op
BenchmarkBetterguid-16      	11732860	      109.2 ns/op	     24 B/op	      1 allocs/op
BenchmarkBetterguid-32      	9067160	      132.9 ns/op	     24 B/op	      1 allocs/op
PASS
ok  	github.com/mwyvr/rid/eval/bench	53.126s
```
