![Aero Go Logo](docs/images/aero.go.png)

[![Godoc reference][godoc-image]][godoc-url]
[![Discord][discord-image]][discord-url]

Aero is a high-performance web server with a clean API for web development.

## Installation

```bash
go get github.com/aerogo/aero
```

## Features

- Makes it easy to reach 100/100 scores in [Google PageSpeed](https://developers.google.com/speed/pagespeed/insights/), [Mozilla Observatory](https://observatory.mozilla.org/) and [Google Lighthouse](https://developers.google.com/web/tools/lighthouse/)
- Optimizes your website for high latency environments (mobile networks)
- Simplifies single page app development
- Supports HTTP/2 & IPv6
- Provides HTTP and HTTPS listener
- Has a strict content security policy to improve security
- Shows response time and size for your routes
- Finishes ongoing requests on a server shutdown
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
