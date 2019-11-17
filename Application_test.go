package aero_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/aerogo/aero"
	"github.com/aerogo/http/client"
	"github.com/akyoto/assert"
)

const helloWorld = "Hello World"

func TestApplicationAny(t *testing.T) {
	app := aero.New()

	app.Any("/", func(ctx aero.Context) error {
		return ctx.Text(helloWorld)
	})

	app.OnError(func(ctx aero.Context, err error) {
		t.Fatal(err)
	})

	methods := []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
	}

	for _, method := range methods {
		// Test existing route
		request := httptest.NewRequest(method, "/", nil)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)

		assert.Equal(t, response.Code, http.StatusOK)
		assert.Equal(t, response.Body.String(), helloWorld)

		// Test non-existing route
		request = httptest.NewRequest(method, "/404", nil)
		response = httptest.NewRecorder()
		app.ServeHTTP(response, request)

		assert.Equal(t, response.Code, http.StatusNotFound)
		assert.Equal(t, response.Body.String(), "")
	}
}

func TestApplicationRewrite(t *testing.T) {
	app := aero.New()

	app.Get("/hello", func(ctx aero.Context) error {
		return ctx.Text(helloWorld)
	})

	// Rewrite route
	app.Rewrite(func(ctx aero.RewriteContext) {
		if ctx.Path() == "/" {
			ctx.SetPath("/hello")
			return
		}
	})

	response := test(app, "/")

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Body.String(), helloWorld)
}

func TestApplicationLoadConfig(t *testing.T) {
	app := aero.New()
	workingDirectory, _ := os.Getwd()

	err := os.Chdir("testdata")
	assert.Nil(t, err)

	app.Load()

	err = os.Chdir(workingDirectory)
	assert.Nil(t, err)
}

func TestApplicationRun(t *testing.T) {
	start := time.Now()
	app := aero.New()

	// When frontpage is requested, kill the server
	app.Get("/", func(ctx aero.Context) error {
		return ctx.HTML(helloWorld)
	})

	// When the server is started, we request the frontpage
	app.OnStart(func() {
		_, err := client.Get(fmt.Sprintf("http://localhost:%d/", app.Config.Ports.HTTP)).End()
		assert.Nil(t, err)

		err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		assert.Nil(t, err)
	})

	// When the server ends, check elapsed time
	app.OnEnd(func() {
		elapsed := time.Since(start)
		assert.Equal(t, elapsed < 2*time.Second, true)
	})

	// Run
	app.Run()
}

func TestApplicationRunHTTPS(t *testing.T) {
	app := aero.New()
	app.Security.Load("testdata/fullchain.pem", "testdata/privkey.pem")

	app.Get("/", func(ctx aero.Context) error {
		return ctx.HTML(helloWorld)
	})

	// When the server is started, we request the frontpage
	app.OnStart(func() {
		_, err := client.Get(fmt.Sprintf("http://localhost:%d/", app.Config.Ports.HTTP)).End()
		assert.Nil(t, err)

		_, err = client.Get(fmt.Sprintf("https://localhost:%d/", app.Config.Ports.HTTPS)).End()
		assert.Nil(t, err)

		go func() {
			err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			assert.Nil(t, err)
		}()
	})

	// Run
	app.Run()
}

func TestApplicationRouteTests(t *testing.T) {
	app := aero.New()

	app.Get("/user/:id", func(ctx aero.Context) error {
		return ctx.Text(ctx.Get("id"))
	})

	app.Get("/untested/:untested", func(ctx aero.Context) error {
		return ctx.Text(ctx.Get("untested"))
	})

	// Specify a test route explicitly
	app.Test("/user/:id", "/user/123")
	app.TestRoutes()
}

func TestApplicationUnavailablePort(t *testing.T) {
	defer func() {
		_ = recover()
		// assert.NotNil(t, r)
		// assert.Contains(t, r.(error).Error(), "bind: permission denied")
	}()

	app := aero.New()
	app.Config.Ports.HTTP = 80
	app.Config.Ports.HTTPS = 443
	app.ListenAndServe()
}

// test sends a request to the server and returns the response.
func test(app http.Handler, route string) *httptest.ResponseRecorder {
	request := httptest.NewRequest("GET", route, nil)
	request.Header.Set("Accept-Encoding", "gzip")

	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	return response
}

func TestApplicationOnError(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		return errors.New("something happened")
	})

	app.OnError(func(ctx aero.Context, err error) {
		assert.Equal(t, err.Error(), "something happened")
	})

	test(app, "/")
}
