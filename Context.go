package aero

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aerogo/session"
	"github.com/blitzprog/color"
	jsoniter "github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"
	"github.com/tomasen/realip"
)

// This should be close to the MTU size of a TCP packet.
// Regarding performance it makes no sense to compress smaller files.
// Bandwidth can be saved however the savings are minimal for small files
// and the overhead of compressing can lead up to a 75% reduction
// in server speed under high load. Therefore in this case
// we're trying to optimize for performance, not bandwidth.
const gzipThreshold = 1450

const (
	cacheControlHeader            = "Cache-Control"
	cacheControlAlwaysValidate    = "must-revalidate"
	cacheControlMedia             = "public, max-age=13824000"
	contentTypeOptionsHeader      = "X-Content-Type-Options"
	contentTypeOptions            = "nosniff"
	xssProtectionHeader           = "X-XSS-Protection"
	xssProtection                 = "1; mode=block"
	etagHeader                    = "ETag"
	contentTypeHeader             = "Content-Type"
	contentTypeHTML               = "text/html; charset=utf-8"
	contentTypeCSS                = "text/css; charset=utf-8"
	contentTypeJavaScript         = "application/javascript; charset=utf-8"
	contentTypeJSON               = "application/json; charset=utf-8"
	contentTypeJSONLD             = "application/ld+json; charset=utf-8"
	contentTypePlainText          = "text/plain; charset=utf-8"
	contentTypeEventStream        = "text/event-stream; charset=utf-8"
	contentEncodingHeader         = "Content-Encoding"
	contentEncodingGzip           = "gzip"
	acceptEncodingHeader          = "Accept-Encoding"
	contentLengthHeader           = "Content-Length"
	ifNoneMatchHeader             = "If-None-Match"
	referrerPolicyHeader          = "Referrer-Policy"
	referrerPolicySameOrigin      = "no-referrer"
	strictTransportSecurityHeader = "Strict-Transport-Security"
	strictTransportSecurity       = "max-age=31536000; includeSubDomains; preload"
	contentSecurityPolicyHeader   = "Content-Security-Policy"

	// responseTimeHeader            = "X-Response-Time"
	// xFrameOptionsHeader           = "X-Frame-Options"
	// xFrameOptions                 = "SAMEORIGIN"
	// serverHeader                  = "Server"
	// server                        = "Aero"
)

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
	// net/http
	request  *http.Request
	response http.ResponseWriter
	params   httprouter.Params

	// Responded tells if the request has been dealt with already
	responded bool

	// A pointer to the application this request occurred on.
	App *Application

	// Status code
	StatusCode int

	// Error message
	ErrorMessage string

	// Custom data
	Data interface{}

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

// JSON encodes the object to a JSON string and responds.
func (ctx *Context) JSON(value interface{}) string {
	ctx.response.Header().Set(contentTypeHeader, contentTypeJSON)

	bytes, err := jsoniter.Marshal(value)

	if err != nil {
		ctx.StatusCode = http.StatusInternalServerError
		return `{"error": "Could not encode object to JSON"}`
	}

	return string(bytes)
}

// JSONLinkedData encodes the object to a JSON linked data string and responds.
func (ctx *Context) JSONLinkedData(value interface{}) string {
	ctx.response.Header().Set(contentTypeHeader, contentTypeJSONLD)

	bytes, err := jsoniter.Marshal(value)

	if err != nil {
		ctx.StatusCode = http.StatusInternalServerError
		return `{"error": "Could not encode object to JSON"}`
	}

	return string(bytes)
}

// HTML sends a HTML string.
func (ctx *Context) HTML(html string) string {
	ctx.response.Header().Set(contentTypeHeader, contentTypeHTML)
	ctx.response.Header().Set(contentTypeOptionsHeader, contentTypeOptions)
	ctx.response.Header().Set(xssProtectionHeader, xssProtection)
	// ctx.response.Header().Set(xFrameOptionsHeader, xFrameOptions)
	ctx.response.Header().Set(referrerPolicyHeader, referrerPolicySameOrigin)

	if ctx.App.Security.Certificate != "" {
		ctx.response.Header().Set(strictTransportSecurityHeader, strictTransportSecurity)
		ctx.response.Header().Set(contentSecurityPolicyHeader, ctx.App.ContentSecurityPolicy.String())
	}

	return html
}

