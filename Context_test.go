package aero_test

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aerogo/session"
	qt "github.com/frankban/quicktest"
	jsoniter "github.com/json-iterator/go"

	"github.com/aerogo/aero"
)

func TestContextResponseHeader(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		ctx.Response().Header().Set("X-Custom", "42")
		return ctx.Text(helloWorld)
	})

	// Get response
	response := getResponse(app, "/")

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, helloWorld)
	c.Assert(response.Header().Get("X-Custom"), qt.Equals, "42")
}

func TestContextError(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.Error(http.StatusUnauthorized, "Not authorized", errors.New("Not logged in"))
	})

	app.Get("/explanation-only", func(ctx *aero.Context) error {
		return ctx.Error(http.StatusUnauthorized, "Not authorized", nil)
	})

	app.Get("/unknown-error", func(ctx *aero.Context) error {
		return ctx.Error(http.StatusUnauthorized)
	})

	// Verify response with known error
	c := qt.New(t)
	response := getResponse(app, "/")
	c.Assert(response.Code, qt.Equals, http.StatusUnauthorized)
	c.Assert(response.Body.String(), qt.Contains, "Not logged in")

	// Verify response with explanation only
	response = getResponse(app, "/explanation-only")
	c.Assert(response.Code, qt.Equals, http.StatusUnauthorized)
	c.Assert(response.Body.String(), qt.Contains, "Not authorized")

	// Verify response with unknown error
	response = getResponse(app, "/unknown-error")
	c.Assert(response.Code, qt.Equals, http.StatusUnauthorized)
	c.Assert(response.Body.String(), qt.Contains, http.StatusText(http.StatusUnauthorized))
}

func TestContextURI(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/uri", func(ctx *aero.Context) error {
		return ctx.Text(ctx.URI())
	})

	app.Get("/set-uri", func(ctx *aero.Context) error {
		ctx.SetURI("/hello")
		return ctx.Text(ctx.URI())
	})

	// Verify response with read-only URI
	c := qt.New(t)
	response := getResponse(app, "/uri")
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Contains, "/uri")

	// Verify response with modified URI
	response = getResponse(app, "/set-uri")
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Contains, "/hello")
}

func TestContextRealIP(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/ip", func(ctx *aero.Context) error {
		return ctx.Text(ctx.RealIP())
	})

	// Get response
	response := getResponse(app, "/ip")

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Contains, "")
}

func TestContextSession(t *testing.T) {
	app := aero.New()
	c := qt.New(t)

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		c.Assert(ctx.HasSession(), qt.Equals, false)
		ctx.Session().Set("custom", helloWorld)
		c.Assert(ctx.HasSession(), qt.Equals, true)

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	// Get response
	response := getResponse(app, "/")

	// Verify response
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, helloWorld)
}

func TestContextSessionInvalidCookie(t *testing.T) {
	app := aero.New()
	c := qt.New(t)

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		c.Assert(ctx.HasSession(), qt.Equals, false)
		ctx.Session().Set("custom", helloWorld)
		c.Assert(ctx.HasSession(), qt.Equals, true)

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
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, helloWorld)
}

