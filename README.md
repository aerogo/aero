![Aero Go Logo](docs/images/aero.go.png)

[![Godoc reference][godoc-image]][godoc-url]
[![Go report card][goreportcard-image]][goreportcard-url]
[![Tests][travis-image]][travis-url]
[![Code coverage][codecov-image]][codecov-url]
[![License][license-image]][license-url]

Aero is a high-performance web server with a clean API for web development.

## Installation

```bash
go get -u github.com/aerogo/aero
go install github.com/aerogo/aero/cmd/aero
```

## Usage

![Aero usage](docs/usage.gif)

Run this in an empty directory:

```bash
aero -newapp
```

Now you can build your app with `go build` or use the more advanced [run](https://github.com/aerogo/run) tool.

## Features

- Makes it easy to reach top scores in [Lighthouse](https://developers.google.com/web/tools/lighthouse/), [PageSpeed](https://developers.google.com/speed/pagespeed/insights/) and [Mozilla Observatory](https://observatory.mozilla.org/)
- Optimized for high latency environments (mobile networks)
- Has a strict content security policy
- Calculates E-Tags out of the box
- Finishes ongoing requests on a server shutdown
- Supports HTTP/2, IPv6 and Web Manifest
- Automatic HTTP/2 push of configured resources
- Supports session data with custom stores
- Provides http and https listener
- Shows response time and size for your routes
- Can run standalone without `nginx` babysitting it

## Optional

- [layout](https://github.com/aerogo/layout) as a layout system
- [pack](https://github.com/aerogo/pack) to compile Pixy, Scarlet and JS assets in record time
- [run](https://github.com/aerogo/run) which automatically restarts your server on code/template/style changes
- [pixy](https://github.com/aerogo/pixy) as a high-performance template engine similar to Jade/Pug
- [scarlet](https://github.com/aerogo/scarlet) as an aggressively compressing stylesheet preprocessor
- [nano](https://github.com/aerogo/nano) as a fast, decentralized and git-trackable database
- [api](https://github.com/aerogo/api) to automatically implement your REST API routes
- [markdown](https://github.com/aerogo/markdown) as an overly simplified markdown wrapper
- [http](https://github.com/aerogo/http) as an HTTP client with a simple and clean API
- [log](https://github.com/aerogo/log) for simple & performant logging

## Documentation

- [API](docs/API.md)
- [Configuration](docs/Configuration.md)
- [Benchmarks](docs/Benchmarks.md)

## Others

- [Slides for OWDDM talk](https://docs.google.com/presentation/d/166I69goLEVuvuFeeRfUu8c5lwl2_HAeSi2SZyzIuEKg/edit) (Osaka, May 2018)
- [Discord community][discord-url]
- [Twitter account](https://twitter.com/aeroframework)
- [Facebook page](https://www.facebook.com/aeroframework/)

## Author

| [![Eduard Urbach on Twitter](https://gravatar.com/avatar/16ed4d41a5f244d1b10de1b791657989?s=70)](https://twitter.com/eduardurbach "Follow @eduardurbach on Twitter") |
|---|
| [Eduard Urbach](https://eduardurbach.com) |

[godoc-image]: https://godoc.org/github.com/aerogo/aero?status.svg
[godoc-url]: https://godoc.org/github.com/aerogo/aero
[goreportcard-image]: https://goreportcard.com/badge/github.com/aerogo/aero
[goreportcard-url]: https://goreportcard.com/report/github.com/aerogo/aero
[travis-image]: https://travis-ci.org/aerogo/aero.svg?branch=master
[travis-url]: https://travis-ci.org/aerogo/aero
[codecov-image]: https://codecov.io/gh/aerogo/aero/branch/master/graph/badge.svg
[codecov-url]: https://codecov.io/gh/aerogo/aero
[sourcegraph-image]: https://sourcegraph.com/github.com/aerogo/aero/-/badge.svg
[sourcegraph-url]: https://sourcegraph.com/github.com/aerogo/aero?badge
[license-image]: https://img.shields.io/badge/license-MIT-blue.svg
[license-url]: https://github.com/aerogo/aero/blob/master/LICENSE
[discord-image]: https://img.shields.io/badge/discord-aero-738bd7.svg
[discord-url]: https://discord.gg/vyk2MnK
