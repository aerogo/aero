package aero_test

import (
	"net/http"
	"testing"

	"github.com/aerogo/aero"
	"github.com/akyoto/assert"
)

func TestApplicationMiddleware(t *testing.T) {
	app := aero.New()

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

	response := test(app, "/")

	assert.Equal(t, response.Code, http.StatusPermanentRedirect)
	assert.Equal(t, response.Body.String(), helloWorld)
}

func TestApplicationMiddlewareSkipNext(t *testing.T) {
	app := aero.New()

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

	response := test(app, "/")
	assert.Equal(t, response.Body.String(), "")
}
