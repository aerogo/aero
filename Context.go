package aero

import (
	stdContext "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aerogo/session"
	"github.com/akyoto/color"
	"github.com/akyoto/stringutils/unsafe"
)

const (
	// gzipThreshold should be close to the size of a TCP packet.
	// Regarding performance it makes no sense to compress smaller files.
	// This value used to be around 1.4 KB, however with the rise
	// of mobile connections that have tiny MTUs even small savings
	// provide a reasonable benefit and therefore we're setting this
	// to 256 which reduces the final packet size by roughly 70 bytes
	// in the smallest case.
	gzipThreshold = 256

	// maxParams defines the maximum number of parameters per route.
	maxParams = 16

	// maxModifiers defines the maximum number of modifiers per context.
	maxModifiers = 4
)

// Context represents the interface for a request & response context.
type Context interface {
	AddModifier(Modifier)
	App() *Application
	Bytes([]byte) error
	Close()
	CSS(string) error
	Get(string) string
	GetInt(string) (int, error)
	Error(int, ...interface{}) error
	EventStream(stream *EventStream) error
	File(string) error
	HasSession() bool
	HTML(string) error
	IP() string
	JavaScript(string) error
	JSON(interface{}) error
	Path() string
	Query(param string) string
	ReadAll(io.Reader) error
	Reader(io.Reader) error
	ReadSeeker(io.ReadSeeker) error
	Redirect(status int, url string) error
	RemoteIP() string
	Request() Request
	Response() Response
	Session() *session.Session
	SetStatus(int)
	Status() int
	String(string) error
	Text(string) error
}

// context represents a request & response context.
type context struct {
	app           *Application
	status        int
	request       request
	response      response
	session       *session.Session
	handler       Handler
	paramNames    [maxParams]string
	paramValues   [maxParams]string
	paramCount    int
	modifiers     [maxModifiers]Modifier
	modifierCount int
}

// AddModifier adds a modifier that can change the response body
// contents of in-memory responses before the actual response happens.
func (ctx *context) AddModifier(modifier Modifier) {
	ctx.modifiers[ctx.modifierCount] = modifier
	ctx.modifierCount++
}

// App returns the application the context occurred in.
func (ctx *context) App() *Application {
	return ctx.app
}

// Bytes responds either with raw text or gzipped if the
// text length is greater than the gzip threshold. Requires a byte slice.
func (ctx *context) Bytes(body []byte) error {
	// If the request has been canceled by the client, stop.
	if ctx.request.Context().Err() != nil {
		return ErrRequestInterruptedByClient
	}

	// If we registered any response body modifiers, invoke them.
	if ctx.modifierCount > 0 {
		for i := 0; i < ctx.modifierCount; i++ {
			body = ctx.modifiers[i](body)
		}
	}

	// Small response
	if len(body) < gzipThreshold {
		ctx.response.inner.WriteHeader(ctx.status)
		_, err := ctx.response.inner.Write(body)
		return err
	}

	// ETag generation
	etag := ETag(body)

	// If client cache is up to date, send 304 with no response body.
	clientETag := ctx.request.Header(ifNoneMatchHeader)

	if etag == clientETag {
		ctx.response.inner.WriteHeader(304)
		return nil
	}

	// Set ETag
	header := ctx.response.inner.Header()
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
	clientSupportsGZip := strings.Contains(ctx.request.Header(acceptEncodingHeader), "gzip")

	if !ctx.app.Config.GZip || !clientSupportsGZip || !canCompress(contentType) {
		header.Set(contentLengthHeader, strconv.Itoa(len(body)))
		ctx.response.inner.WriteHeader(ctx.status)
		_, err := ctx.response.inner.Write(body)
		return err
	}

	// GZip
	header.Set(contentEncodingHeader, contentEncodingGzip)
	ctx.response.inner.WriteHeader(ctx.status)

	// Write response body
	writer := ctx.app.acquireGZipWriter(ctx.response.inner)
	_, err := writer.Write(body)
	writer.Close()

	// Put the writer back into the pool
	ctx.app.gzipWriterPool.Put(writer)

	// Return the error value of the last Write call
	return err
}

// Close frees up resources and is automatically called
// in the ServeHTTP part of the web server.
func (ctx *context) Close() {
	ctx.app.contextPool.Put(ctx)
}

// CSS sends a style sheet.
func (ctx *context) CSS(text string) error {
	ctx.response.SetHeader(contentTypeHeader, contentTypeCSS)
	return ctx.String(text)
}

