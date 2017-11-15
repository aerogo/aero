# Configuration

Aero allows you to configure your server via a `config.json` file in your project directory.

The configuration is always incremental. This means you only need to include the settings that you changed. The settings that have not been specified will be loaded from the [default configuration](https://github.com/aerogo/aero/blob/go/Configuration.go#L54-L64).

## Fields

### title

Your public website title. Usually used in layout files and in the manifest.

```json
{
	"title": "My Awesome Site!"
}
```

### domain

The website domain you are using in production.

```json
{
	"domain": "example.com"
}
```
