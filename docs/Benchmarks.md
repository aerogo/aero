# Benchmarks

The following benchmarks were performed on an Intel Core i7-8700 with Go version 1.17.3. Last update: 2021-11-19.

## Julien Schmidt's benchmark

You can [run these on your own](https://github.com/julienschmidt/go-http-routing-benchmark).
Make sure you run the latest version via `go get -u all`.
Invoke a benchmark via `go test -bench=_GithubAll`.

### GitHub routes

2021-11-19 (Go 1.17.3):

```text
BenchmarkAce_GithubAll                30292        39445 ns/op       13792 B/op         167 allocs/op
BenchmarkAero_GithubAll               81526        14432 ns/op           0 B/op           0 allocs/op
BenchmarkBear_GithubAll               10000       133282 ns/op       86448 B/op         943 allocs/op
BenchmarkBeego_GithubAll               9314       158744 ns/op       71456 B/op         609 allocs/op
BenchmarkBone_GithubAll                 664      1780877 ns/op      722832 B/op        8620 allocs/op
BenchmarkChi_GithubAll                10000       137437 ns/op       90944 B/op         609 allocs/op
BenchmarkCloudyKitRouter_GithubAll    71870        16239 ns/op           0 B/op           0 allocs/op
BenchmarkDenco_GithubAll              29060        40105 ns/op       20224 B/op         167 allocs/op
BenchmarkEcho_GithubAll               47343        24813 ns/op           0 B/op           0 allocs/op
BenchmarkGin_GithubAll                43820        26952 ns/op           0 B/op           0 allocs/op
BenchmarkGocraftWeb_GithubAll          6020       186604 ns/op      130032 B/op        1686 allocs/op
BenchmarkGoji_GithubAll                4681       272356 ns/op       56112 B/op         334 allocs/op
BenchmarkGojiv2_GithubAll              2228       529902 ns/op      362464 B/op        4321 allocs/op
BenchmarkGoJsonRest_GithubAll          4813       225523 ns/op      127875 B/op        2737 allocs/op
BenchmarkGoRestful_GithubAll            788      1451192 ns/op      905272 B/op        2938 allocs/op
BenchmarkGorillaMux_GithubAll           487      2341250 ns/op      258146 B/op        1994 allocs/op
BenchmarkGowwwRouter_GithubAll        13741        86902 ns/op       74816 B/op         501 allocs/op
BenchmarkHttpRouter_GithubAll         37346        31390 ns/op       13792 B/op         167 allocs/op
BenchmarkHttpTreeMux_GithubAll        12152        98425 ns/op       65856 B/op         671 allocs/op
BenchmarkKocha_GithubAll              16977        68615 ns/op       23304 B/op         843 allocs/op
BenchmarkLARS_GithubAll               67502        17191 ns/op           0 B/op           0 allocs/op
BenchmarkMacaron_GithubAll             5026       225779 ns/op      147784 B/op        1624 allocs/op
BenchmarkMartini_GithubAll              524      2194311 ns/op      231418 B/op        2731 allocs/op
BenchmarkPat_GithubAll                  482      2428417 ns/op     1410624 B/op       22515 allocs/op
BenchmarkPossum_GithubAll             12115        98480 ns/op       84448 B/op         609 allocs/op
BenchmarkR2router_GithubAll           12121        98280 ns/op       75704 B/op         979 allocs/op
BenchmarkRivet_GithubAll              23821        48665 ns/op       16272 B/op         167 allocs/op
BenchmarkTango_GithubAll               8010       168809 ns/op       58721 B/op        1418 allocs/op
BenchmarkTigerTonic_GithubAll          2919       416451 ns/op      190848 B/op        4474 allocs/op
BenchmarkTraffic_GithubAll              560      2104084 ns/op      819047 B/op       14114 allocs/op
BenchmarkVulcan_GithubAll             10000       106652 ns/op       19894 B/op         609 allocs/op
```

2019-11-15 (Go 1.13.4)

```text
BenchmarkAce_GithubAll                22688        51771 ns/op       13792 B/op         167 allocs/op
BenchmarkAero_GithubAll               74337        15961 ns/op           0 B/op           0 allocs/op
BenchmarkBear_GithubAll               10000       156590 ns/op       86448 B/op         943 allocs/op
BenchmarkBeego_GithubAll               6836       195807 ns/op       71456 B/op         609 allocs/op
BenchmarkBone_GithubAll                 558      2114383 ns/op      720160 B/op        8620 allocs/op
BenchmarkChi_GithubAll                10000       175337 ns/op       87696 B/op         609 allocs/op
BenchmarkCloudyKitRouter_GithubAll    63306        18881 ns/op           0 B/op           0 allocs/op
BenchmarkDenco_GithubAll              24750        48046 ns/op       20224 B/op         167 allocs/op
BenchmarkEcho_GithubAll               40288        29617 ns/op           0 B/op           0 allocs/op
BenchmarkGin_GithubAll                37447        30780 ns/op           0 B/op           0 allocs/op
BenchmarkGocraftWeb_GithubAll          5401       224097 ns/op      131656 B/op        1686 allocs/op
BenchmarkGoji_GithubAll                4064       312670 ns/op       56112 B/op         334 allocs/op
BenchmarkGojiv2_GithubAll              2190       608078 ns/op      352720 B/op        4321 allocs/op
BenchmarkGoJsonRest_GithubAll          4096       273647 ns/op      134371 B/op        2737 allocs/op
BenchmarkGoRestful_GithubAll            517      2337735 ns/op      910144 B/op        2938 allocs/op
BenchmarkGorillaMux_GithubAll           414      2822359 ns/op      251650 B/op        1994 allocs/op
BenchmarkGowwwRouter_GithubAll        10000       105378 ns/op       72144 B/op         501 allocs/op
BenchmarkHttpRouter_GithubAll         34100        35705 ns/op       13792 B/op         167 allocs/op
BenchmarkHttpTreeMux_GithubAll        10000       108362 ns/op       65856 B/op         671 allocs/op
BenchmarkKocha_GithubAll              15163        78528 ns/op       23304 B/op         843 allocs/op
BenchmarkLARS_GithubAll               59350        19525 ns/op           0 B/op           0 allocs/op
BenchmarkMacaron_GithubAll             4477       268154 ns/op      149408 B/op        1624 allocs/op
BenchmarkMartini_GithubAll              478      2523755 ns/op      226551 B/op        2325 allocs/op
BenchmarkPat_GithubAll                  392      3035309 ns/op     1483152 B/op       26963 allocs/op
BenchmarkPossum_GithubAll             10000       137991 ns/op       84448 B/op         609 allocs/op
BenchmarkR2router_GithubAll           10000       119743 ns/op       77328 B/op         979 allocs/op
BenchmarkRivet_GithubAll              19720        60986 ns/op       16272 B/op         167 allocs/op
BenchmarkTango_GithubAll               7989       209209 ns/op       63825 B/op        1618 allocs/op
BenchmarkTigerTonic_GithubAll          2620       496523 ns/op      193856 B/op        4474 allocs/op
BenchmarkTraffic_GithubAll              486      2447732 ns/op      820744 B/op       14114 allocs/op
BenchmarkVulcan_GithubAll              9208       142581 ns/op       19894 B/op         609 allocs/op
```

## Latency

Simple latency test using `Hello World` as output:

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

Aero's router uses highly optimized [radix trees](https://en.wikipedia.org/wiki/Radix_tree) with 0 allocations during route lookup and is therefore extremely fast.

In addition to being fast, it also supports smart route prioritization: **Static > Parameter > Wildcard**.
