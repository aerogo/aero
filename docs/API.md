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
app.Get("/hello/:who", func(ctx *aero.Context) string {
	who := ctx.Get("who")
	return ctx.Text("Hello " + who)
})
```

## AJAX routing

Registers `/hello` which renders the full page and `/_/hello` rendering only the page contents.

```go
app.Ajax("/hello", func(ctx *aero.Context) string {
	return ctx.HTML("<h1>Hello</h1>")
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

## Multiple middleware (in one call)

```go
app.Use(
	First(),
	Second(),
	Third(),
)
```

## Rewrite (change URI before routing happens)

```go
app.Rewrite(func(ctx *aero.RewriteContext) {
	uri := ctx.URI()

	if uri == "/old" {
		ctx.SetURI("/new")
		return
	}
})
```