package aero_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aerogo/aero"
	"github.com/akyoto/assert"
)

func TestRequestBody(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		body := ctx.Request().Body()
		assert.NotNil(t, ctx.Request().Body().Reader())
		bodyText, _ := body.String()
		return ctx.Text(bodyText)
	})

	requestBody := []byte(helloWorld)
	request := httptest.NewRequest("GET", "/", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Body.String(), helloWorld)
}

func TestRequestBodyJSON(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		body := ctx.Request().Body()
		obj, _ := body.JSONObject()
		return ctx.Text(fmt.Sprint(obj["key"]))
	})

	requestBody := []byte(`{"key":"value"}`)
	request := httptest.NewRequest("GET", "/", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Body.String(), "value")
}

func TestRequestBodyErrors(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		body := ctx.Request().Body()
		bodyJSON, err := body.JSON()

		assert.NotNil(t, err)
		assert.Nil(t, bodyJSON)

		return ctx.Text(helloWorld)
	})

	app.Get("/json-object", func(ctx aero.Context) error {
		body := ctx.Request().Body()
		bodyJSONObject, err := body.JSONObject()

		assert.NotNil(t, err)
		assert.Nil(t, bodyJSONObject)

		return ctx.Text(helloWorld)
	})

	// No body
	request := httptest.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	assert.Equal(t, response.Code, http.StatusOK)

	// Invalid JSON
	request = httptest.NewRequest("GET", "/", strings.NewReader("{"))
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	assert.Equal(t, response.Code, http.StatusOK)

	// Not a JSON object
	request = httptest.NewRequest("GET", "/json-object", strings.NewReader("{"))
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	assert.Equal(t, response.Code, http.StatusOK)
}
