# Benchmarks

## Routing

Simple routing test on an Intel Core i7-8700 using `Hello World` as output:

```text
λ wrk -t8 -c400 -d2s http://localhost:4000/
Running 2s test @ http://localhost:4000/
  8 threads and 400 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.41ms    1.42ms  14.88ms   86.18%
    Req/Sec    42.17k     5.43k   59.83k    71.25%
  672784 requests in 2.05s, 102.66MB read
Requests/sec: 327837.89
Transfer/sec:     50.02MB
```

## Conclusion

Considering that it's possible to reach over 300k requests per second on a decent CPU, Aero which is based on `net/http` from the standard library will not be the bottleneck of your website.

Databases and complex application logic are usually the bigger factor in your web application performance.