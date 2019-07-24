# Benchmark


```sh
$ go test -bench=.
goos: linux
goarch: amd64
pkg: github.com/moeryomenko/once
BenchmarkOnce-8                 2000000000               0.75 ns/op
BenchmarkCustomOnce-8           2000000000               0.23 ns/op
PASS
ok      github.com/moeryomenko/once     2.071s
```

on my other laptop(AMD Ryzen 3 2200U):

```sh
go test -bench=.
goos: linux
goarch: amd64
pkg: github.com/moeryomenko/once
BenchmarkOnce-4         	1000000000	         2.34 ns/op
BenchmarkCustomOnce-4   	2000000000	         0.79 ns/op
PASS
ok  	github.com/moeryomenko/once	4.248s
```
