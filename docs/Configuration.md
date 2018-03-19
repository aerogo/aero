# Configuration

Aero allows you to configure your server via a `config.json` file in your project directory.

## title

Your public website title.

```json
{
	"title": "My Awesome Site!"
}
```

Usually used in layout files and in the default web manifest.

## domain

The website domain you are using in production.

```json
{
	"domain": "example.com"
}
```

This is only a guideline. The actual field value is not used anywhere in the server code.

## ports

The ports that will be used for the HTTP and HTTPS listener. Note that both ports return nearly the same content, there are no automatic redirects.

```json
{
	"ports": {
		"http": 4000,
		"https": 4001
	}
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

## manifest

Specifies [web manifest](https://developer.mozilla.org/en-US/docs/Web/Manifest) fields that should be overwritten:

```json
{
	"manifest": {
		"short_name": "Example",
		"theme_color": "#aabbcc"
	}
}
```
