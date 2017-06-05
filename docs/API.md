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

## Middleware
```go
app.Use(func(ctx *aero.Context, next func()) {
	start := time.Now()
	next()
	responseTime := time.Since(start)
	fmt.Println(responseTime)
})
```