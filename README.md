# Aero
This project is work in progress.

## Goals

- [x] Reach 100/100 [Google PageSpeed](https://developers.google.com/speed/pagespeed/insights/) score
- [x] Reach 100/100 [Mozilla Observatory](https://observatory.mozilla.org/) score
- [ ] Reach 100/100 [Google Lighthouse](https://developers.google.com/web/tools/lighthouse/) score

## Benchmark
Simple routing test using `Hello World` as output (data payload too small for gzip):

*TODO: Post benchmark with ~50 KB of data (performance difference is crazy).*

In a real-world example AeroGo delivers about 8-10 times more requests than AeroJS.

AeroJS (node.js 8.0):
```
# AeroJS HTTP (local)
~ λ b http://localhost:4000/hello
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
# AeroGo HTTP (local)
~ λ b http://dev.local.host/hello
Running 2s test @ http://dev.local.host/hello
  8 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     3.11ms    2.25ms  24.34ms   73.38%
    Req/Sec    16.19k     5.08k   63.64k    91.98%
  261763 requests in 2.08s, 149.78MB read
Requests/sec: 125768.69
Transfer/sec:     71.97MB
```

AeroGo using an SSL certificate (go 1.8.3)
```
# AeroGo HTTPS (local)
~ λ b https://dev.local.host/hello
Running 2s test @ https://dev.local.host/hello
  8 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     3.94ms    2.49ms  27.84ms   72.75%
    Req/Sec    11.93k     2.84k   20.00k    86.76%
  168982 requests in 2.05s, 96.69MB read
Requests/sec:  82567.22
Transfer/sec:     47.25MB
```

## Documentation
* [Application](docs/Application.md)