// Error should be used for sending error messages to the client.
func (ctx *context) Error(statusCode int, errorList ...interface{}) error {
	ctx.status = statusCode

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

// EventStream sends server events to the client.
func (ctx *context) EventStream(stream *EventStream) error {
	defer close(stream.Closed)

	// Flush
	flusher, ok := ctx.response.inner.(http.Flusher)

	if !ok {
		return ctx.Error(http.StatusNotImplemented, "Flushing not supported")
	}

	// Catch disconnect events
	disconnectedContext := ctx.request.Context()
	disconnectedContext, cancel := stdContext.WithDeadline(disconnectedContext, time.Now().Add(2*time.Hour))
	disconnected := disconnectedContext.Done()
	defer cancel()

	// Send headers
	header := ctx.response.inner.Header()
	header.Set(contentTypeHeader, contentTypeEventStream)
	header.Set(cacheControlHeader, "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("Access-Control-Allow-Origin", "*")
	ctx.response.inner.WriteHeader(200)

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
					data, err = json.Marshal(data)

					if err != nil {
						color.Red("Failed encoding event data as JSON: %v", data)
					}
				}

				fmt.Fprintf(ctx.response.inner, "event: %s\ndata: %s\n\n", event.Name, data)
				flusher.Flush()
			}
		}
	}
}

// File sends the contents of a local file and determines its mime type by extension.
func (ctx *context) File(file string) error {
	extension := filepath.Ext(file)
	contentType := mime.TypeByExtension(extension)

	// Cache control header
	if isMedia(contentType) {
		ctx.response.SetHeader(cacheControlHeader, cacheControlMedia)
	}

	http.ServeFile(ctx.response.inner, ctx.request.inner, file)
	return nil
}

// Get retrieves an URL parameter.
func (ctx *context) Get(param string) string {
	for i := 0; i < ctx.paramCount; i++ {
		if ctx.paramNames[i] == param {
			return ctx.paramValues[i]
		}
	}

	return ""
}

// GetInt retrieves an URL parameter as an integer.
func (ctx *context) GetInt(param string) (int, error) {
	return strconv.Atoi(ctx.Get(param))
}

// HasSession indicates whether the client has a valid session or not.
func (ctx *context) HasSession() bool {
	if ctx.session != nil {
		return true
	}

	cookie, err := ctx.request.inner.Cookie("sid")

	if err != nil || !session.IsValidID(cookie.Value) {
		return false
	}

	ctx.session, err = ctx.app.Sessions.Store.Get(cookie.Value)

	if err != nil {
		return false
	}

	return ctx.session != nil
}

// HTML sends a HTML string.
func (ctx *context) HTML(html string) error {
	header := ctx.response.inner.Header()
	header.Set(contentTypeHeader, contentTypeHTML)
	header.Set(contentTypeOptionsHeader, contentTypeOptions)
	header.Set(xssProtectionHeader, xssProtection)
	header.Set(referrerPolicyHeader, referrerPolicySameOrigin)

	if ctx.app.Security.Certificate != "" {
		header.Set(strictTransportSecurityHeader, strictTransportSecurity)
		header.Set(contentSecurityPolicyHeader, ctx.app.ContentSecurityPolicy.String())
	}

	if len(ctx.app.Config.Push) > 0 {
		err := ctx.pushResources()

		if err != nil {
			for _, callback := range ctx.app.onError {
				callback(ctx, err)
			}
		}
	}

	return ctx.String(html)
}

// IP tries to determine the real IP address of the client.
func (ctx *context) IP() string {
	return strings.Trim(realIP(ctx.request.inner), "[]")
}

// JavaScript sends a script.
func (ctx *context) JavaScript(code string) error {
	ctx.response.SetHeader(contentTypeHeader, contentTypeJavaScript)
	return ctx.String(code)
}

// JSON encodes the object to a JSON string and responds.
func (ctx *context) JSON(value interface{}) error {
	ctx.response.SetHeader(contentTypeHeader, contentTypeJSON)
	bytes, err := json.Marshal(value)

	if err != nil {
		return err
	}

	return ctx.Bytes(bytes)
}

// Path returns the relative request path, e.g. /blog/post/123.
func (ctx *context) Path() string {
	return ctx.request.inner.URL.Path
}

// ReadAll returns the contents of the reader.
// This will create an in-memory copy and calculate the E-Tag before sending the data.
// Compression will be applied if necessary.
func (ctx *context) ReadAll(reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)

	if err != nil {
		return err
	}

	return ctx.Bytes(data)
}

