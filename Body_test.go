package aero_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aerogo/aero"
	qt "github.com/frankban/quicktest"
	"github.com/tdewolff/parse/buffer"
)

func TestBody(t *testing.T) {
	app := aero.New()
	c := qt.New(t)

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		body := ctx.Request().Body()
		c.Assert(ctx.Request().Body().Reader(), qt.Not(qt.IsNil))
		bodyText, _ := body.String()
		return ctx.Text(bodyText)
	})

	// Get response
	requestBody := []byte(helloWorld)
	request, _ := http.NewRequest("GET", "/", buffer.NewReader(requestBody))
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, helloWorld)
}

func TestBodyJSON(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		body := ctx.Request().Body()
		obj, _ := body.JSONObject()
		return ctx.Text(fmt.Sprint(obj["key"]))
	})

	// Get response
	requestBody := []byte(`{"key":"value"}`)
	request, _ := http.NewRequest("GET", "/", buffer.NewReader(requestBody))
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, "value")
}

func TestBodyErrors(t *testing.T) {
	app := aero.New()
	c := qt.New(t)

	app.Get("/", func(ctx *aero.Context) error {
		body := ctx.Request().Body()
		bodyJSON, err := body.JSON()

		c.Assert(err, qt.Not(qt.IsNil))
		c.Assert(bodyJSON, qt.IsNil)

		return ctx.Text(helloWorld)
	})

	app.Get("/json-object", func(ctx *aero.Context) error {
		body := ctx.Request().Body()
		bodyJSONObject, err := body.JSONObject()

		c.Assert(err, qt.Not(qt.IsNil))
		c.Assert(bodyJSONObject, qt.IsNil)

		return ctx.Text(helloWorld)
	})

	// No body
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	c.Assert(response.Code, qt.Equals, http.StatusOK)

	// Invalid JSON
	request, _ = http.NewRequest("GET", "/", strings.NewReader("{"))
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	c.Assert(response.Code, qt.Equals, http.StatusOK)

	// Not a JSON object
	request, _ = http.NewRequest("GET", "/json-object", strings.NewReader("{"))
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
}
