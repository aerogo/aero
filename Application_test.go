package aero_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/aerogo/aero"
	"github.com/aerogo/http/client"
	qt "github.com/frankban/quicktest"
)

const helloWorld = "Hello World"

func TestApplicationGet(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.Text(helloWorld)
	})

	// Get response
	response := getResponse(app, "/")

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, helloWorld)
}

func TestApplicationPost(t *testing.T) {
	app := aero.New()

	// Register route
	app.Post("/", func(ctx *aero.Context) error {
		return ctx.Text(helloWorld)
	})

	// Get response
	request, _ := http.NewRequest("POST", "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, helloWorld)
}

func TestApplicationDelete(t *testing.T) {
	app := aero.New()

	// Register route
	app.Delete("/", func(ctx *aero.Context) error {
		return ctx.Text(helloWorld)
	})

	// Get response
	request, _ := http.NewRequest("DELETE", "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, helloWorld)
}

func TestApplicationRewrite(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/hello", func(ctx *aero.Context) error {
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
	response := getResponse(app, "/")

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, helloWorld)
}

func TestApplicationLoadConfig(t *testing.T) {
	app := aero.New()
	workingDirectory, _ := os.Getwd()
	c := qt.New(t)

	err := os.Chdir("testdata")
	c.Assert(err, qt.IsNil)

	app.Load()

	err = os.Chdir(workingDirectory)
	c.Assert(err, qt.IsNil)
}

func TestApplicationRun(t *testing.T) {
	app := aero.New()
	c := qt.New(t)

	// When frontpage is requested, kill the server
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.HTML(helloWorld)
	})

	// When the server is started, we request the frontpage
	app.OnStart(func() {
		_, err := client.Get(fmt.Sprintf("http://localhost:%d/", app.Config.Ports.HTTP)).End()
		c.Assert(err, qt.IsNil)

		err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		c.Assert(err, qt.IsNil)
	})

	// When the server ends, check elapsed time
	app.OnEnd(func() {
		elapsed := time.Since(app.StartTime())
		c.Assert(elapsed < 2*time.Second, qt.Equals, true)
	})

	// Run
	app.Run()
}

func TestApplicationRunHTTPS(t *testing.T) {
	app := aero.New()
	app.Security.Load("testdata/fullchain.pem", "testdata/privkey.pem")
	c := qt.New(t)

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.HTML(helloWorld)
	})

	// When the server is started, we request the frontpage
	app.OnStart(func() {
		_, err := client.Get(fmt.Sprintf("http://localhost:%d/", app.Config.Ports.HTTP)).End()
		c.Assert(err, qt.IsNil)

		_, err = client.Get(fmt.Sprintf("https://localhost:%d/", app.Config.Ports.HTTPS)).End()
		c.Assert(err, qt.IsNil)

		go func() {
			err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			c.Assert(err, qt.IsNil)
		}()
	})

	// Run
	app.Run()
}

func TestApplicationRouteTests(t *testing.T) {
	app := aero.New()

	app.Get("/user/:id", func(ctx *aero.Context) error {
		return ctx.Text(ctx.Get("id"))
	})

	app.Get("/untested/:untested", func(ctx *aero.Context) error {
		return ctx.Text(ctx.Get("untested"))
	})

	// Specify a test route explicitly
	app.Test("/user/:id", "/user/123")
	app.TestRoutes()
}

func TestApplicationUnavailablePort(t *testing.T) {
	defer func() {
		_ = recover()
		// c.Assert(r, qt.Not(qt.IsNil))
		// c.Assert(r.(error).Error(), qt.Contains, "bind: permission denied")
	}()

	app := aero.New()
	app.Config.Ports.HTTP = 80
	app.Config.Ports.HTTPS = 443
	app.ListenAndServe()
}

// getResponse sends a request to the server and returns the response.
func getResponse(app *aero.Application, route string) *httptest.ResponseRecorder {
	// Create request
	request, _ := http.NewRequest("GET", route, nil)
	request.Header.Set("Accept-Encoding", "gzip")

	// Get response
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	return response
}
