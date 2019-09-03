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
$ go test -bench=.
goos: linux
goarch: amd64
pkg: github.com/moeryomenko/once
BenchmarkOnce-4         	1000000000	         2.34 ns/op
BenchmarkCustomOnce-4   	2000000000	         0.79 ns/op
PASS
ok  	github.com/moeryomenko/once	4.248s
```

so... Go 1.13 really fast(:

AMD Ryzen 3 2200U:
```sh
~/.local/go/bin/go test -bench=.
goos: linux
goarch: amd64
pkg: github.com/moeryomenko/once
BenchmarkOnce-4                 1000000000               0.854 ns/op
BenchmarkCustomOnce-4           1000000000               0.793 ns/op
PASS
ok      github.com/moeryomenko/once     1.817s
```

Intel Core i7-8565U:
```sh
$ ~/.local/go/bin/go version      
go version go1.13 linux/amd64
$ ~/.local/go/bin/go test -bench=.
goos: linux
goarch: amd64
pkg: github.com/moeryomenko/once
BenchmarkOnce-8                 1000000000               0.229 ns/op
BenchmarkCustomOnce-8           1000000000               0.224 ns/op
PASS
ok      github.com/moeryomenko/once     0.504s
```
