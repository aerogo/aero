package aero_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aerogo/aero"
	"github.com/stretchr/testify/assert"
)

const helloWorld = "Hello World"

func TestApplicationGet(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		return ctx.Text(helloWorld)
	})

	// Get response
	response := request(app, "/")

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, helloWorld, response.Body.String())
}

func TestApplicationPost(t *testing.T) {
	app := aero.New()

	// Register route
	app.Post("/", func(ctx *aero.Context) string {
		return ctx.Text(helloWorld)
	})

	// Get response
	request, _ := http.NewRequest("POST", "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, helloWorld, response.Body.String())
}

func TestApplicationRewrite(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/hello", func(ctx *aero.Context) string {
		return ctx.Text(helloWorld)
	})

	// Rewrite route
	app.Rewrite(func(ctx *aero.RewriteContext) {
		if ctx.URI() == "/" {
			ctx.SetURI("/hello")
			return
		}
	})

	// Get response
	response := request(app, "/")

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, helloWorld, response.Body.String())
}

func TestBigResponse(t *testing.T) {
	text := strings.Repeat("Hello World", 1000000)
	app := aero.New()

	// Make sure GZip is enabled
	assert.Equal(t, true, app.Config.GZip)

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		return ctx.Text(text)
	})

	// Get response
	response := request(app, "/")

	// Verify the response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "gzip", response.Header().Get("Content-Encoding"))
}

// request sends a request to the server and returns the response.
func request(app *aero.Application, route string) *httptest.ResponseRecorder {
	// Create request
	request, _ := http.NewRequest("GET", route, nil)
	request.Header.Set("Accept-Encoding", "gzip")

	// Get response
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	return response
}
