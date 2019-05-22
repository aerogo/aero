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

## Features

- Makes it easy to reach top scores in [Lighthouse](https://developers.google.com/web/tools/lighthouse/), [PageSpeed](https://developers.google.com/speed/pagespeed/insights/) and [Mozilla Observatory](https://observatory.mozilla.org/)
- Optimized for high latency environments (mobile networks)
- Has a strict content security policy
- Calculates E-Tags out of the box
- Saves you a lot of bandwidth using browser cache validation
- Finishes ongoing requests on a server shutdown
- Automatic HTTP/2 push of configured resources
- Supports session data with custom stores
- Allows pushing live data to the client via SSE
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
- [Discord community](https://discord.gg/V3y4xTY)
- [Twitter account](https://twitter.com/aeroframework)
- [Facebook page](https://www.facebook.com/aeroframework/)

{go:footer}
