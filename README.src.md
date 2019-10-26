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
- Provides http and https listener
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

{go:footer}
