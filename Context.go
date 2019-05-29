package aero

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aerogo/session"
	"github.com/akyoto/color"
	"github.com/akyoto/stringutils/unsafe"
	jsoniter "github.com/json-iterator/go"
)

// This should be close to the MTU size of a TCP packet.
// Regarding performance it makes no sense to compress smaller files.
// Bandwidth can be saved however the savings are minimal for small files
// and the overhead of compressing can lead up to a 75% reduction
// in server speed under high load. Therefore in this case
// we're trying to optimize for performance, not bandwidth.
const gzipThreshold = 1450

// This defines the maximum number of parameters per route.
const maxParams = 16

// Push options describes the headers that are sent
// to our server to retrieve the push response.
var pushOptions = http.PushOptions{
	Method: "GET",
	Header: http.Header{
		acceptEncodingHeader: []string{"gzip"},
	},
}

// Context represents a single request & response.
type Context struct {
	// A pointer to the application this request occurred on.
	App *Application

	// Status code
	StatusCode int

	// Custom data
	// TODO: Find a cleaner solution to deal with this?
	Data interface{}

	// net/http
	request  *http.Request
	response http.ResponseWriter

	// Parameters
	paramNames  [maxParams]string
	paramValues [maxParams]string

	// User session
	session *session.Session
}

// Request returns the HTTP request.
func (ctx *Context) Request() Request {
	return Request{
		inner: ctx.request,
	}
}

// Response returns the HTTP response.
func (ctx *Context) Response() Response {
	return Response{
		inner: ctx.response,
	}
}

// Session returns the session of the context or creates and caches a new session.
func (ctx *Context) Session() *session.Session {
	// Return cached session if available.
	if ctx.session != nil {
		return ctx.session
	}

	// Check if the client has a session cookie already.
	cookie, err := ctx.request.Cookie("sid")

	if err == nil {
		sid := cookie.Value

		if session.IsValidID(sid) {
			ctx.session, err = ctx.App.Sessions.Store.Get(sid)

			if err != nil {
				color.Red(err.Error())
			}

			if ctx.session != nil {
				return ctx.session
			}
		}
	}

	// Create a new session
	ctx.session = ctx.App.Sessions.New()

	// Create a session cookie in the client
	ctx.createSessionCookie()

	return ctx.session
}

// HasSession indicates whether the client has a valid session or not.
func (ctx *Context) HasSession() bool {
	if ctx.session != nil {
		return true
	}

	cookie, err := ctx.request.Cookie("sid")

	if err != nil || !session.IsValidID(cookie.Value) {
		return false
	}

	ctx.session, err = ctx.App.Sessions.Store.Get(cookie.Value)

	if err != nil {
		return false
	}

	return ctx.session != nil
}

// createSessionCookie creates a session cookie in the client.
func (ctx *Context) createSessionCookie() {
	sessionCookie := http.Cookie{
		Name:     "sid",
		Value:    ctx.session.ID(),
		HttpOnly: true,
		Secure:   true,
		MaxAge:   ctx.App.Sessions.Duration,
		Path:     "/",
	}

	http.SetCookie(ctx.response, &sessionCookie)
}

// JSON encodes the object to a JSON string and responds.
func (ctx *Context) JSON(value interface{}) error {
	ctx.response.Header().Set(contentTypeHeader, contentTypeJSON)
	bytes, err := jsoniter.Marshal(value)

	if err != nil {
		return err
	}

	return ctx.Bytes(bytes)
}

// JSONLinkedData encodes the object to a JSON linked data string and responds.
func (ctx *Context) JSONLinkedData(value interface{}) error {
	ctx.response.Header().Set(contentTypeHeader, contentTypeJSONLD)
	bytes, err := jsoniter.Marshal(value)

	if err != nil {
		return err
	}

	return ctx.Bytes(bytes)
}

// HTML sends a HTML string.
func (ctx *Context) HTML(html string) error {
	header := ctx.response.Header()
	header.Set(contentTypeHeader, contentTypeHTML)
	header.Set(contentTypeOptionsHeader, contentTypeOptions)
	header.Set(xssProtectionHeader, xssProtection)
	header.Set(referrerPolicyHeader, referrerPolicySameOrigin)

	if ctx.App.Security.Certificate != "" {
		header.Set(strictTransportSecurityHeader, strictTransportSecurity)
		header.Set(contentSecurityPolicyHeader, ctx.App.ContentSecurityPolicy.String())
	}

	if len(ctx.App.Config.Push) > 0 {
		ctx.pushResources()
	}

	return ctx.String(html)
}

