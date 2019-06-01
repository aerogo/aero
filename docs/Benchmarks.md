# Benchmarks

## Routing

```text
BenchmarkAeroStatic-12           100000            16643 ns/op          700 B/op         0 allocs/op
BenchmarkAeroGitHubAPI-12         50000            27487 ns/op         1409 B/op         0 allocs/op
BenchmarkAeroGplusAPI-12        1000000             1390 ns/op           69 B/op         0 allocs/op
BenchmarkAeroParseAPI-12         500000             2558 ns/op          138 B/op         0 allocs/op

BenchmarkEchoStatic-12            50000            30702 ns/op         1950 B/op       157 allocs/op
BenchmarkEchoGitHubAPI-12         30000            45431 ns/op         2782 B/op       203 allocs/op
BenchmarkEchoGplusAPI-12         500000             2500 ns/op          173 B/op        13 allocs/op
BenchmarkEchoParseAPI-12         300000             4234 ns/op          323 B/op        26 allocs/op

BenchmarkGinStatic-12             50000            37885 ns/op         8231 B/op       157 allocs/op
BenchmarkGinGitHubAPI-12          30000            55092 ns/op        10903 B/op       203 allocs/op
BenchmarkGinGplusAPI-12          500000             3059 ns/op          693 B/op        13 allocs/op
BenchmarkGinParseAPI-12          300000             5687 ns/op         1363 B/op        26 allocs/op
```

You can [run these by yourself](https://github.com/akyoto/web-framework-benchmark).

## Latency

Simple latency test on an Intel Core i7-8700 using `Hello World` as output:

```text
λ wrk -t12 -c400 -d2s http://localhost:4000/
Running 2s test @ http://localhost:4000/
  12 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.39ms    1.57ms  22.91ms   87.67%
    Req/Sec    30.91k     5.23k   55.90k    76.11%
  761999 requests in 2.08s, 93.02MB read
Requests/sec: 365857.04
Transfer/sec:     44.66MB
```

## Explanation

Aero's router uses highly optimized [radix trees](https://en.wikipedia.org/wiki/Radix_tree) with 0 allocations during route lookup and is therefore extremely fast. I am fairly confident that it is currently the fastest routing implementation out there.

In addition to being fast, it also supports smart route prioritization: **Static > Parameter > Wildcard**.
