# Benchmark

This implementation based on spinlock from `https://github.com/moeryomenko/sync`

```sh
$ go version
go version go1.16.3 linux/amd64
$ go test -race=1 -bench=.
goos: linux
goarch: amd64
pkg: github.com/moeryomenko/once
cpu: AMD Ryzen 5 3500U with Radeon Vega Mobile Gfx
BenchmarkOnce-8         	15947769	        77.77 ns/op
BenchmarkCustomOnce-8   	17909221	        66.80 ns/op
PASS
ok  	github.com/moeryomenko/once	2.621s
```