// Reader sends the contents of the io.Reader without creating an in-memory copy.
// E-Tags will not be generated for the content and compression will not be applied.
// Use this function if your reader contains huge amounts of data.
func (ctx *context) Reader(reader io.Reader) error {
	_, err := io.Copy(ctx.response.inner, reader)
	return err
}

// ReadSeeker sends the contents of the io.ReadSeeker without creating an in-memory copy.
// E-Tags will not be generated for the content and compression will not be applied.
// Use this function if your reader contains huge amounts of data.
func (ctx *context) ReadSeeker(reader io.ReadSeeker) error {
	http.ServeContent(ctx.response.inner, ctx.request.inner, "", time.Time{}, reader)
	return nil
}

// Redirect redirects to the given URL.
func (ctx *context) Redirect(status int, url string) error {
	ctx.status = status
	ctx.response.SetHeader("Location", url)
	ctx.response.inner.WriteHeader(ctx.status)
	return nil
}

// RemoteIP returns the remote IP address. This will return
// the IP address of the endpoint (e.g. a proxy) but not
// necessarily the IP of the client.
func (ctx *context) RemoteIP() string {
	remoteIP := ctx.request.inner.RemoteAddr

	// If there is a colon in the remote address,
	// remove the port number.
	if strings.ContainsRune(remoteIP, ':') {
		remoteIP, _, _ = net.SplitHostPort(remoteIP)
	}

	return remoteIP
}

// Request returns the HTTP request.
func (ctx *context) Request() Request {
	return &ctx.request
}

// Response returns the HTTP response.
func (ctx *context) Response() Response {
	return &ctx.response
}

// Session returns the session of the context or creates and caches a new session.
func (ctx *context) Session() *session.Session {
	// Return cached session if available.
	if ctx.session != nil {
		return ctx.session
	}

	// Check if the client has a session cookie already.
	cookie, err := ctx.request.inner.Cookie("sid")

	if err == nil {
		sid := cookie.Value

		if session.IsValidID(sid) {
			ctx.session, err = ctx.app.Sessions.Store.Get(sid)

			if err != nil {
				color.Red(err.Error())
			}

			if ctx.session != nil {
				return ctx.session
			}
		}
	}

	// Create a new session
	ctx.session = ctx.app.Sessions.New()
	http.SetCookie(ctx.response.inner, ctx.app.Sessions.Cookie(ctx.session))
	return ctx.session
}

// SetPath sets the relative request path, e.g. /blog/post/123.
func (ctx *context) SetPath(path string) {
	ctx.request.inner.URL.Path = path
}

// SetStatus sets the HTTP status.
func (ctx *context) SetStatus(status int) {
	ctx.status = status
}

// Status returns the HTTP status.
func (ctx *context) Status() int {
	return ctx.status
}

// String responds either with raw text or gzipped if the
// text length is greater than the gzip threshold.
func (ctx *context) String(body string) error {
	return ctx.Bytes(unsafe.StringToBytes(body))
}

// Text sends a plain text string.
func (ctx *context) Text(text string) error {
	ctx.response.SetHeader(contentTypeHeader, contentTypePlainText)
	return ctx.String(text)
}

// Query retrieves the value for the given URL query parameter.
func (ctx *context) Query(param string) string {
	return ctx.request.inner.URL.Query().Get(param)
}

// addParameter adds a new parameter to the context.
func (ctx *context) addParameter(name string, value string) {
	ctx.paramNames[ctx.paramCount] = name
	ctx.paramValues[ctx.paramCount] = value
	ctx.paramCount++
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

// push will start pushing the given resources in a separate goroutine.
func (ctx *context) push(paths ...string) error {
	// Check if we can push
	pusher, ok := ctx.response.inner.(http.Pusher)

	if !ok {
		return nil
	}

	// OnPush callbacks
	for _, callback := range ctx.app.onPush {
		callback(ctx)
	}

	// Push every resource
	for _, path := range paths {
		err := pusher.Push(path, &ctx.app.pushOptions)

		if err != nil {
			return err
		}
	}

	return nil
}

// pushResources will start pushing the given resources
// in a separate goroutine if the defined conditions are true.
func (ctx *context) pushResources() error {
	for _, pushCondition := range ctx.app.pushConditions {
		if !pushCondition(ctx) {
			return nil
		}
	}

	return ctx.push(ctx.app.Config.Push...)
}
