package aero_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/aerogo/aero"
	"github.com/akyoto/assert"
)

func TestRequest(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		request := ctx.Request()

		assert.Equal(t, request.Header("Accept-Encoding"), "gzip")
		assert.Equal(t, request.Host(), "example.com")
		assert.Equal(t, request.Protocol(), "HTTP/1.1")
		assert.Equal(t, request.Method(), "GET")
		assert.Equal(t, request.Path(), "/")
		assert.Equal(t, request.Scheme(), "http")

		return ctx.Text(helloWorld)
	})

	response := test(app, "/")
	assert.Equal(t, response.Code, http.StatusOK)
}

func TestMultiRequest(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		return ctx.Text(strings.Repeat(helloWorld, 1000))
	})

	// Repeating the request will trigger the gzip writer pool
	for i := 0; i < 10; i++ {
		test(app, "/")
	}
}
