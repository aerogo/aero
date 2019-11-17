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

## Usage

![Aero usage](docs/media/usage.apng)

Run this in an empty directory:

```bash
aero -new
```

Now you can build your app with `go build` or use the [run](https://github.com/aerogo/run) development server.

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
- Shows response time and size for your routes
- Can run standalone without `nginx` babysitting it

## Links

- [API](docs/API.md)
- [Configuration](docs/Configuration.md)
- [Benchmarks](docs/Benchmarks.md)
- [Slides](https://docs.google.com/presentation/d/166I69goLEVuvuFeeRfUu8c5lwl2_HAeSi2SZyzIuEKg/edit) (Osaka, May 2018)
- [Community](http://t.me/aeroframework) on Telegram

## Optional

- [http](https://github.com/aerogo/http) as an HTTP client with a simple and clean API
- [log](https://github.com/aerogo/log) for simple & performant logging
- [manifest](https://github.com/aerogo/manifest) to load and manipulate web manifests
- [markdown](https://github.com/aerogo/markdown) as an overly simplified markdown wrapper
- [nano](https://github.com/aerogo/nano) as a fast, decentralized and git-trackable database
- [pack](https://github.com/aerogo/pack) to compile Pixy, Scarlet and JS assets in record time
- [packet](https://github.com/aerogo/packet) as a way to send TCP/UDP messages between nodes
- [pixy](https://github.com/aerogo/pixy) as a high-performance template engine similar to Jade/Pug
- [run](https://github.com/aerogo/run) which automatically restarts your server on code/template/style changes
- [scarlet](https://github.com/aerogo/scarlet) as an aggressively compressing stylesheet preprocessor

## Style

Please take a look at the [style guidelines](https://github.com/akyoto/quality/blob/master/STYLE.md) if you'd like to make a pull request.

## Sponsors

| [![Cedric Fung](https://avatars3.githubusercontent.com/u/2269238?s=70&v=4)](https://github.com/cedricfung) | [![Scott Rayapoullé](https://avatars3.githubusercontent.com/u/11772084?s=70&v=4)](https://github.com/soulcramer) | [![Eduard Urbach](https://avatars3.githubusercontent.com/u/438936?s=70&v=4)](https://eduardurbach.com) |
| --- | --- | --- |
| [Cedric Fung](https://github.com/cedricfung) | [Scott Rayapoullé](https://github.com/soulcramer) | [Eduard Urbach](https://eduardurbach.com) |

Want to see [your own name here?](https://github.com/users/akyoto/sponsorship)

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
