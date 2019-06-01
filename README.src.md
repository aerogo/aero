![Aero Go Logo](docs/media/aero.go.png)

{go:header}

Aero is a high-performance web server with a clean API for web development.

{go:install}

## Usage

![Aero usage](docs/media/usage.apng)

Run this in an empty directory:

```bash
aero -new
```

Now you can build your app with `go build` or use the [run](https://github.com/aerogo/run) development server.

## Benchmarks

```text
BenchmarkAeroStatic-12            100000             16643 ns/op             700 B/op          0 allocs/op
BenchmarkAeroGitHubAPI-12          50000             27487 ns/op            1409 B/op          0 allocs/op
BenchmarkAeroGplusAPI-12         1000000              1390 ns/op              69 B/op          0 allocs/op
BenchmarkAeroParseAPI-12          500000              2558 ns/op             138 B/op          0 allocs/op

BenchmarkEchoStatic-12             50000             30702 ns/op            1950 B/op        157 allocs/op
BenchmarkEchoGitHubAPI-12          30000             45431 ns/op            2782 B/op        203 allocs/op
BenchmarkEchoGplusAPI-12          500000              2500 ns/op             173 B/op         13 allocs/op
BenchmarkEchoParseAPI-12          300000              4234 ns/op             323 B/op         26 allocs/op

BenchmarkGinStatic-12              50000             37885 ns/op            8231 B/op        157 allocs/op
BenchmarkGinGitHubAPI-12           30000             55092 ns/op           10903 B/op        203 allocs/op
BenchmarkGinGplusAPI-12           500000              3059 ns/op             693 B/op         13 allocs/op
BenchmarkGinParseAPI-12           300000              5687 ns/op            1363 B/op         26 allocs/op
```

You can run these by yourself using [web-framework-benchmark](https://github.com/akyoto/web-framework-benchmark). Read more [here](docs/Benchmarks.md).

## Features

- Makes it easy to reach top scores in [Lighthouse](https://developers.google.com/web/tools/lighthouse/), [PageSpeed](https://developers.google.com/speed/pagespeed/insights/) and [Mozilla Observatory](https://observatory.mozilla.org/)
- Optimized for low latency
- Best practices are enabled by default
- Has a strict content security policy
- Calculates E-Tags out of the box
- Saves you a lot of bandwidth using browser cache validation
- Finishes ongoing requests on a server shutdown
- Lets you push resources via HTTP/2
- Supports session data with custom stores
- Allows sending live data to the client via SSE
- Provides http and https listener
- Shows response time and size for your routes
- Can run standalone without `nginx` babysitting it

## Optional

- [pack](https://github.com/aerogo/pack) to compile Pixy, Scarlet and JS assets in record time
- [run](https://github.com/aerogo/run) which automatically restarts your server on code/template/style changes
- [pixy](https://github.com/aerogo/pixy) as a high-performance template engine similar to Jade/Pug
- [scarlet](https://github.com/aerogo/scarlet) as an aggressively compressing stylesheet preprocessor
- [nano](https://github.com/aerogo/nano) as a fast, decentralized and git-trackable database
- [layout](https://github.com/aerogo/layout) as a layout system
- [manifest](https://github.com/aerogo/manifest) to load and manipulate web manifests
- [markdown](https://github.com/aerogo/markdown) as an overly simplified markdown wrapper
- [graphql](https://github.com/aerogo/graphql) to automatically implement your GraphQL API
- [packet](https://github.com/aerogo/packet) as a way to send TCP/UDP messages between nodes
- [http](https://github.com/aerogo/http) as an HTTP client with a simple and clean API
- [log](https://github.com/aerogo/log) for simple & performant logging

## Documentation

- [API](docs/API.md)
- [Configuration](docs/Configuration.md)
- [Benchmarks](docs/Benchmarks.md)

## Others

- [Slides for OWDDM talk](https://docs.google.com/presentation/d/166I69goLEVuvuFeeRfUu8c5lwl2_HAeSi2SZyzIuEKg/edit) (Osaka, May 2018)
- [Discord](https://discord.gg/V3y4xTY)
- [Twitter](https://twitter.com/aeroframework)
- [Facebook](https://www.facebook.com/aeroframework/)

{go:footer}
