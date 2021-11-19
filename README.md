![Aero Go Logo](docs/media/aero.go.png)

[![Godoc][godoc-image]][godoc-url]
[![Report][report-image]][report-url]
[![Tests][tests-image]][tests-url]
[![Coverage][coverage-image]][coverage-url]
[![Sponsor][sponsor-image]][sponsor-url]

Aero is a high-performance web server with a clean API.

## Installation

```shell
go get -u github.com/aerogo/aero/...
```

## Benchmarks

[![Web server performance](docs/media/server-performance.png)](docs/Benchmarks.md)

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
- Provides a context interface for custom contexts

## Links

- [API](docs/API.md)
- [Configuration](docs/Configuration.md)
- [Benchmarks](docs/Benchmarks.md)

[godoc-image]: https://godoc.org/github.com/aerogo/aero?status.svg
[godoc-url]: https://godoc.org/github.com/aerogo/aero
[report-image]: https://goreportcard.com/badge/github.com/aerogo/aero
[report-url]: https://goreportcard.com/report/github.com/aerogo/aero
[tests-image]: https://cloud.drone.io/api/badges/aerogo/aero/status.svg
[tests-url]: https://cloud.drone.io/aerogo/aero
[coverage-image]: https://codecov.io/gh/aerogo/aero/graph/badge.svg
[coverage-url]: https://codecov.io/gh/aerogo/aero
[sponsor-image]: https://img.shields.io/badge/github-donate-green.svg
[sponsor-url]: https://github.com/users/akyoto/sponsorship
