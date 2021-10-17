# Synx

Synx is enhanced version of standard `sync` package.

## Spinlock

TODO

## Once

```sh
$ go version
go version go1.17.1 linux/amd64
$ go test -race -bench=. ./...
goos: linux
goarch: amd64
pkg: github.com/moeryomenko/synx
cpu: AMD Ryzen 5 3500U with Radeon Vega Mobile Gfx
BenchmarkOnce-8         	19812547	       61.74 ns/op
BenchmarkCustomOnce-8   	23804229	       50.94 ns/op
PASS
ok  	github.com/moeryomenko/synx	2.571s
```
