# BenchmarksThe following benchmarks were performed on an Intel Core i7-8700 with Go version 1.13.4. Last update: 2019-11-14.

## Julien Schmidt's benchmark

### GitHub routes

```text
BenchmarkAce_GithubAll                21979        53705 ns/op       13792 B/op         167 allocs/op
BenchmarkAero_GithubAll               68666        17447 ns/op           0 B/op           0 allocs/op
BenchmarkBear_GithubAll               10000       162031 ns/op       86448 B/op         943 allocs/op
BenchmarkBeego_GithubAll               6987       201472 ns/op       71456 B/op         609 allocs/op
BenchmarkBone_GithubAll                 519      2208528 ns/op      720160 B/op        8620 allocs/op
BenchmarkChi_GithubAll                10000       177235 ns/op       87696 B/op         609 allocs/op
BenchmarkCloudyKitRouter_GithubAll    65502        18990 ns/op           0 B/op           0 allocs/op
BenchmarkDenco_GithubAll              24248        49496 ns/op       20224 B/op         167 allocs/op
BenchmarkEcho_GithubAll               41863        28730 ns/op           0 B/op           0 allocs/op
BenchmarkGin_GithubAll                40316        29733 ns/op           0 B/op           0 allocs/op
BenchmarkGocraftWeb_GithubAll          5114       231093 ns/op      131656 B/op        1686 allocs/op
BenchmarkGoji_GithubAll                4220       328501 ns/op       56112 B/op         334 allocs/op
BenchmarkGojiv2_GithubAll              1993       619642 ns/op      352720 B/op        4321 allocs/op
BenchmarkGoJsonRest_GithubAll          4424       281503 ns/op      134371 B/op        2737 allocs/op
BenchmarkGoRestful_GithubAll            514      2318252 ns/op      910144 B/op        2938 allocs/op
BenchmarkGorillaMux_GithubAll           416      2831774 ns/op      251650 B/op        1994 allocs/op
BenchmarkGowwwRouter_GithubAll        10000       108915 ns/op       72144 B/op         501 allocs/op
BenchmarkHttpRouter_GithubAll         32518        37113 ns/op       13792 B/op         167 allocs/op
BenchmarkHttpTreeMux_GithubAll        10000       109145 ns/op       65856 B/op         671 allocs/op
BenchmarkKocha_GithubAll              14562        81510 ns/op       23304 B/op         843 allocs/op
BenchmarkLARS_GithubAll               65749        18254 ns/op           0 B/op           0 allocs/op
BenchmarkMacaron_GithubAll             4489       270050 ns/op      149408 B/op        1624 allocs/op
BenchmarkMartini_GithubAll              458      2568245 ns/op      226551 B/op        2325 allocs/op
BenchmarkPat_GithubAll                  390      3117694 ns/op     1483152 B/op       26963 allocs/op
BenchmarkPossum_GithubAll             10000       140037 ns/op       84448 B/op         609 allocs/op
BenchmarkR2router_GithubAll           10000       123319 ns/op       77328 B/op         979 allocs/op
BenchmarkRivet_GithubAll              18517        62217 ns/op       16272 B/op         167 allocs/op
BenchmarkTango_GithubAll               7993       210040 ns/op       63825 B/op        1618 allocs/op
BenchmarkTigerTonic_GithubAll          2523       503390 ns/op      193856 B/op        4474 allocs/op
BenchmarkTraffic_GithubAll              484      2498177 ns/op      820744 B/op       14114 allocs/op
BenchmarkVulcan_GithubAll              8958       145914 ns/op       19894 B/op         609 allocs/op
```

### 1 Parameter

