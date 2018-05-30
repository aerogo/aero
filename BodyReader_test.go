package aero_test

import (
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

		bodyJSON, err := body.JSON()

		assert.Error(t, err)
		assert.Nil(t, bodyJSON)

		bodyJSONObject, err := body.JSONObject()

		assert.Error(t, err)
		assert.Nil(t, bodyJSONObject)

		return ctx.Text(helloWorld)
	})

	response := request(app, "/")
	assert.Equal(t, http.StatusOK, response.Code)
}
