# Aero
This project is work in progress.

## Goals

- [x] Reach 100/100 [Google PageSpeed](https://developers.google.com/speed/pagespeed/insights/) score
- [x] Reach 120/100 [Mozilla Observatory](https://observatory.mozilla.org/) score (A+)
- [ ] Reach 100/100 [Google Lighthouse](https://developers.google.com/web/tools/lighthouse/) score

## Documentation
* [Application](docs/Application.md)
* [API](docs/API.md)

## Benchmark
Simple routing test using `Hello World` as output (data payload too small for gzip):

*TODO: Post benchmark with ~50 KB of data (performance difference is crazy).*

AeroJS (node.js 8.0):
```
Running 2s test @ http://localhost:4000/hello
  8 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    10.62ms    2.38ms  35.95ms   86.46%
    Req/Sec     4.11k   777.46     9.25k    90.85%
  67040 requests in 2.10s, 7.10MB read
Requests/sec:  31933.33
Transfer/sec:      3.38MB
```

AeroGo (go 1.8.3):
```
Running 2s test @ http://dev.local.host/hello
  8 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.36ms    1.73ms  22.67ms   75.86%
    Req/Sec    22.12k     3.34k   35.27k    73.12%
  353303 requests in 2.05s, 66.38MB read
Requests/sec: 172496.91
Transfer/sec:     32.41MB
```

AeroGo using an SSL certificate (go 1.8.3)
```
Running 2s test @ https://dev.local.host/hello
  8 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     3.14ms    2.22ms  22.43ms   73.05%
    Req/Sec    15.15k     3.91k   25.09k    82.35%
  215143 requests in 2.04s, 40.42MB read
Requests/sec: 105528.83
Transfer/sec:     19.83MB
```
