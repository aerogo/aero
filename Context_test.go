package aero_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/aerogo/session"
	jsoniter "github.com/json-iterator/go"

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
	response := getResponse(app, "/")

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, helloWorld, response.Body.String())
	assert.Equal(t, "42", response.Header().Get("X-Custom"))
}

func TestContextError(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		return ctx.Error(http.StatusUnauthorized, "Not authorized", errors.New("Not logged in"))
	})

	app.Get("/explanation-only", func(ctx *aero.Context) string {
		return ctx.Error(http.StatusUnauthorized, "Not authorized", nil)
	})

	app.Get("/unknown-error", func(ctx *aero.Context) string {
		return ctx.Error(http.StatusUnauthorized)
	})

	// Verify response with known error
	response := getResponse(app, "/")
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, response.Body.String(), "Not logged in")

	// Verify response with explanation only
	response = getResponse(app, "/explanation-only")
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, response.Body.String(), "Not authorized")

	// Verify response with unknown error
	response = getResponse(app, "/unknown-error")
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, response.Body.String(), "Unknown error")
}

func TestContextURI(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/uri", func(ctx *aero.Context) string {
		return ctx.URI()
	})

	app.Get("/set-uri", func(ctx *aero.Context) string {
		ctx.SetURI("/hello")
		return ctx.URI()
	})

	// Verify response with read-only URI
	response := getResponse(app, "/uri")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), "/uri")

	// Verify response with modified URI
	response = getResponse(app, "/set-uri")
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), "/hello")
}

func TestContextRealIP(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/ip", func(ctx *aero.Context) string {
		return ctx.RealIP()
	})

	// Get response
	response := getResponse(app, "/ip")

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), "")
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
	response := getResponse(app, "/")

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
		assert.Equal(t, ctx.Session().GetString("sid"), ctx.Session().ID())

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	app.Get("/2", func(ctx *aero.Context) string {
		assert.Equal(t, true, ctx.HasSession())
		assert.Equal(t, ctx.Session().GetString("sid"), ctx.Session().ID())

		return ctx.Text(ctx.Session().GetString("custom"))
	})

	app.Get("/3", func(ctx *aero.Context) string {
		assert.Equal(t, ctx.Session().GetString("sid"), ctx.Session().ID())

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
	assert.Equal(t, http.StatusOK, response3.Code)
	assert.Equal(t, helloWorld, response3.Body.String())
}

func TestContextContentTypes(t *testing.T) {
	app := aero.New()

	// Register routes
	app.Get("/json", func(ctx *aero.Context) string {
		return ctx.JSON(app.Config)
	})

	app.Get("/jsonld", func(ctx *aero.Context) string {
		return ctx.JSONLinkedData(app.Config)
	})

	app.Get("/html", func(ctx *aero.Context) string {
		return ctx.HTML("<html></html>")
	})

	app.Get("/css", func(ctx *aero.Context) string {
		return ctx.CSS("body{}")
	})

	app.Get("/js", func(ctx *aero.Context) string {
		return ctx.JavaScript("console.log(42)")
	})

	app.Get("/files/*file", func(ctx *aero.Context) string {
		return ctx.File(ctx.Get("file"))
	})

	// Get responses
	responseJSON := getResponse(app, "/json")
	responseJSONLD := getResponse(app, "/jsonld")
	responseHTML := getResponse(app, "/html")
	responseCSS := getResponse(app, "/css")
	responseJS := getResponse(app, "/js")
	responseFile := getResponse(app, "/files/Application.go")
	responseMediaFile := getResponse(app, "/files/docs/usage.gif")

	// Verify JSON response
	json, _ := jsoniter.Marshal(app.Config)
	assert.Equal(t, http.StatusOK, responseJSON.Code)
	assert.Equal(t, json, responseJSON.Body.Bytes())
	assert.Contains(t, responseJSON.Header().Get("Content-Type"), "application/json")

	// Verify JSON+LD response
	assert.Equal(t, http.StatusOK, responseJSONLD.Code)
	assert.Equal(t, json, responseJSONLD.Body.Bytes())
	assert.Contains(t, responseJSONLD.Header().Get("Content-Type"), "application/ld+json")

	// Verify HTML response
	assert.Equal(t, http.StatusOK, responseHTML.Code)
	assert.Equal(t, "<html></html>", responseHTML.Body.String())
	assert.Contains(t, responseHTML.Header().Get("Content-Type"), "text/html")

	// Verify CSS response
	assert.Equal(t, http.StatusOK, responseCSS.Code)
	assert.Equal(t, "body{}", responseCSS.Body.String())
	assert.Contains(t, responseCSS.Header().Get("Content-Type"), "text/css")

	// Verify JS response
	assert.Equal(t, http.StatusOK, responseJS.Code)
	assert.Equal(t, "console.log(42)", responseJS.Body.String())
	assert.Contains(t, responseJS.Header().Get("Content-Type"), "application/javascript")

	// Verify file response
	appSourceCode, _ := ioutil.ReadFile("Application.go")
	assert.Equal(t, http.StatusOK, responseFile.Code)
	assert.Equal(t, appSourceCode, responseFile.Body.Bytes())
	assert.Contains(t, responseFile.Header().Get("Content-Type"), "text/plain")

	// Verify media file response
	imageData, _ := ioutil.ReadFile("docs/usage.gif")
	assert.Equal(t, http.StatusOK, responseMediaFile.Code)
	assert.Equal(t, imageData, responseMediaFile.Body.Bytes())
	assert.Contains(t, responseMediaFile.Header().Get("Content-Type"), "image/gif")
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

	// Add no-op push condition
	app.AddPushCondition(func(ctx *aero.Context) bool {
		return true
	})

	// Get response
	response := getResponse(app, "/")

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "<html></html>", response.Body.String())
}

