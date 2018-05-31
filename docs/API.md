# API

Unless specified otherwise, the API is considered to be stable.

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

## Shortcuts for different content types

```go
app.Get("/", func(ctx *aero.Context) string {
	// Choose one:
	return ctx.HTML("<html></html>")
	return ctx.CSS("body{}")
	return ctx.JavaScript("console.log(42)")
	return ctx.JSON(app.Config)
	return ctx.Text("just some plain text")
})
```

## Middleware

You can run middleware functions that are executed after the routing phase and before the final request handler.

```go
app.Use(func(ctx *aero.Context, next func()) {
	start := time.Now()
	next()
	responseTime := time.Since(start)
	fmt.Println(responseTime)
})
```

It is possible to implement a firewall by filtering requests and denying the `next()` call as the final request handler is also part of the middleware chain. Not calling `next()` means the request will not be handled.

It is also possible to create a request / access log that includes performance timings as shown in the code example above. `ctx.URI()` will retrieve the URI of the request. Note that the actual logging happens **after** the request has been dealt with (`next()` call) which makes it efficient.

## Multiple middleware

You can use multiple `Use()` calls or combine them into a single call:

```go
app.Use(
	First(),
	Second(),
	Third(),
)
```

## Starting server

This will start the server and block until a termination signal arrives.

```go
app.Run()
```

## Layout

The server package by itself does **not** concern itself with the implementation of your layout system but you can add [aerogo/layout](https://github.com/aerogo/layout) to register full-page and content-only routes at once.

```go
// Create a new aerogo/layout
l := layout.New(app)

// Specify the page frame
l.Render = func(ctx *aero.Context, content string) string {
	return "<html><head></head><body>" + content + "</body></html>"
}

// Register the /hello page.
// The page without the page frame will be available under /_/hello
l.Page("/hello", func(ctx *aero.Context) string {
	return ctx.HTML("<h1>Hello</h1>")
})
```

## Rewrite

Rewrites the internal URI before routing happens:

```go
app.Rewrite(func(ctx *aero.RewriteContext) {
	uri := ctx.URI()

	if uri == "/old" {
		ctx.SetURI("/new")
		return
	}
})
```

Only one rewrite function can be active in an Application. Multiple calls will overwrite the previously registered function.

## OnStart

Schedules the function to be called when the server has started. Calling `OnStart` multiple times will register multiple callbacks.

```go
app.OnStart(func() {
	// Do something.
})
```

## OnEnd

In case the server is terminated by outside factors such as a kill signal sent by the operating system, you can specify a function to be called in that event. Calling `OnEnd` multiple times will register multiple callbacks.

```go
app.OnEnd(func() {
	// Free up resources.
})
```

## AddPushCondition

By default, HTTP/2 push will only trigger on `text/html` responses. You can add more conditions via:

```go
// Do not use HTTP/2 push on service worker requests
app.AddPushCondition(func(ctx *aero.Context) bool {
	return ctx.Request().Header().Get("X-Source") != "service-worker"
})
```

Returning `true` for a given request will allow the push of resources while returning `false` will cancel the push immediately in the given request.