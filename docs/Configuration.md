# Configuration

Aero allows you to configure your server via a `config.json` file in your project directory.

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

## push

Specifies resources that you want to be HTTP/2 pushed on first load:

```json
{
	"push": [
		"/scripts.js",
		"/image.webp"
	]
}
```

These resources will be queried by synthetic requests to your request handler and then pushed to the client.

# Guidelines

The following fields are not required, but can be set via the configuration. There is a small chance that they will be removed in a future update.

## title

Your public website title.

```json
{
	"title": "My Awesome Site!"
}
```

This is only a guideline. The actual field value is not used anywhere in the server code. The field is usually used in template files.

## domain

The website domain you are using in production.

```json
{
	"domain": "example.com"
}
```

This is only a guideline. The actual field value is not used anywhere in the server code.
