# Benchmarks

The following benchmarks were performed on an Intel Core i7-8700 with Go version 1.13.4. Last update: 2019-11-14.

## Comparison with Echo and Gin

```text
BenchmarkAeroStatic-12            73549            15788 ns/op          943 B/op         0 allocs/op
BenchmarkAeroGitHubAPI-12         49735            24383 ns/op         1396 B/op         0 allocs/op
BenchmarkAeroGplusAPI-12        1004732             1226 ns/op           68 B/op         0 allocs/op
BenchmarkAeroParseAPI-12         508233             2268 ns/op          136 B/op         0 allocs/op

BenchmarkEchoStatic-12            39878            30057 ns/op         2127 B/op       157 allocs/op
BenchmarkEchoGitHubAPI-12         26320            45215 ns/op         2945 B/op       203 allocs/op
BenchmarkEchoGplusAPI-12         479107             2410 ns/op          176 B/op        13 allocs/op
BenchmarkEchoParseAPI-12         273926             4247 ns/op          334 B/op        26 allocs/op

BenchmarkGinStatic-12             29456            39892 ns/op         8715 B/op       157 allocs/op
BenchmarkGinGitHubAPI-12          20192            58868 ns/op        10609 B/op       203 allocs/op
BenchmarkGinGplusAPI-12          354789             3338 ns/op          721 B/op        13 allocs/op
BenchmarkGinParseAPI-12          199041             6017 ns/op         1422 B/op        26 allocs/op
```

You can [run these by yourself](https://github.com/akyoto/web-framework-benchmark).

## Julien Schmidt's benchmark

```text
BenchmarkAero_Param        	25261422	        46.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_Param5       	16850889	        71.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_ParamWrite   	15703929	        75.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_GithubStatic 	23779368	        50.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_GithubParam  	13484388	        89.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_GithubAll    	   61533	     19341 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_GPlusStatic  	28713459	        41.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_GPlusParam   	20435862	        58.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_GPlus2Params 	13035148	        90.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_GPlusAll     	 1233676	       977 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_ParseStatic  	26852730	        44.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_ParseParam   	21789049	        55.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_Parse2Params 	17876079	        66.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_ParseAll     	  732868	      1605 ns/op	       0 B/op	       0 allocs/op
BenchmarkAero_StaticAll    	   90670	     13162 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/julienschmidt/go-http-routing-benchmark	21.883s
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