// Text sends a plain text string.
func (ctx *Context) Text(text string) error {
	// Don't set text/plain as a content type here
	// because it's actually faster to let net/http do it.
	return ctx.String(text)
}

// CSS sends a style sheet.
func (ctx *Context) CSS(text string) error {
	ctx.response.Header().Set(contentTypeHeader, contentTypeCSS)
	return ctx.String(text)
}

// JavaScript sends a script.
func (ctx *Context) JavaScript(code string) error {
	ctx.response.Header().Set(contentTypeHeader, contentTypeJavaScript)
	return ctx.String(code)
}

// EventStream sends server events to the client.
func (ctx *Context) EventStream(stream *EventStream) error {
	defer close(stream.Closed)

	// Flush
	flusher, ok := ctx.response.(http.Flusher)

	if !ok {
		return ctx.Error(http.StatusNotImplemented, "Flushing not supported")
	}

	// Catch disconnect events
	disconnectedContext := ctx.request.Context()
	disconnectedContext, cancel := context.WithDeadline(disconnectedContext, time.Now().Add(2*time.Hour))
	disconnected := disconnectedContext.Done()
	defer cancel()

	// Send headers
	header := ctx.response.Header()
	header.Set(contentTypeHeader, contentTypeEventStream)
	header.Set(cacheControlHeader, "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("Access-Control-Allow-Origin", "*")
	ctx.response.WriteHeader(200)

	for {
		select {
		case <-disconnected:
			return nil

		case event := <-stream.Events:
			if event != nil {
				data := event.Data

				switch data.(type) {
				case string, []byte:
					// Do nothing with the data if it's already a string or byte slice.
				default:
					var err error
					data, err = jsoniter.Marshal(data)

					if err != nil {
						color.Red("Failed encoding event data as JSON: %v", data)
					}
				}

				fmt.Fprintf(ctx.response, "event: %s\ndata: %s\n\n", event.Name, data)
				flusher.Flush()
			}
		}
	}
}

// File sends the contents of a local file and determines its mime type by extension.
func (ctx *Context) File(file string) error {
	extension := filepath.Ext(file)
	contentType := mime.TypeByExtension(extension)

	// Cache control header
	if isMedia(contentType) {
		ctx.response.Header().Set(cacheControlHeader, cacheControlMedia)
	}

	http.ServeFile(ctx.response, ctx.request, file)
	return nil
}

// ReadAll returns the contents of the reader.
// This will create an in-memory copy and calculate the E-Tag before sending the data.
// Compression will be applied if necessary.
func (ctx *Context) ReadAll(reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		return err
	}

	return ctx.Bytes(data)
}

// Reader sends the contents of the io.Reader without creating an in-memory copy.
// E-Tags will not be generated for the content and compression will not be applied.
// Use this function if your reader contains huge amounts of data.
func (ctx *Context) Reader(reader io.Reader) error {
	_, err := io.Copy(ctx.response, reader)
	return err
}

// ReadSeeker sends the contents of the io.ReadSeeker without creating an in-memory copy.
// E-Tags will not be generated for the content and compression will not be applied.
// Use this function if your reader contains huge amounts of data.
func (ctx *Context) ReadSeeker(reader io.ReadSeeker) error {
	http.ServeContent(ctx.response, ctx.request, "", time.Time{}, reader)
	return nil
}

// Error should be used for sending error messages to the client.
func (ctx *Context) Error(statusCode int, errorList ...interface{}) error {
	ctx.StatusCode = statusCode

	if len(errorList) == 0 {
		message := http.StatusText(statusCode)
		_ = ctx.String(message)
		return errors.New(message)
	}

	messageBuffer := strings.Builder{}

	for index, param := range errorList {
		switch err := param.(type) {
		case string:
			messageBuffer.WriteString(err)
		case error:
			messageBuffer.WriteString(err.Error())
		default:
			continue
		}

		if index != len(errorList)-1 {
			messageBuffer.WriteString(": ")
		}
	}

	message := messageBuffer.String()
	_ = ctx.String(message)
	return errors.New(message)
}

// URI returns the relative path, e.g. /blog/post/123.
func (ctx *Context) URI() string {
	return ctx.request.URL.Path
}

// SetURI sets the relative path, e.g. /blog/post/123.
func (ctx *Context) SetURI(b string) {
	ctx.request.URL.Path = b
}

// Get retrieves an URL parameter.
func (ctx *Context) Get(param string) string {
	// TODO: ...
	return ""
	// return strings.TrimPrefix(ctx.params.ByName(param), "/")
}

// GetInt retrieves an URL parameter as an integer.
func (ctx *Context) GetInt(param string) (int, error) {
	return strconv.Atoi(ctx.Get(param))
}

