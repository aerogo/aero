# Configuration

Aero allows you to configure your server via a `config.json` file in your project directory. You can also set these directly via `app.Config` in your code.

## ports

The ports that will be used for the HTTP and HTTPS listener. Note that both ports return nearly the same content, there are no automatic redirects (possible change in future updates).

```json
{
	"ports": {
		"http": 4000,
		"https": 4001
	}
}
```

## gzip

Enable or disable gzip compression for your server. Setting this to `true` is highly recommended as it will only trigger on responses that are worth compressing and only when the client supports it.

```json
{
	"gzip": true
}
```

## timeouts

Defines timeouts in nanoseconds.

```json
{
	"timeouts": {
		"idle": 1000000000,
		"readHeader": 1000000000,
		"write": 1000000000,
		"shutdown": 1000000000
    }
}
```

## push

Specifies resources that you want to be HTTP/2 pushed on `ctx.HTML` responses:

```json
{
	"push": [
		"/scripts.js",
		"/image.webp"
	]
}
```

These resources will be queried by synthetic requests to your request handler and then pushed to the client asynchronously.