func TestContextSessionValidCookie(t *testing.T) {
	app := aero.New()
	c := qt.New(t)

	// Register routes
	app.Get("/1", func(ctx *aero.Context) error {
		c.Assert(ctx.HasSession(), qt.Equals, false)
		ctx.Session().Set("custom", helloWorld)
		c.Assert(ctx.HasSession(), qt.Equals, true)
		c.Assert(ctx.Session().ID(), qt.Equals, ctx.Session().GetString("sid"))

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	app.Get("/2", func(ctx *aero.Context) error {
		c.Assert(ctx.HasSession(), qt.Equals, true)
		c.Assert(ctx.Session().ID(), qt.Equals, ctx.Session().GetString("sid"))

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	app.Get("/3", func(ctx *aero.Context) error {
		c.Assert(ctx.Session().ID(), qt.Equals, ctx.Session().GetString("sid"))

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	// Create request 1
	request1, _ := http.NewRequest("GET", "/1", nil)

	// Get response 1
	response1 := httptest.NewRecorder()
	app.Handler().ServeHTTP(response1, request1)

	// Verify response 1
	c.Assert(response1.Code, qt.Equals, http.StatusOK)
	c.Assert(response1.Body.String(), qt.Equals, helloWorld)

	setCookie := response1.Header().Get("Set-Cookie")
	c.Assert(setCookie, qt.Not(qt.Equals), "")
	c.Assert(setCookie, qt.Contains, "sid=")

	cookieParts := strings.Split(setCookie, ";")
	sidLine := strings.TrimSpace(cookieParts[0])
	sidParts := strings.Split(sidLine, "=")
	sid := sidParts[1]
	c.Assert(session.IsValidID(sid), qt.Equals, true)

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
	c.Assert(response2.Code, qt.Equals, http.StatusOK)
	c.Assert(response2.Body.String(), qt.Equals, helloWorld)

	// Create request 3
	request3, _ := http.NewRequest("GET", "/3", nil)
	request3.AddCookie(&http.Cookie{
		Name:  "sid",
		Value: sid,
	})

	// Get response 3
	response3 := httptest.NewRecorder()
	app.Handler().ServeHTTP(response3, request3)

	// Verify response 3
	c.Assert(response3.Code, qt.Equals, http.StatusOK)
	c.Assert(response3.Body.String(), qt.Equals, helloWorld)
}

func TestContextContentTypes(t *testing.T) {
	app := aero.New()

	// Register routes
	app.Get("/json", func(ctx *aero.Context) error {
		return ctx.JSON(app.Config)
	})

	app.Get("/jsonld", func(ctx *aero.Context) error {
		return ctx.JSONLinkedData(app.Config)
	})

	app.Get("/html", func(ctx *aero.Context) error {
		return ctx.HTML("<html></html>")
	})

	app.Get("/css", func(ctx *aero.Context) error {
		return ctx.CSS("body{}")
	})

	app.Get("/js", func(ctx *aero.Context) error {
		return ctx.JavaScript("console.log(42)")
	})

	app.Get("/files/*file", func(ctx *aero.Context) error {
		return ctx.File(ctx.Get("file"))
	})

	// Get responses
	responseJSON := getResponse(app, "/json")
	responseJSONLD := getResponse(app, "/jsonld")
	responseHTML := getResponse(app, "/html")
	responseCSS := getResponse(app, "/css")
	responseJS := getResponse(app, "/js")
	responseFile := getResponse(app, "/files/Application.go")
	responseMediaFile := getResponse(app, "/files/docs/media/usage.apng")

	// Verify JSON response
	c := qt.New(t)
	json, err := jsoniter.Marshal(app.Config)
	c.Assert(err, qt.IsNil)
	c.Assert(responseJSON.Code, qt.Equals, http.StatusOK)
	c.Assert(responseJSON.Body.Bytes(), qt.DeepEquals, json)
	c.Assert(responseJSON.Header().Get("Content-Type"), qt.Matches, `application/json.*`)

	// Verify JSON+LD response
	c.Assert(responseJSONLD.Code, qt.Equals, http.StatusOK)
	c.Assert(responseJSONLD.Body.Bytes(), qt.DeepEquals, json)
	c.Assert(responseJSONLD.Header().Get("Content-Type"), qt.Matches, `application/ld\+json.*`)

	// Verify HTML response
	c.Assert(responseHTML.Code, qt.Equals, http.StatusOK)
	c.Assert(responseHTML.Body.String(), qt.Equals, "<html></html>")
	c.Assert(responseHTML.Header().Get("Content-Type"), qt.Matches, `text/html.*`)

	// Verify CSS response
	c.Assert(responseCSS.Code, qt.Equals, http.StatusOK)
	c.Assert(responseCSS.Body.String(), qt.Equals, "body{}")
	c.Assert(responseCSS.Header().Get("Content-Type"), qt.Matches, `text/css.*`)

	// Verify JS response
	c.Assert(responseJS.Code, qt.Equals, http.StatusOK)
	c.Assert(responseJS.Body.String(), qt.Equals, "console.log(42)")
	c.Assert(responseJS.Header().Get("Content-Type"), qt.Matches, `application/javascript.*`)

	// Verify file response
	appSourceCode, err := ioutil.ReadFile("Application.go")
	c.Assert(err, qt.IsNil)
	c.Assert(responseFile.Code, qt.Equals, http.StatusOK)
	c.Assert(responseFile.Body.Bytes(), qt.DeepEquals, appSourceCode)
	c.Assert(responseFile.Header().Get("Content-Type"), qt.Matches, `text/plain.*`)

	// Verify media file response
	imageData, err := ioutil.ReadFile("docs/media/usage.apng")
	c.Assert(err, qt.IsNil)
	c.Assert(responseMediaFile.Code, qt.Equals, http.StatusOK)
	c.Assert(responseMediaFile.Body.Bytes(), qt.DeepEquals, imageData)
	c.Assert(responseMediaFile.Header().Get("Content-Type"), qt.Equals, `image/apng`)
}

func TestContextReader(t *testing.T) {
	app := aero.New()
	config, err := jsoniter.MarshalToString(app.Config)
	c := qt.New(t)
	c.Assert(err, qt.IsNil)

	// ReadAll
	app.Get("/readall", func(ctx *aero.Context) error {
		reader, writer := io.Pipe()

		go func() {
			defer writer.Close()
			encoder := jsoniter.NewEncoder(writer)
			err := encoder.Encode(app.Config)
			c.Assert(err, qt.IsNil)
		}()

		return ctx.ReadAll(reader)
	})

	// Reader
	app.Get("/reader", func(ctx *aero.Context) error {
		reader, writer := io.Pipe()

		go func() {
			defer writer.Close()
			encoder := jsoniter.NewEncoder(writer)
			err := encoder.Encode(app.Config)
			c.Assert(err, qt.IsNil)
		}()

		return ctx.Reader(reader)
	})

	// ReadSeeker
	app.Get("/readseeker", func(ctx *aero.Context) error {
		return ctx.ReadSeeker(strings.NewReader(config))
	})

	routes := []string{
		"/readall",
		"/reader",
		"/readseeker",
	}

	for _, route := range routes {
		// Verify response
		response := getResponse(app, route)
		c.Assert(response.Code, qt.Equals, http.StatusOK)
		c.Assert(strings.TrimSpace(response.Body.String()), qt.Equals, config)
	}
}

func TestContextHTTP2Push(t *testing.T) {
	app := aero.New()
	app.Config.Push = append(app.Config.Push, "/pushed.css")

	// Register routes
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.HTML("<html></html>")
	})

	app.Get("/pushed.css", func(ctx *aero.Context) error {
		return ctx.CSS("body{}")
	})

	// Add no-op push condition
	app.AddPushCondition(func(ctx *aero.Context) bool {
		return true
	})

	// Get response
	response := getResponse(app, "/")

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, "<html></html>")
}

func TestContextGetInt(t *testing.T) {
	app := aero.New()
	c := qt.New(t)

	// Register route
	app.Get("/:number", func(ctx *aero.Context) error {
		number, err := ctx.GetInt("number")
		c.Assert(err, qt.IsNil)
		c.Assert(number, qt.Not(qt.Equals), 0)

		return ctx.Text(strconv.Itoa(number * 2))
	})

	// Get response
	response := getResponse(app, "/21")

	// Verify response
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, "42")
}

func TestContextUserAgent(t *testing.T) {
	app := aero.New()
	agent := "Luke Skywalker"

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		userAgent := ctx.UserAgent()
		return ctx.Text(userAgent)
	})

	// Create request
	request, _ := http.NewRequest("GET", "/", nil)
	request.Header.Set("User-Agent", agent)

	// Get response
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, agent)
}

