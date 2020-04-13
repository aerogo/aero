package aero_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aerogo/aero"
	"github.com/aerogo/session"
	"github.com/akyoto/assert"
)

func TestContextResponseHeader(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		ctx.Response().SetHeader("X-Custom", "42")
		return ctx.Text(helloWorld)
	})

	response := test(app, "/")

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Body.String(), helloWorld)
	assert.Equal(t, response.Header().Get("X-Custom"), "42")
}

func TestContextError(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		return ctx.Error(http.StatusUnauthorized, "Not authorized", errors.New("Not logged in"))
	})

	app.Get("/explanation-only", func(ctx aero.Context) error {
		return ctx.Error(http.StatusUnauthorized, "Not authorized", nil)
	})

	app.Get("/unknown-error", func(ctx aero.Context) error {
		return ctx.Error(http.StatusUnauthorized)
	})

	response := test(app, "/")
	assert.Equal(t, response.Code, http.StatusUnauthorized)
	assert.Contains(t, response.Body.String(), "Not logged in")

	response = test(app, "/explanation-only")
	assert.Equal(t, response.Code, http.StatusUnauthorized)
	assert.Contains(t, response.Body.String(), "Not authorized")

	response = test(app, "/unknown-error")
	assert.Equal(t, response.Code, http.StatusUnauthorized)
	assert.Contains(t, response.Body.String(), http.StatusText(http.StatusUnauthorized))
}

func TestContextURI(t *testing.T) {
	app := aero.New()

	app.Get("/uri", func(ctx aero.Context) error {
		return ctx.Text(ctx.Path())
	})

	response := test(app, "/uri")
	assert.Equal(t, response.Code, http.StatusOK)
	assert.Contains(t, response.Body.String(), "/uri")
}

func TestContextRealIP(t *testing.T) {
	app := aero.New()

	app.Get("/ip", func(ctx aero.Context) error {
		return ctx.Text(ctx.IP())
	})

	response := test(app, "/ip")

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Contains(t, response.Body.String(), "")
}

func TestContextSession(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		assert.Equal(t, ctx.HasSession(), false)
		ctx.Session().Set("custom", helloWorld)
		assert.Equal(t, ctx.HasSession(), true)

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	response := test(app, "/")

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Body.String(), helloWorld)
}

func TestContextSessionInvalidCookie(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		assert.Equal(t, ctx.HasSession(), false)
		ctx.Session().Set("custom", helloWorld)
		assert.Equal(t, ctx.HasSession(), true)

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	// Create request
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	request.Header.Set("Cookie", "sid=invalid")

	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Body.String(), helloWorld)
}

