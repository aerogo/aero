# API

Unless specified otherwise, the API is considered to be stable.

## Creating an app

```go
app := aero.New()
```

## Routing

```go
app.Get("/hello", func(ctx aero.Context) error {
	return ctx.String("Hello World")
})
```

## Routing with parameters

```go
app.Get("/hello/:person", func(ctx aero.Context) error {
	return ctx.String("Hello " + ctx.Get("person"))
})
```

## Shortcuts for different content types

```go
app.Get("/", func(ctx aero.Context) error {
	// Choose one:
	return ctx.HTML("<html></html>")
	return ctx.CSS("body{}")
	return ctx.JavaScript("console.log(42)")
	return ctx.JSON(app.Config)
	return ctx.Text("just some plain text")
})
```

## Starting the server

This will start the server and block until a termination signal arrives.

```go
app.Run()
```

## Middleware

You can run middleware functions that are executed after the routing phase and before the final request handler.
A middleware function is a function that accepts an `aero.Handler` and returns a modified one.
The accepted handler is the next handler in the middleware chain.

```go
app.Use(func(next aero.Handler) {
	return func(ctx aero.Context) error {
		// Measure response time
		start := time.Now()
		err := next(ctx)
		responseTime := time.Since(start)

		// Write it to the log
		fmt.Println(responseTime)

		// Make sure to pass the error back!
		return err
	}
})
```

It is possible to implement a firewall by filtering requests and denying the `next` call as the final request handler is also part of the middleware chain. Not calling `next` means the request will not be handled.

It is also possible to create a request / access log that includes performance timings as shown in the code example above. `ctx.Path()` will retrieve the path of the request. Note that the actual logging happens **after** the request has been dealt with, which makes it efficient.

## Multiple middleware

You can use multiple `Use()` calls or combine them into a single call:

```go
app.Use(
	first,
	second,
	third,
)
```

## Rewrite

Rewrites the internal URI before routing happens:

```go
app.Rewrite(func(ctx aero.RewriteContext) {
	path := ctx.Path()

	if path == "/old" {
		ctx.SetPath("/new")
		return
	}
})
```

Calling `Rewrite` multiple times will register multiple callbacks.

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

## Sessions

You can use `HasSession` and `Session().Modified()` to store the sessions in your preferred backend storage. I highly recommend using [nano](https://github.com/aerogo/nano) with [session-store-nano](https://github.com/aerogo/session-store-nano) for maximum performance.

```go
app.Use(func(next aero.Context) aero.Handler {
	return func(ctx aero.Context) error {
		// Handle the request first.
		err := next(ctx)

		// If the session was modified, store it.
		if ctx.HasSession() && ctx.Session().Modified() {
			ctx.App.Sessions.Store.Set(ctx.Session().ID(), ctx.Session())
		}
		
		return err
	}
})
```

Here is an example of a view counter in an actual request handler:

```go
app.Get("/", func(ctx aero.Context) error {
	// Load number of views
	views := 0
	storedViews := ctx.Session().Get("views")

	if storedViews != nil {
		views = storedViews.(int)
	}

	// Increment
	views++

	// Store number of views
	ctx.Session().Set("views", views)

	// Display current number of views
	return ctx.Text(fmt.Sprintf("%d views", views))
})
```

## EventStream

*SSE (server sent events) have recently been added as an experimental feature. The API is subject to change.*

Using an event stream, you can push data from your server at any time to your client.

```go
app.Get("/events/live", func(ctx aero.Context) error {
	stream := aero.NewEventStream()

	go func() {
		defer println("disconnected")

		for {
			select {
			case <-stream.Closed:
				return

			case <-time.After(1 * time.Second):
				stream.Events <- &aero.Event{
					Name: "ping",
					Data: "Hello World",
				}
			}
		}
	}()

	return ctx.EventStream(stream)
})
```

On the client side, use [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource#Examples) to receive events.

## AddPushCondition

By default, HTTP/2 push for your [configured resources](Configuration.md#push) will only trigger on `text/html` response via `ctx.HTML`. You can add more conditions via:

```go
// Do not use HTTP/2 push on service worker requests.
// Our service worker will add "X-Source" to the headers.
// Skip the push when the header is set.
app.AddPushCondition(func(ctx aero.Context) bool {
	return ctx.Request().Header().Get("X-Source") != "service-worker"
})
```

Returning `true` for a given request will allow the push of resources while returning `false` will cancel the push immediately in the given request.