func TestContextRedirect(t *testing.T) {
	app := aero.New()
	c := qt.New(t)

	// Register routes
	app.Get("/permanent", func(ctx *aero.Context) error {
		return ctx.RedirectPermanently("/target")
	})

	app.Get("/temporary", func(ctx *aero.Context) error {
		return ctx.Redirect("/target")
	})

	// Get temporary response
	response := getResponse(app, "/temporary")

	// Verify response
	c.Assert(response.Code, qt.Equals, http.StatusFound)
	c.Assert(response.Body.String(), qt.Equals, "")

	// Get permanent response
	response = getResponse(app, "/permanent")

	// Verify response
	c.Assert(response.Code, qt.Equals, http.StatusMovedPermanently)
	c.Assert(response.Body.String(), qt.Equals, "")
}

func TestContextQuery(t *testing.T) {
	app := aero.New()
	search := "Luke Skywalker"
	c := qt.New(t)

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		search := ctx.Query("search")
		return ctx.Text(search)
	})

	// Create request
	request, _ := http.NewRequest("GET", "/?search="+search, nil)

	// Get response
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Equals, search)
}

func TestContextEventStream(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
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
	request, _ := http.NewRequest("GET", "/", nil)
	ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
	defer cancel()
	request = request.WithContext(ctx)

	// Get response
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
}

func TestBigResponse(t *testing.T) {
	text := strings.Repeat("Hello World", 1000000)
	app := aero.New()
	c := qt.New(t)

	// Make sure GZip is enabled
	c.Assert(app.Config.GZip, qt.Equals, true)

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.Text(text)
	})

	// Get response
	response := getResponse(app, "/")

	// Verify the response
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Header().Get("Content-Encoding"), qt.Equals, "gzip")
}

func BenchmarkHelloWorld(b *testing.B) {
	text := "Hello World"
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.Text(text)
	})

	// Create request
	request, _ := http.NewRequest("GET", "/", nil)
	handler := app.Handler()

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

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.Text(text)
	})

	// Create request
	request, _ := http.NewRequest("GET", "/", nil)
	request.Header.Set("Accept-Encoding", "gzip")
	handler := app.Handler()

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
	text := strings.Repeat("HelloWorld", 1000000)
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.Text(text)
	})

	// Create request and record response
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify the response
	c := qt.New(t)
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Header().Get("Content-Encoding"), qt.Equals, "")
}

func TestBigResponse304(t *testing.T) {
	text := strings.Repeat("HelloWorld", 1000000)
	app := aero.New()
	c := qt.New(t)

	// Register route
	app.Get("/", func(ctx *aero.Context) error {
		return ctx.Text(text)
	})

	// Create request and record response
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	etag := response.Header().Get("ETag")

	// Verify the response
	c.Assert(response.Code, qt.Equals, http.StatusOK)
	c.Assert(response.Body.String(), qt.Not(qt.Equals), "")

	// Set if-none-match to the etag we just received
	request, _ = http.NewRequest("GET", "/", nil)
	request.Header.Set("If-None-Match", etag)
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify the response
	c.Assert(response.Code, qt.Equals, 304)
	c.Assert(response.Body.String(), qt.Equals, "")
}