func TestContextGetInt(t *testing.T) {
	app := aero.New()

	// Register route
	app.Get("/:number", func(ctx *aero.Context) string {
		number, err := ctx.GetInt("number")
		assert.NoError(t, err)
		assert.NotZero(t, number)

		return ctx.Text(strconv.Itoa(number * 2))
	})

	// Get response
	response := getResponse(app, "/21")

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "42", response.Body.String())
}

func TestContextUserAgent(t *testing.T) {
	app := aero.New()
	agent := "Luke Skywalker"

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
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
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, agent, response.Body.String())
}

func TestContextRedirect(t *testing.T) {
	app := aero.New()

	// Register routes
	app.Get("/permanent", func(ctx *aero.Context) string {
		return ctx.RedirectPermanently("/target")
	})

	app.Get("/temporary", func(ctx *aero.Context) string {
		return ctx.Redirect("/target")
	})

	// Get temporary response
	response := getResponse(app, "/temporary")

	// Verify response
	assert.Equal(t, http.StatusFound, response.Code)
	assert.Equal(t, "", response.Body.String())

	// Get permanent response
	response = getResponse(app, "/permanent")

	// Verify response
	assert.Equal(t, http.StatusMovedPermanently, response.Code)
	assert.Equal(t, "", response.Body.String())
}

func TestContextQuery(t *testing.T) {
	app := aero.New()
	search := "Luke Skywalker"

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		search := ctx.Query("search")
		return ctx.Text(search)
	})

	// Create request
	request, _ := http.NewRequest("GET", "/?search="+search, nil)

	// Get response
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, search, response.Body.String())
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
	response := getResponse(app, "/")

	// Verify the response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "gzip", response.Header().Get("Content-Encoding"))
}

func TestBigResponseNoGzip(t *testing.T) {
	text := strings.Repeat("Hello World", 1000000)
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		return ctx.Text(text)
	})

	// Create request and record response
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify the response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "", response.Header().Get("Content-Encoding"))
}

func TestBigResponse304(t *testing.T) {
	text := strings.Repeat("Hello World", 1000000)
	app := aero.New()

	// Register route
	app.Get("/", func(ctx *aero.Context) string {
		return ctx.Text(text)
	})

	// Create request and record response
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	etag := response.Header().Get("ETag")

	// Verify the response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NotEmpty(t, response.Body.String())

	// Set if-none-match to the etag we just received
	request, _ = http.NewRequest("GET", "/", nil)
	request.Header.Set("If-None-Match", etag)
	response = httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	// Verify the response
	assert.Equal(t, 304, response.Code)
	assert.Empty(t, response.Body.String())
}