// RealIP tries to determine the real IP address of the request.
func (ctx *Context) RealIP() string {
	return strings.Trim(realIP(ctx.request), "[]")
}

// UserAgent retrieves the user agent for the given request.
func (ctx *Context) UserAgent() string {
	return ctx.request.UserAgent()
}

// Query retrieves the value for the given URL query parameter.
func (ctx *Context) Query(param string) string {
	return ctx.request.URL.Query().Get(param)
}

// Redirect redirects to the given URL using status code 302.
func (ctx *Context) Redirect(url string) error {
	ctx.StatusCode = http.StatusFound
	ctx.response.Header().Set("Location", url)
	ctx.response.WriteHeader(ctx.StatusCode)
	return nil
}

// RedirectPermanently redirects to the given URL and indicates that this is a permanent change using status code 301.
func (ctx *Context) RedirectPermanently(url string) error {
	ctx.StatusCode = http.StatusMovedPermanently
	ctx.response.Header().Set("Location", url)
	ctx.response.WriteHeader(ctx.StatusCode)
	return nil
}

// isMedia returns whether the given content type is a media type.
func isMedia(contentType string) bool {
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return true
	case strings.HasPrefix(contentType, "video/"):
		return true
	case strings.HasPrefix(contentType, "audio/"):
		return true
	default:
		return false
	}
}

// canCompress returns whether the given content type should be compressed via gzip.
func canCompress(contentType string) bool {
	switch {
	case strings.HasPrefix(contentType, "image/") && contentType != contentTypeSVG:
		return false
	case strings.HasPrefix(contentType, "video/"):
		return false
	case strings.HasPrefix(contentType, "audio/"):
		return false
	default:
		return true
	}
}

// pushResources will push the given resources to the HTTP response.
func (ctx *Context) pushResources() {
	// Check if all the conditions for a push are met
	for _, pushCondition := range ctx.App.pushConditions {
		if !pushCondition(ctx) {
			return
		}
	}

	// Check if we can push
	pusher, ok := ctx.response.(http.Pusher)

	if !ok {
		return
	}

	// OnPush callbacks
	for _, callback := range ctx.App.onPush {
		callback(ctx)
	}

	// Push every resource defined in config.json
	for _, resource := range ctx.App.Config.Push {
		if err := pusher.Push(resource, &pushOptions); err != nil {
			color.Red("Failed to push %s: %v", resource, err)
		}
	}
}

// String responds either with raw text or gzipped if the
// text length is greater than the gzip threshold.
func (ctx *Context) String(body string) error {
	return ctx.Bytes(unsafe.StringToBytes(body))
}

// Bytes responds either with raw text or gzipped if the
// text length is greater than the gzip threshold. Requires a byte slice.
func (ctx *Context) Bytes(body []byte) error {
	// If the request has been canceled by the client, stop.
	if ctx.request.Context().Err() != nil {
		return errors.New("Request interrupted by the client")
	}

	// Small response
	if len(body) < gzipThreshold {
		ctx.response.WriteHeader(ctx.StatusCode)
		_, err := ctx.response.Write(body)
		return err
	}

	// ETag generation
	etag := ETag(body)

	// If client cache is up to date, send 304 with no response body.
	clientETag := ctx.request.Header.Get(ifNoneMatchHeader)

	if etag == clientETag {
		ctx.response.WriteHeader(304)
		return nil
	}

	// Set ETag
	header := ctx.response.Header()
	header.Set(etagHeader, etag)

	// Content type
	contentType := header.Get(contentTypeHeader)
	isMediaType := isMedia(contentType)

	// Cache control header
	if isMediaType {
		header.Set(cacheControlHeader, cacheControlMedia)
	} else {
		header.Set(cacheControlHeader, cacheControlAlwaysValidate)
	}

	// No GZip?
	clientSupportsGZip := strings.Contains(ctx.request.Header.Get(acceptEncodingHeader), "gzip")

	if !ctx.App.Config.GZip || !clientSupportsGZip || !canCompress(contentType) {
		header.Set(contentLengthHeader, strconv.Itoa(len(body)))
		ctx.response.WriteHeader(ctx.StatusCode)
		_, err := ctx.response.Write(body)
		return err
	}

	// GZip
	header.Set(contentEncodingHeader, contentEncodingGzip)
	ctx.response.WriteHeader(ctx.StatusCode)

	// Write response body
	writer := ctx.App.acquireGZipWriter(ctx.response)
	_, err := writer.Write(body)
	writer.Close()

	// Put the writer back into the pool
	ctx.App.gzipWriterPool.Put(writer)

	// Return the error value of the last Write call
	return err
}
