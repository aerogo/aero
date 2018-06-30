package aero_test

import (
	"net/http"
	"testing"

	"github.com/aerogo/aero"
	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx *aero.Context) string {
		request := ctx.Request()

		assert.NotEmpty(t, request.Header())
		assert.Empty(t, request.Host())
		assert.Equal(t, "HTTP/1.1", request.Protocol())
		assert.Equal(t, "GET", request.Method())
		assert.NotNil(t, request.URL())
		assert.Equal(t, "/", request.URL().Path)

		return ctx.Text(helloWorld)
	})

	response := getResponse(app, "/")
	assert.Equal(t, http.StatusOK, response.Code)
}