// Text sends a plain text string.
func (ctx *Context) Text(text string) string {
	ctx.response.Header().Set(contentTypeHeader, contentTypePlainText)
	return text
}

// CSS sends a style sheet.
func (ctx *Context) CSS(text string) string {
	ctx.response.Header().Set(contentTypeHeader, contentTypeCSS)
	return text
}

// JavaScript sends a script.
func (ctx *Context) JavaScript(code string) string {
	ctx.response.Header().Set(contentTypeHeader, contentTypeJavaScript)
	return code
}

// EventStream sends server events to the client.
func (ctx *Context) EventStream(stream *EventStream) string {
	defer close(stream.Closed)

	// Flush
	flusher, ok := ctx.response.(http.Flusher)

	if !ok {
		return ctx.Error(http.StatusNotImplemented, "Flushing not supported")
	}

	// Catch disconnect events
	disconnected := ctx.request.Context().Done()

	// Send headers
	header := ctx.response.Header()
	header.Set(contentTypeHeader, contentTypeEventStream)
	header.Set(cacheControlHeader, "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("Access-Control-Allow-Origin", "*")
	ctx.response.WriteHeader(200)
	ctx.responded = true

	for {
		select {
		case <-disconnected:
			return ""

		case event := <-stream.Events:
			if event != nil {
				data := event.Data

				switch data.(type) {
				case string, []byte:
					// Do nothing with the data if it's already a string or byte slice.
				default:
					data, _ = jsoniter.Marshal(data)
				}

				fmt.Fprintf(ctx.response, "event: %s\ndata: %s\n\n", event.Name, data)
				flusher.Flush()
			}

		case <-time.After(5 * time.Second):
			// Send one byte to keep alive the connection
			// which will also check for disconnection.
			ctx.response.Write([]byte("\n"))
			flusher.Flush()
		}
	}
}

// File sends the contents of a local file and determines its mime type by extension.
func (ctx *Context) File(file string) string {
	extension := filepath.Ext(file)
	contentType := mime.TypeByExtension(extension)

	// Cache control header
	if IsMediaType(contentType) {
		ctx.response.Header().Set(cacheControlHeader, cacheControlMedia)
	}

	http.ServeFile(ctx.response, ctx.request, file)
	ctx.responded = true
	return ""
}

// ReadAll returns the contents of the reader.
// This will create an in-memory copy and calculate the E-Tag before sending the data.
// Compression will be applied if necessary.
func (ctx *Context) ReadAll(reader io.Reader) string {
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		return ctx.Error(http.StatusInternalServerError, err)
	}

	return BytesToStringUnsafe(data)
}

// Reader sends the contents of the io.Reader without creating an in-memory copy.
// E-Tags will not be generated for the content and compression will not be applied.
// Use this function if your reader contains huge amounts of data.
func (ctx *Context) Reader(reader io.Reader) string {
	io.Copy(ctx.response, reader)
	ctx.responded = true
	return ""
}

// ReadSeeker sends the contents of the io.ReadSeeker without creating an in-memory copy.
// E-Tags will not be generated for the content and compression will not be applied.
// Use this function if your reader contains huge amounts of data.
func (ctx *Context) ReadSeeker(reader io.ReadSeeker) string {
	http.ServeContent(ctx.response, ctx.request, "", time.Time{}, reader)
	ctx.responded = true
	return ""
}

// Error should be used for sending error messages to the user.
func (ctx *Context) Error(statusCode int, errors ...interface{}) string {
	ctx.StatusCode = statusCode
	ctx.response.Header().Set(contentTypeHeader, contentTypeHTML)

	message := bytes.Buffer{}

	if len(errors) == 0 {
		message.WriteString(fmt.Sprintf("Unknown error: %d", statusCode))
	} else {
		for index, param := range errors {
			switch err := param.(type) {
			case string:
				message.WriteString(err)
			case error:
				message.WriteString(err.Error())
			default:
				continue
			}

			if index != len(errors)-1 {
				message.WriteString(": ")
			}
		}
	}

	ctx.ErrorMessage = message.String()
	color.Red(ctx.ErrorMessage)
	return ctx.ErrorMessage
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
	return strings.TrimPrefix(ctx.params.ByName(param), "/")
}

