package aero_test

import (
	"encoding/json"
	"net/http"
	"testing"

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

	// app.Get("/files/*file", func(ctx *aero.Context) string {
	// 	return ctx.File(ctx.Get("file"))
	// })

	// Get responses
	responseJSON := request(app, "/json")
	responseHTML := request(app, "/html")
	responseCSS := request(app, "/css")
	// responseFile := request(app, "/files/Application.go")

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
	// appSourceCode, _ := ioutil.ReadFile("Application.go")
	// assert.Equal(t, http.StatusOK, responseFile.Code)
	// assert.Equal(t, appSourceCode, responseFile.Body.Bytes())
	// assert.Contains(t, responseFile.Header().Get("Content-Type"), "text/plain")
}
