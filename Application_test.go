package aero_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aerogo/aero"
	"github.com/stretchr/testify/assert"
)

func TestApplicationGet(t *testing.T) {
	helloWorld := "Hello World"
	app := aero.New()

	app.Get("/", func(ctx *aero.Context) string {
		return ctx.Text(helloWorld)
	})

	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(responseRecorder, request)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	assert.Equal(t, helloWorld, responseRecorder.Body.String())
}

func TestApplicationPost(t *testing.T) {
	helloWorld := "Hello World"
	app := aero.New()

	app.Post("/", func(ctx *aero.Context) string {
		return ctx.Text(helloWorld)
	})

	request, err := http.NewRequest("POST", "/", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(responseRecorder, request)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	assert.Equal(t, helloWorld, responseRecorder.Body.String())
}

func TestApplicationRewrite(t *testing.T) {
	helloWorld := "Hello World"
	app := aero.New()

	app.Get("/hello", func(ctx *aero.Context) string {
		return ctx.Text(helloWorld)
	})

	app.Rewrite(func(ctx *aero.RewriteContext) {
		if ctx.URI() == "/" {
			ctx.SetURI("/hello")
			return
		}
	})

	request, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	responseRecorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(responseRecorder, request)

	assert.Equal(t, http.StatusOK, responseRecorder.Code)
	assert.Equal(t, helloWorld, responseRecorder.Body.String())
}
