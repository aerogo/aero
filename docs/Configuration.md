# Configuration

Aero allows you to configure your server via a `config.json` file in your project directory.

## title

Your public website title. Usually used in layout files and in the manifest.

```json
{
	"title": "My Awesome Site!"
}
```

## domain

The website domain you are using in production.

```json
{
	"domain": "example.com"
}
```

## ports

The ports that will be used for the HTTP and HTTPS listener.

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
