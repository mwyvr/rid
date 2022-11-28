# rid Performance

**Acknowledgement**: This package borrows heavily from the
[rs/xid](https://github.com/rs/xid) package which itself levers ideas from
[MongoDB](https://docs.mongodb.com/manual/reference/method/ObjectId/). Where
this package differs is the use of a random number as opposed to xid's trailing
counter for the last 4 bytes of the ID.

To improve performance and scaling on multiple cores, random numbers are
obtained from a runtime function via a slightly hacky use of maphash.Hash.

Unexplored, random number generation performance on my 8-core AMD Ryzen 7 3800
doesn't keep up with my Dell Latitude 7420 laptop with an Intel mobile Core i7.
In time will track down reports of AMD related issues.

Regardless of architecture, generation of a million unique rids takes less than
a second.


|AMD Ryzen 7 3800X 8-Core          | 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz|
|----------------------------------|-----------------------------------------------|
|$ time rid -c 1000000 > /dev/null | $ time rid -c 1000000 > /dev/null|
|real	0m0.710s                     | real	0m0.413s|
|user	0m0.552s                     | user	0m0.324s|
|sys	0m0.167s                     | sys	0m0.094s|

## Benchmark

`rid` using random number generation is inherently slower than an incrementing
counter such as used in `xid`. That said even my laptop can generate 1 million
unique ids in less than a second, and performance does not degrade
significantly as core count increases.

All benchmarks were run on [Void Linux](https://voidlinux.org/) (Linux 6.09 kernel) and Go 1.19.2.

AMD based Desktop with 8 cores/16 cpus:

    goos: linux
    goarch: amd64
    pkg: github.com/solutionroute/rid
    cpu: AMD Ryzen 7 3800X 8-Core Processor             

    $ go test -cpu 1 -benchmem  -run=^$   -bench  ^.*$
    BenchmarkNew        	 8089182	       178.6 ns/op	      13 B/op	       0 allocs/op
    BenchmarkNewString  	 6620941	       185.9 ns/op	       0 B/op	       0 allocs/op
    BenchmarkString     	126099992	         9.539 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFromString 	55525843	        21.19 ns/op	       0 B/op	       0 allocs/op

    $ go test -cpu 8 -benchmem  -run=^$   -bench  ^.*$
    BenchmarkNew-8          	 2209707	       492.0 ns/op	      16 B/op	       0 allocs/op
    BenchmarkNewString-8    	 2472127	       494.7 ns/op	       0 B/op	       0 allocs/op
    BenchmarkString-8       	826893270	         1.242 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFromString-8   	427606477	         2.764 ns/op	       0 B/op	       0 allocs/op

    $ go test -cpu 16 -benchmem  -run=^$   -bench  ^.*$
    BenchmarkNew-16           	 2099038	       556.1 ns/op	      25 B/op	       0 allocs/op
    BenchmarkNewString-16     	 1987956	       598.5 ns/op	       0 B/op	       0 allocs/op
    BenchmarkString-16        	961245805	         1.154 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFromString-16    	463253820	         2.541 ns/op	       0 B/op	       0 allocs/op

Intel based Laptop with 4 cores/8 cpus:

    goos: linux
    goarch: amd64
    pkg: github.com/solutionroute/rid
    cpu: 11th Gen Intel(R) Core(TM) i7-1185G7 @ 3.00GHz

    $ go test -cpu 1 -benchmem  -run=^$   -bench  ^.*$ 
    BenchmarkNew        	 8739328	       174.3 ns/op	      13 B/op	       0 allocs/op
    BenchmarkNewString  	 6491241	       180.3 ns/op	       1 B/op	       0 allocs/op
    BenchmarkString     	137119748	         8.829 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFromString 	48120292	        22.87 ns/op	       0 B/op	       0 allocs/op

    $ go test -cpu 8 -benchmem  -run=^$   -bench  ^.*$ 
    BenchmarkNew-8          	 5880898	       208.7 ns/op	      18 B/op	       0 allocs/op
    BenchmarkNewString-8    	 5576970	       215.7 ns/op	       0 B/op	       0 allocs/op
    BenchmarkString-8       	373706143	         3.151 ns/op	       0 B/op	       0 allocs/op
    BenchmarkFromString-8   	181656850	         6.585 ns/op	       0 B/op	       0 allocs/op

