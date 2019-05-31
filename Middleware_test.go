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
	app.Get("/", func(ctx aero.Context) error {
		return ctx.Text(helloWorld)
	})

	// Register middleware
	app.Use(func(next aero.Handler) aero.Handler {
		return func(ctx aero.Context) error {
			ctx.SetStatus(http.StatusPermanentRedirect)
			return next(ctx)
		}
	})

	// Bind middleware because we are not going to call app.Run
	app.BindMiddleware()

	// Get response
	response := test(app, "/")

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusPermanentRedirect)
	c.Assert(response.Body.String(), qt.Equals, helloWorld)
}

func TestApplicationMiddlewareSkipNext(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx aero.Context) error {
		return ctx.Text(helloWorld)
	})

	// Register middleware
	app.Use(func(next aero.Handler) aero.Handler {
		return func(ctx aero.Context) error {
			// Not calling next(ctx) will stop the response chain
			return nil
		}
	})

	// Bind middleware because we are not going to call app.Run
	app.BindMiddleware()

	// Get response
	response := test(app, "/")

	// Verify response
	c := qt.New(t)
	c.Assert(response.Body.String(), qt.Equals, "")
}