func TestContextSessionValidCookie(t *testing.T) {
	app := aero.New()

	app.Get("/1", func(ctx aero.Context) error {
		assert.Equal(t, ctx.HasSession(), false)
		ctx.Session().Set("custom", helloWorld)
		assert.Equal(t, ctx.HasSession(), true)
		assert.Equal(t, ctx.Session().ID(), ctx.Session().GetString("sid"))

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	app.Get("/2", func(ctx aero.Context) error {
		assert.Equal(t, ctx.HasSession(), true)
		assert.Equal(t, ctx.Session().ID(), ctx.Session().GetString("sid"))

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	app.Get("/3", func(ctx aero.Context) error {
		assert.Equal(t, ctx.Session().ID(), ctx.Session().GetString("sid"))

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	// Create request 1
	request1 := httptest.NewRequest("GET", "/1", nil)

	response1 := httptest.NewRecorder()
	app.ServeHTTP(response1, request1)

	assert.Equal(t, response1.Code, http.StatusOK)
	assert.Equal(t, response1.Body.String(), helloWorld)

	setCookie := response1.Header().Get("Set-Cookie")
	assert.NotEqual(t, setCookie, "")
	assert.Contains(t, setCookie, "sid=")

	cookieParts := strings.Split(setCookie, ";")
	sidLine := strings.TrimSpace(cookieParts[0])
	sidParts := strings.Split(sidLine, "=")
	sid := sidParts[1]
	assert.Equal(t, session.IsValidID(sid), true)

	// Create request 2
	request2 := httptest.NewRequest("GET", "/2", nil)
	request2.AddCookie(&http.Cookie{
		Name:  "sid",
		Value: sid,
	})

	response2 := httptest.NewRecorder()
	app.ServeHTTP(response2, request2)

	assert.Equal(t, response2.Code, http.StatusOK)
	assert.Equal(t, response2.Body.String(), helloWorld)

	// Create request 3
	request3 := httptest.NewRequest("GET", "/3", nil)
	request3.AddCookie(&http.Cookie{
		Name:  "sid",
		Value: sid,
	})

	response3 := httptest.NewRecorder()
	app.ServeHTTP(response3, request3)

	assert.Equal(t, response3.Code, http.StatusOK)
	assert.Equal(t, response3.Body.String(), helloWorld)
}

func TestContextContentTypes(t *testing.T) {
	app := aero.New()

	app.Get("/json", func(ctx aero.Context) error {
		return ctx.JSON(app.Config)
	})

	app.Get("/html", func(ctx aero.Context) error {
		return ctx.HTML("<html></html>")
	})

	app.Get("/css", func(ctx aero.Context) error {
		return ctx.CSS("body{}")
	})

	app.Get("/js", func(ctx aero.Context) error {
		return ctx.JavaScript("console.log(42)")
	})

	app.Get("/files/*file", func(ctx aero.Context) error {
		return ctx.File(ctx.Get("file"))
	})

	responseJSON := test(app, "/json")
	responseHTML := test(app, "/html")
	responseCSS := test(app, "/css")
	responseJS := test(app, "/js")
	responseFile := test(app, "/files/Application.go")
	responseMediaFile := test(app, "/files/docs/media/usage.apng")

	// Verify JSON response
	json, err := json.Marshal(app.Config)
	assert.Nil(t, err)
	assert.Equal(t, responseJSON.Code, http.StatusOK)
	assert.DeepEqual(t, responseJSON.Body.Bytes(), json)
	assert.Contains(t, responseJSON.Header().Get("Content-Type"), "application/json")

	// Verify HTML response
	assert.Equal(t, responseHTML.Code, http.StatusOK)
	assert.Equal(t, responseHTML.Body.String(), "<html></html>")
	assert.Contains(t, responseHTML.Header().Get("Content-Type"), "text/html")

	// Verify CSS response
	assert.Equal(t, responseCSS.Code, http.StatusOK)
	assert.Equal(t, responseCSS.Body.String(), "body{}")
	assert.Contains(t, responseCSS.Header().Get("Content-Type"), "text/css")

	// Verify JS response
	assert.Equal(t, responseJS.Code, http.StatusOK)
	assert.Equal(t, responseJS.Body.String(), "console.log(42)")
	assert.Contains(t, responseJS.Header().Get("Content-Type"), "text/javascript")

	// Verify file response
	appSourceCode, err := ioutil.ReadFile("Application.go")
	assert.Nil(t, err)
	assert.Equal(t, responseFile.Code, http.StatusOK)
	assert.DeepEqual(t, responseFile.Body.Bytes(), appSourceCode)
	assert.Contains(t, responseFile.Header().Get("Content-Type"), "text/plain")

	// Verify media file response
	imageData, err := ioutil.ReadFile("docs/media/usage.apng")
	assert.Nil(t, err)
	assert.Equal(t, responseMediaFile.Code, http.StatusOK)
	assert.DeepEqual(t, responseMediaFile.Body.Bytes(), imageData)
	assert.Equal(t, responseMediaFile.Header().Get("Content-Type"), "image/apng")
}

func TestContextReader(t *testing.T) {
	app := aero.New()
	config, err := json.Marshal(app.Config)
	assert.Nil(t, err)

	// ReadAll
	app.Get("/readall", func(ctx aero.Context) error {
		reader, writer := io.Pipe()

		go func() {
			defer writer.Close()
			encoder := json.NewEncoder(writer)
			err := encoder.Encode(app.Config)
			assert.Nil(t, err)
		}()

		return ctx.ReadAll(reader)
	})

	// Reader
	app.Get("/reader", func(ctx aero.Context) error {
		reader, writer := io.Pipe()

		go func() {
			defer writer.Close()
			encoder := json.NewEncoder(writer)
			err := encoder.Encode(app.Config)
			assert.Nil(t, err)
		}()

		return ctx.Reader(reader)
	})

	// ReadSeeker
	app.Get("/readseeker", func(ctx aero.Context) error {
		return ctx.ReadSeeker(bytes.NewReader(config))
	})

	routes := []string{
		"/readall",
		"/reader",
		"/readseeker",
	}

	for _, route := range routes {
		response := test(app, route)
		assert.Equal(t, response.Code, http.StatusOK)
		assert.DeepEqual(t, bytes.TrimSpace(response.Body.Bytes()), config)
	}
}

func TestContextHTTP2Push(t *testing.T) {
	app := aero.New()
	app.Config.Push = append(app.Config.Push, "/pushed.css")

	app.Get("/", func(ctx aero.Context) error {
		return ctx.HTML("<html></html>")
	})

	app.Get("/pushed.css", func(ctx aero.Context) error {
		return ctx.CSS("body{}")
	})

	// Add no-op push condition
	app.AddPushCondition(func(ctx aero.Context) bool {
		return true
	})

	response := test(app, "/")

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Body.String(), "<html></html>")
}

func TestContextGetInt(t *testing.T) {
	app := aero.New()

	app.Get("/:number", func(ctx aero.Context) error {
		number, err := ctx.GetInt("number")
		assert.Nil(t, err)
		assert.NotEqual(t, number, 0)

		return ctx.Text(strconv.Itoa(number * 2))
	})

	response := test(app, "/21")

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Body.String(), "42")
}

func TestContextUserAgent(t *testing.T) {
	app := aero.New()
	agent := "Luke Skywalker"

	app.Get("/", func(ctx aero.Context) error {
		userAgent := ctx.Request().Header("User-Agent")
		return ctx.Text(userAgent)
	})

	// Create request
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("User-Agent", agent)

	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Body.String(), agent)
}

func TestContextRedirect(t *testing.T) {
	app := aero.New()

	app.Get("/permanent", func(ctx aero.Context) error {
		return ctx.Redirect(http.StatusMovedPermanently, "/target")
	})

	app.Get("/temporary", func(ctx aero.Context) error {
		return ctx.Redirect(http.StatusFound, "/target")
	})

	// Get temporary response
	response := test(app, "/temporary")

	assert.Equal(t, response.Code, http.StatusFound)
	assert.Equal(t, response.Body.String(), "")

	// Get permanent response
	response = test(app, "/permanent")

	assert.Equal(t, response.Code, http.StatusMovedPermanently)
	assert.Equal(t, response.Body.String(), "")
}

func TestContextQuery(t *testing.T) {
	app := aero.New()
	search := "Skywalker"

	app.Get("/", func(ctx aero.Context) error {
		search := ctx.Query("search")
		return ctx.Text(search)
	})

	// Create request
	request := httptest.NewRequest("GET", "/?search="+search, nil)

	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Body.String(), search)
}

func TestContextEventStream(t *testing.T) {
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		stream := aero.NewEventStream()

		go func() {
			for {
				select {
				case <-stream.Closed:
					close(stream.Events)
					return

				case <-time.After(10 * time.Millisecond):
					stream.Events <- &aero.Event{
						Name: "ping",
						Data: "{}",
					}

					stream.Events <- &aero.Event{
						Name: "ping",
						Data: []byte("{}"),
					}

					stream.Events <- &aero.Event{
						Name: "ping",
						Data: struct {
							Message string `json:"message"`
						}{
							Message: "Hello",
						},
					}

					stream.Events <- &aero.Event{
						Name: "ping",
						Data: nil,
					}
				}
			}
		}()

		return ctx.EventStream(stream)
	})

	// Create request
	request := httptest.NewRequest("GET", "/", nil)
	ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
	defer cancel()
	request = request.WithContext(ctx)

	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	assert.Equal(t, response.Code, http.StatusOK)
}

