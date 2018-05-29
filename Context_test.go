package aero_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aerogo/session"

	"github.com/aerogo/aero"
	"github.com/stretchr/testify/assert"
)

func TestContextResponseHeader(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		ctx.Response().Header().Set("X-Custom", "42")
		return ctx.Text(helloWorld)
	})

	// Get response
	response := request(app, "/")

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, helloWorld, response.Body.String())
	assert.Equal(t, "42", response.Header().Get("X-Custom"))
}

func TestContextSession(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		assert.Equal(t, false, ctx.HasSession())
		ctx.Session().Set("custom", helloWorld)
		assert.Equal(t, true, ctx.HasSession())

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	// Get response
	response := request(app, "/")

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, helloWorld, response.Body.String())
}

func TestContextSessionInvalidCookie(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		assert.Equal(t, false, ctx.HasSession())
		ctx.Session().Set("custom", helloWorld)
		assert.Equal(t, true, ctx.HasSession())

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	// Create request
	request, _ := http.NewRequest("GET", "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	request.Header.Set("Cookie", "sid=invalid")

	// Get response
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, helloWorld, response.Body.String())
}

func TestContextSessionValidCookie(t *testing.T) {
	app := aero.New()

	// Register routes
	app.Get("/1", func(ctx *aero.Context) string {
		assert.Equal(t, false, ctx.HasSession())
		ctx.Session().Set("custom", helloWorld)
		assert.Equal(t, true, ctx.HasSession())

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	app.Get("/2", func(ctx *aero.Context) string {
		assert.Equal(t, true, ctx.HasSession())

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	// Create request 1
	request1, _ := http.NewRequest("GET", "/1", nil)

	// Get response 1
	response1 := httptest.NewRecorder()
	app.Handler().ServeHTTP(response1, request1)

	// Verify response 1
	assert.Equal(t, http.StatusOK, response1.Code)
	assert.Equal(t, helloWorld, response1.Body.String())

	setCookie := response1.Header().Get("Set-Cookie")
	assert.NotEmpty(t, setCookie)
	assert.Contains(t, setCookie, "sid=")

	cookieParts := strings.Split(setCookie, ";")
	sidLine := strings.TrimSpace(cookieParts[0])
	sidParts := strings.Split(sidLine, "=")
	sid := sidParts[1]
	assert.True(t, session.IsValidID(sid))

	// Create request 2
	request2, _ := http.NewRequest("GET", "/2", nil)
	request2.AddCookie(&http.Cookie{
		Name:  "sid",
		Value: sid,
	})

	// Get response 2
	response2 := httptest.NewRecorder()
	app.Handler().ServeHTTP(response2, request2)

	// Verify response 2
	assert.Equal(t, http.StatusOK, response2.Code)
	assert.Equal(t, helloWorld, response2.Body.String())
}

func TestContextContentTypes(t *testing.T) {
	app := aero.New()

	// Register routes
	app.Get("/json", func(ctx *aero.Context) string {
		return ctx.JSON(app.Config)
	})

	app.Get("/html", func(ctx *aero.Context) string {
		return ctx.HTML("<html></html>")
	})

	app.Get("/css", func(ctx *aero.Context) string {
		return ctx.CSS("body{}")
	})

	app.Get("/files/*file", func(ctx *aero.Context) string {
		return ctx.File(strings.TrimPrefix(ctx.Get("file"), "/"))
	})

	// Get responses
	responseJSON := request(app, "/json")
	responseHTML := request(app, "/html")
	responseCSS := request(app, "/css")
	responseFile := request(app, "/files/Application.go")

	// Verify JSON response
	json, _ := json.Marshal(app.Config)
	assert.Equal(t, http.StatusOK, responseJSON.Code)
	assert.Equal(t, json, responseJSON.Body.Bytes())
	assert.Contains(t, responseJSON.Header().Get("Content-Type"), "application/json")

	// Verify HTML response
	assert.Equal(t, http.StatusOK, responseHTML.Code)
	assert.Equal(t, "<html></html>", responseHTML.Body.String())
	assert.Contains(t, responseHTML.Header().Get("Content-Type"), "text/html")

	// Verify CSS response
	assert.Equal(t, http.StatusOK, responseCSS.Code)
	assert.Equal(t, "body{}", responseCSS.Body.String())
	assert.Contains(t, responseCSS.Header().Get("Content-Type"), "text/css")

	// // Verify File response
	appSourceCode, _ := ioutil.ReadFile("Application.go")
	assert.Equal(t, http.StatusOK, responseFile.Code)
	assert.Equal(t, appSourceCode, responseFile.Body.Bytes())
	assert.Contains(t, responseFile.Header().Get("Content-Type"), "text/plain")
}

func TestContextHTTP2Push(t *testing.T) {
	app := aero.New()
	app.Config.Push = append(app.Config.Push, "/pushed.css")

	// Register routes
	app.Get("/", func(ctx *aero.Context) string {
		return ctx.HTML("<html></html>")
	})

	app.Get("/pushed.css", func(ctx *aero.Context) string {
		return ctx.CSS("body{}")
	})

	// Get response
	response := request(app, "/")

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "<html></html>", response.Body.String())
}