```text
BenchmarkAce_Param                  7891017          141 ns/op          32 B/op           1 allocs/op
BenchmarkAero_Param                29958960           39.9 ns/op         0 B/op           0 allocs/op
BenchmarkBear_Param                 1845138          656 ns/op         456 B/op           5 allocs/op
BenchmarkBeego_Param                1429641          830 ns/op         352 B/op           3 allocs/op
BenchmarkBone_Param                 1000000         1121 ns/op         816 B/op           6 allocs/op
BenchmarkChi_Param                  1842027          644 ns/op         432 B/op           3 allocs/op
BenchmarkCloudyKitRouter_Param     51542817           23.1 ns/op         0 B/op           0 allocs/op
BenchmarkDenco_Param               12743156           98.0 ns/op        32 B/op           1 allocs/op
BenchmarkEcho_Param                21220845           56.4 ns/op         0 B/op           0 allocs/op
BenchmarkGin_Param                 21031538           56.9 ns/op         0 B/op           0 allocs/op
BenchmarkGocraftWeb_Param           1253589          958 ns/op         648 B/op           8 allocs/op
BenchmarkGoji_Param                 2449986          478 ns/op         336 B/op           2 allocs/op
BenchmarkGojiv2_Param                827762         1736 ns/op        1328 B/op          11 allocs/op
BenchmarkGoJsonRest_Param           1000000         1028 ns/op         649 B/op          13 allocs/op
BenchmarkGoRestful_Param             333255         4193 ns/op        4192 B/op          14 allocs/op
BenchmarkGorillaMux_Param            649772         1894 ns/op        1280 B/op          10 allocs/op
BenchmarkGowwwRouter_Param          2190996          535 ns/op         432 B/op           3 allocs/op
BenchmarkHttpRouter_Param          15739543           75.8 ns/op        32 B/op           1 allocs/op
BenchmarkHttpTreeMux_Param          2773458          436 ns/op         352 B/op           3 allocs/op
BenchmarkKocha_Param                6384886          173 ns/op          56 B/op           3 allocs/op
BenchmarkLARS_Param                26426916           45.0 ns/op         0 B/op           0 allocs/op
BenchmarkMacaron_Param               804492         1757 ns/op        1072 B/op          10 allocs/op
BenchmarkMartini_Param               383647         3134 ns/op        1072 B/op          10 allocs/op
BenchmarkPat_Param                  1239183          961 ns/op         536 B/op          11 allocs/op
BenchmarkPossum_Param               1467655          807 ns/op         496 B/op           5 allocs/op
BenchmarkR2router_Param             2209725          543 ns/op         432 B/op           5 allocs/op
BenchmarkRivet_Param                9176148          131 ns/op          48 B/op           1 allocs/op
BenchmarkTango_Param                1714149          708 ns/op         248 B/op           8 allocs/op
BenchmarkTigerTonic_Param           1000000         1578 ns/op         776 B/op          16 allocs/op
BenchmarkTraffic_Param               474066         2714 ns/op        1856 B/op          21 allocs/op
BenchmarkVulcan_Param               3142342          371 ns/op          98 B/op           3 allocs/op
```

### 5 Parameters

```text
BenchmarkAce_Param5                 3655114          327 ns/op         160 B/op           1 allocs/op
BenchmarkAero_Param5               18535549           64.6 ns/op         0 B/op           0 allocs/op
BenchmarkBear_Param5                1416276          845 ns/op         501 B/op           5 allocs/op
BenchmarkBeego_Param5               1000000         1010 ns/op         352 B/op           3 allocs/op
BenchmarkBone_Param5                1000000         1432 ns/op         864 B/op           6 allocs/op
BenchmarkChi_Param5                 1417866          848 ns/op         432 B/op           3 allocs/op
BenchmarkCloudyKitRouter_Param5    13845614           87.1 ns/op         0 B/op           0 allocs/op
BenchmarkDenco_Param5               4161303          288 ns/op         160 B/op           1 allocs/op
BenchmarkEcho_Param5                8530136          138 ns/op           0 B/op           0 allocs/op
BenchmarkGin_Param5                10458777          117 ns/op           0 B/op           0 allocs/op
BenchmarkGocraftWeb_Param5          1000000         1560 ns/op         920 B/op          11 allocs/op
BenchmarkGoji_Param5                1984887          597 ns/op         336 B/op           2 allocs/op
BenchmarkGojiv2_Param5               588602         1955 ns/op        1392 B/op          11 allocs/op
BenchmarkGoJsonRest_Param5           676644         1887 ns/op        1097 B/op          16 allocs/op
BenchmarkGoRestful_Param5            251626         5151 ns/op        4288 B/op          14 allocs/op
BenchmarkGorillaMux_Param5           426506         2615 ns/op        1344 B/op          10 allocs/op
BenchmarkGowwwRouter_Param5         2056905          585 ns/op         432 B/op           3 allocs/op
BenchmarkHttpRouter_Param5          4833859          242 ns/op         160 B/op           1 allocs/op
BenchmarkHttpTreeMux_Param5         1283140          958 ns/op         576 B/op           6 allocs/op
BenchmarkKocha_Param5               1456714          823 ns/op         440 B/op          10 allocs/op
BenchmarkLARS_Param5               15111555           80.0 ns/op         0 B/op           0 allocs/op
BenchmarkMacaron_Param5              635964         1993 ns/op        1072 B/op          10 allocs/op
BenchmarkMartini_Param5              291230         3751 ns/op        1232 B/op          11 allocs/op
BenchmarkPat_Param5                  577328         2456 ns/op         888 B/op          29 allocs/op
BenchmarkPossum_Param5              1459933          809 ns/op         496 B/op           5 allocs/op
BenchmarkR2router_Param5            1897666          638 ns/op         432 B/op           5 allocs/op
BenchmarkRivet_Param5               2981439          398 ns/op         240 B/op           1 allocs/op
BenchmarkTango_Param5               1274169          924 ns/op         360 B/op           8 allocs/op
BenchmarkTigerTonic_Param5           227000         5649 ns/op        2279 B/op          39 allocs/op
BenchmarkTraffic_Param5              294430         4149 ns/op        2208 B/op          27 allocs/op
BenchmarkVulcan_Param5              2452029          486 ns/op          98 B/op           3 allocs/op
```

You can [run these by yourself](https://github.com/julienschmidt/go-http-routing-benchmark).

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
