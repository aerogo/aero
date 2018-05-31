package aero_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aerogo/aero"
	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/parse/buffer"
)

func TestBodyReader(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		body := ctx.Request().Body()
		bodyText, _ := body.String()
		return ctx.Text(bodyText)
	})

	// Get response
	requestBody := []byte(helloWorld)
	request, _ := http.NewRequest("GET", "/", buffer.NewReader(requestBody))
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, helloWorld, response.Body.String())
}

func TestBodyReaderJSON(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
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
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "value", response.Body.String())
}

func TestBodyReaderErrors(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx *aero.Context) string {
		body := ctx.Request().Body()

		// JSON
		bodyJSON, err := body.JSON()

		assert.Error(t, err)
		assert.Nil(t, bodyJSON)

		// JSON object
		bodyJSONObject, err := body.JSONObject()

		assert.Error(t, err)
		assert.Nil(t, bodyJSONObject)

		return ctx.Text(helloWorld)
	})

	app.Get("/json-object", func(ctx *aero.Context) string {
		body := ctx.Request().Body()
		bodyJSONObject, err := body.JSONObject()

		assert.Error(t, err)
		assert.Nil(t, bodyJSONObject)

		return ctx.Text(helloWorld)
	})

	// No body
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	// Invalid JSON
	request, _ = http.NewRequest("GET", "/", bytes.NewReader([]byte("{")))
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	// Not a JSON object
	request, _ = http.NewRequest("GET", "/json-object", bytes.NewReader([]byte("123")))
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
}
