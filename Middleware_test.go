package aero_test

import (
	"net/http"
	"testing"

	"github.com/aerogo/aero"
	"github.com/stretchr/testify/assert"
)

func TestApplicationMiddleware(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
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
	assert.Equal(t, http.StatusPermanentRedirect, response.Code)
	assert.Equal(t, helloWorld, response.Body.String())
}

func TestApplicationMiddlewareSkipNext(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		return ctx.Text(helloWorld)
	})

	// Register middleware
	app.Use(func(ctx *aero.Context, next func()) {
		// Not calling next() will stop the response chain
	})

	// Get response
	response := getResponse(app, "/")

	// Verify response
	assert.Equal(t, "", response.Body.String())
}
