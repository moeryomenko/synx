# once
custom implemention of sync.Once

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