// GetInt retrieves an URL parameter as an integer.
func (ctx *Context) GetInt(param string) (int, error) {
	return strconv.Atoi(ctx.Get(param))
}

// RealIP tries to determine the real IP address of the request.
func (ctx *Context) RealIP() string {
	return strings.Trim(realip.RealIP(ctx.request), "[]")
}

// UserAgent retrieves the user agent for the given request.
func (ctx *Context) UserAgent() string {
	ctx.request.URL.Query()
	return ctx.request.UserAgent()
}

// Query retrieves the value for the given URL query parameter.
func (ctx *Context) Query(param string) string {
	return ctx.request.URL.Query().Get(param)
}

// Redirect redirects to the given URL using status code 302.
func (ctx *Context) Redirect(url string) string {
	ctx.StatusCode = http.StatusFound
	ctx.response.Header().Set("Location", url)
	return ""
}

// RedirectPermanently redirects to the given URL and indicates that this is a permanent change using status code 301.
func (ctx *Context) RedirectPermanently(url string) string {
	ctx.StatusCode = http.StatusMovedPermanently
	ctx.response.Header().Set("Location", url)
	return ""
}

// IsMediaType returns whether the given content type is a media type.
func IsMediaType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/") || strings.HasPrefix(contentType, "audio/")
}

// pushResources will push the given resources to the HTTP response.
func (ctx *Context) pushResources() {
	// Check if all the conditions for a push are met
	for _, pushCondition := range ctx.App.pushConditions {
		if !pushCondition(ctx) {
			return
		}
	}

	// OnPush callbacks
	for _, callback := range ctx.App.onPush {
		callback(ctx)
	}

	// Check if we can push
	pusher, ok := ctx.response.(http.Pusher)

	if !ok {
		return
	}

	// Push every resource defined in config.json
	for _, resource := range ctx.App.Config.Push {
		if err := pusher.Push(resource, &pushOptions); err != nil {
			log.Printf("Failed to push %s: %v", resource, err)
		}
	}
}

// respond responds either with raw code or gzipped if the
// code length is greater than the gzip threshold.
func (ctx *Context) respond(code string) {
	// If the request has been dealt with already,
	// or if the request has been canceled by the client,
	// there's nothing to do here.
	if ctx.responded || ctx.request.Context().Err() != nil {
		return
	}

	ctx.respondBytes(StringToBytesUnsafe(code))
}

// respondBytes responds either with raw code or gzipped if the
// code length is greater than the gzip threshold. Requires a byte slice.
func (ctx *Context) respondBytes(b []byte) {
	response := ctx.response
	header := response.Header()
	contentType := header.Get(contentTypeHeader)
	isMedia := IsMediaType(contentType)

	// Cache control header
	if isMedia {
		header.Set(cacheControlHeader, cacheControlMedia)
	} else {
		header.Set(cacheControlHeader, cacheControlAlwaysValidate)
	}

	// Push
	if contentType == contentTypeHTML && len(ctx.App.Config.Push) > 0 {
		ctx.pushResources()
	}

	// Small response
	if len(b) < gzipThreshold {
		header.Set(contentLengthHeader, strconv.Itoa(len(b)))
		response.WriteHeader(ctx.StatusCode)
		response.Write(b)
		return
	}

	// ETag generation
	etag := ETag(b)

	// If client cache is up to date, send 304 with no response body.
	clientETag := ctx.request.Header.Get(ifNoneMatchHeader)

	if etag == clientETag {
		response.WriteHeader(304)
		return
	}

	// Set ETag
	header.Set(etagHeader, etag)

	// No GZip?
	supportsGZip := strings.Contains(ctx.request.Header.Get(acceptEncodingHeader), "gzip")

	if !ctx.App.Config.GZip || !supportsGZip || isMedia {
		header.Set(contentLengthHeader, strconv.Itoa(len(b)))
		response.WriteHeader(ctx.StatusCode)
		response.Write(b)
		return
	}

	// GZip
	header.Set(contentEncodingHeader, contentEncodingGzip)
	response.WriteHeader(ctx.StatusCode)

	// Write response body
	writer, _ := gzip.NewWriterLevel(response, gzip.BestCompression)
	writer.Write(b)
	writer.Flush()
}
