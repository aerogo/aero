![Aero Go Logo](docs/images/aero.go.png)

[![Godoc reference][godoc-image]][godoc-url]
[![Discord][discord-image]][discord-url]

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

- Makes it easy to reach top scores in [Google PageSpeed](https://developers.google.com/speed/pagespeed/insights/), [Mozilla Observatory](https://observatory.mozilla.org/) and [Google Lighthouse](https://developers.google.com/web/tools/lighthouse/)
- Optimizes your website for high latency environments (mobile networks)
- Simplifies single page app development
- Has a strict content security policy to improve security
- Shows response time and size for your routes
- Calculates ETags out of the box (client caching)
- Provides http: and https: listener
- Finishes ongoing requests on a server shutdown
- Supports HTTP/2, IPv6 and Web Manifest
- Supports HTTP/2 push of resources (add resource URL to "push" in config.json)
- Can run standalone without `nginx` babysitting it

## Documentation

- [API](docs/API.md)
- [Benchmarks](docs/Benchmarks.md)

## Chat

Feel free to join us on [Discord][discord-url] (better than Slack and IRC).

![Discord](https://puu.sh/y62bO/bfb44dbd11.png)

## In development

This is an ongoing project. Use at your own risk.

---

[![By Eduard Urbach](http://forthebadge.com/images/badges/built-with-love.svg)](https://github.com/blitzprog)

[godoc-image]: https://godoc.org/github.com/aerogo/aero?status.svg
[godoc-url]: https://godoc.org/github.com/aerogo/aero
[discord-image]: https://img.shields.io/badge/discord-aero-738bd7.svg
[discord-url]: https://discord.gg/vyk2MnK
