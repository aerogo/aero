package aero_test

import (
	"net/http"
	"testing"

	"github.com/aerogo/aero"
	qt "github.com/frankban/quicktest"
)

func TestApplicationMiddleware(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.Text(helloWorld)
	})

	// Register middleware
	app.Use(func(ctx *aero.Context, next func()) {
		ctx.StatusCode = http.StatusPermanentRedirect
		next()
	})

	// Get response
	response := getResponse(app, "/")

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusPermanentRedirect)
	c.Assert(response.Body.String(), qt.Equals, helloWorld)
}

func TestApplicationMiddlewareSkipNext(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.Text(helloWorld)
	})

	// Register middleware
	app.Use(func(ctx *aero.Context, next func()) {
		// Not calling next() will stop the response chain
	})

	// Get response
	response := getResponse(app, "/")

	// Verify response
	c := qt.New(t)
	c.Assert(response.Body.String(), qt.Equals, "")
}