func TestBigResponse(t *testing.T) {
	text := strings.Repeat("Hello World", 1000000)
	app := aero.New()

	// Make sure GZip is enabled
	assert.Equal(t, app.Config.GZip, true)

	app.Get("/", func(ctx aero.Context) error {
		return ctx.Text(text)
	})

	response := test(app, "/")

	// Verify the response
	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Header().Get("Content-Encoding"), "gzip")
}

func BenchmarkHelloWorld(b *testing.B) {
	text := "Hello World"
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		return ctx.Text(text)
	})

	// Create request
	request := httptest.NewRequest("GET", "/", nil)
	handler := app

	// Benchmark settings
	b.ReportAllocs()
	b.ResetTimer()

	// Run the benchmark
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
		}
	})
}

func BenchmarkBigResponse(b *testing.B) {
	text := strings.Repeat("HelloWorld", 1000)
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		return ctx.Text(text)
	})

	// Create request
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	handler := app

	// Benchmark settings
	b.ReportAllocs()
	b.ResetTimer()

	// Run the benchmark
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
		}
	})
}

func TestBigResponseNoGzip(t *testing.T) {
	text := strings.Repeat("HelloWorld", 1000)
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		return ctx.Text(text)
	})

	// Create request and record response
	request := httptest.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	// Verify the response
	assert.Equal(t, response.Code, http.StatusOK)
	assert.Equal(t, response.Header().Get("Content-Encoding"), "")
}

func TestBigResponse304(t *testing.T) {
	text := strings.Repeat("HelloWorld", 1000)
	app := aero.New()

	app.Get("/", func(ctx aero.Context) error {
		return ctx.Text(text)
	})

	// Create request and record response
	request := httptest.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	etag := response.Header().Get("ETag")

	// Verify the response
	assert.Equal(t, response.Code, http.StatusOK)
	assert.NotEqual(t, response.Body.String(), "")

	// Set if-none-match to the etag we just received
	request = httptest.NewRequest("GET", "/", nil)
	request.Header.Set("If-None-Match", etag)
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)

	// Verify the response
	assert.Equal(t, response.Code, 304)
	assert.Equal(t, response.Body.String(), "")
}
