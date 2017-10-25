# Benchmarks

## Routing

Very simple routing test using `Hello World` as output:

### AeroGo (go 1.9)

```
Running 2s test @ http://localhost:4000/hello
  8 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.20ms    1.38ms  23.81ms   79.24%
    Req/Sec    22.89k     5.67k   73.95k    98.77%
  368755 requests in 2.10s, 61.19MB read
Requests/sec: 175594.86
Transfer/sec:     29.14MB
```

### AeroJS (node.js 8.8)

```
Running 2s test @ http://localhost:4000
  8 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     9.78ms    5.40ms 228.37ms   95.41%
    Req/Sec     4.47k     1.68k   23.12k    93.79%
  71622 requests in 2.10s, 52.73MB read
Requests/sec:  34114.80
Transfer/sec:     25.12MB
```