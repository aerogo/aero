![Aero Go Logo](docs/media/aero.go.png)

{go:header}

Aero is a high-performance web server with a clean API.

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
- Provides a context interface for custom contexts
- Shows response time and size for your routes
- Can run standalone without `nginx` babysitting it

## Links

- [API](docs/API.md)
- [Configuration](docs/Configuration.md)
- [Benchmarks](docs/Benchmarks.md)
