# API

## Creating an app

```go
app := aero.New()
```

## Routing

```go
app.Get("/hello", func(ctx *aero.Context) string {
	return ctx.Text("Hello World")
})
```

## Routing with parameters

```go
app.Get("/hello/:person", func(ctx *aero.Context) string {
	return ctx.Text("Hello " + ctx.Get("person"))
})
```

## Middleware

```go
app.Use(func(ctx *aero.Context, next func()) {
	start := time.Now()
	next()
	responseTime := time.Since(start)
	fmt.Println(responseTime)
})
```

## Multiple middleware

You can use multiple `Use()` calls or combine them into a single call:

```go
app.Use(
	First(),
	Second(),
	Third(),
)
```

## Layout

You can set a global wrapper for your HTML content which will only be used in routes registered via `Ajax()`. `Get()` and `Post()` routes are not affected.

```go
app.Layout = func(ctx *aero.Context, content string) string {
	return "<html><head></head><body>" + content + "</body></html>"
}
```

It is highly recommended to use a high-performance renderer like [pixy](https://github.com/aerogo/pixy) via [pack](https://github.com/aerogo/pack) for your HTML templates.

## AJAX routing

Registers `/hello` which renders the full page with `app.Layout` and `/_/hello` rendering only the page contents.

```go
app.Ajax("/hello", func(ctx *aero.Context) string {
	return ctx.HTML("<h1>Hello</h1>")
})
```

## Styling

Calculates the SHA-1 hash of the CSS string, sets `Content-Security-Policy` to only accept this hash as the style and registers the CSS to be sent inline in the very first response to avoid [render blocking CSS](https://developers.google.com/web/fundamentals/performance/critical-rendering-path/render-blocking-css).

```go
app.SetStyle("body{color:red}")
```

## Rewrite

You can change the internal URI before routing happens:

```go
app.Rewrite(func(ctx *aero.RewriteContext) {
	uri := ctx.URI()

	if uri == "/old" {
		ctx.SetURI("/new")
		return
	}
})
```