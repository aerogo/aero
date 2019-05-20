package aero_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/aerogo/aero"
	qt "github.com/frankban/quicktest"
)

func TestRequest(t *testing.T) {
	app := aero.New()
	c := qt.New(t)

	app.Get("/", func(ctx *aero.Context) string {
		request := ctx.Request()

		c.Assert(request.Header(), qt.Not(qt.Equals), "")
		c.Assert(request.Host(), qt.Equals, "")
		c.Assert(request.Protocol(), qt.Equals, "HTTP/1.1")
		c.Assert(request.Method(), qt.Equals, "GET")
		c.Assert(request.URL(), qt.Not(qt.IsNil))
		c.Assert(request.URL().Path, qt.Equals, "/")

		return ctx.Text(helloWorld)
	})

	response := getResponse(app, "/")
	c.Assert(response.Code, qt.Equals, http.StatusOK)
}

func TestMultiRequest(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx *aero.Context) string {
		return ctx.Text(strings.Repeat(helloWorld, 1000))
	})

	// Repeating the request will trigger the gzip writer pool
	for i := 0; i < 10; i++ {
		getResponse(app, "/")
	}
}
