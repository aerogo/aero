# Benchmarks

## Routing

Simple routing test on an Intel Core i7-8700 using `Hello World` as output:

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

## Conclusion

Aero's router uses highly optimized [radix trees](https://en.wikipedia.org/wiki/Radix_tree) and is therefore extremely fast.
I am fairly confident that it is the fastest router implementation out there.

Using the GitHub API routes as benchmark data, Aero finishes the benchmark in 15 microseconds while [echo](https://github.com/labstack/echo) requires 25 microseconds for the same routes.

Nonetheless, databases and complex application logic are the most important factor in your web application performance.
You shouldn't need to worry about Aero's routing performance at all.
