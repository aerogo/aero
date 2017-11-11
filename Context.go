package aero

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/OneOfOne/xxhash"
	"github.com/aerogo/session"
	"github.com/fatih/color"
	"github.com/julienschmidt/httprouter"
	cache "github.com/patrickmn/go-cache"
	"github.com/tomasen/realip"
	"github.com/valyala/fasthttp"
)

// This should be close to the MTU size of a TCP packet.
// Regarding performance it makes no sense to compress smaller files.
// Bandwidth can be saved however the savings are minimal for small files
// and the overhead of compressing can lead up to a 75% reduction
// in server speed under high load. Therefore in this case
// we're trying to optimize for performance, not bandwidth.
const gzipThreshold = 1450

const (
	serverHeader                  = "Server"
	server                        = "Aero"
	cacheControlHeader            = "Cache-Control"
	cacheControlAlwaysValidate    = "must-revalidate"
	cacheControlMedia             = "public, max-age=864000"
	contentTypeOptionsHeader      = "X-Content-Type-Options"
	contentTypeOptions            = "nosniff"
	xssProtectionHeader           = "X-XSS-Protection"
	xssProtection                 = "1; mode=block"
	etagHeader                    = "ETag"
	contentTypeHeader             = "Content-Type"
	contentTypeHTML               = "text/html; charset=utf-8"
	contentTypeJavaScript         = "application/javascript; charset=utf-8"
	contentTypeJSON               = "application/json; charset=utf-8"
	contentTypePlainText          = "text/plain; charset=utf-8"
	contentEncodingHeader         = "Content-Encoding"
	contentEncodingGzip           = "gzip"
	contentLengthHeader           = "Content-Length"
	responseTimeHeader            = "X-Response-Time"
	ifNoneMatchHeader             = "If-None-Match"
	xFrameOptionsHeader           = "X-Frame-Options"
	xFrameOptions                 = "SAMEORIGIN"
	referrerPolicyHeader          = "Referrer-Policy"
	referrerPolicySameOrigin      = "no-referrer"
	strictTransportSecurityHeader = "Strict-Transport-Security"
	strictTransportSecurity       = "max-age=31536000; includeSubDomains; preload"
	contentSecurityPolicyHeader   = "Content-Security-Policy"
)

// Context ...
type Context struct {
	// net/http
	request  *http.Request
	response http.ResponseWriter
	params   httprouter.Params

	// A pointer to the application this request occured on.
	App *Application

	// Status code
	StatusCode int

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

	// HACK: Add SameSite attribute
	// Remove this once it's available inside http.Cookie
	// cookieData := ctx.response.Header().Get("Set-Cookie")
	// cookieData += "; SameSite=lax"
	// ctx.response.Header().Set("Set-Cookie", cookieData)
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

	bytes, err := json.Marshal(value)

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

// JavaScript sends a script.
func (ctx *Context) JavaScript(code string) string {
	ctx.response.Header().Set(contentTypeHeader, contentTypeJavaScript)
	return code
}

// File sends the contents of a local file and determines its mime type by extension.
func (ctx *Context) File(file string) string {
	extension := filepath.Ext(file)
	mimeType := mime.TypeByExtension(extension)
	data, _ := ioutil.ReadFile(file)

	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	ctx.response.Header().Set(contentTypeHeader, mimeType)
	return string(data)
}

// TryWebP tries to serve a WebP image but will fall back to the specified extension if needed.
func (ctx *Context) TryWebP(path string, extension string) string {
	if ctx.CanUseWebP() {
		extension = ".webp"
	}

	return ctx.File(path + extension)
}

// Error should be used for sending error messages to the user.
func (ctx *Context) Error(statusCode int, explanation string, err error) string {
	ctx.StatusCode = statusCode
	ctx.response.Header().Set(contentTypeHeader, contentTypeHTML)

	if err != nil {
		detailed := err.Error()
		color.Red(detailed)
		return fmt.Sprintf("%s (%s)", explanation, detailed)
	}

	return explanation
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
	return ctx.params.ByName(param)
}

// GetInt retrieves an URL parameter as an integer.
func (ctx *Context) GetInt(param string) (int, error) {
	return strconv.Atoi(ctx.Get(param))
}

// RealIP tries to determine the real IP address of the request.
func (ctx *Context) RealIP() string {
	return realip.RealIP(ctx.request)
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
	ctx.StatusCode = http.StatusPermanentRedirect
	ctx.response.Header().Set("Location", url)
	return ""
}

// CanUseWebP checks the Accept header to find out if WebP is supported by the client's browser.
func (ctx *Context) CanUseWebP() bool {
	accept := ctx.request.Header.Get("Accept")

	if strings.Index(accept, "image/webp") != -1 {
		return true
	}

	return false
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
		if err := pusher.Push(resource, nil); err != nil {
			log.Printf("Failed to push %s: %v", resource, err)
		}
	}
}

// respond responds either with raw code or gzipped if the
// code length is greater than the gzip threshold.
func (ctx *Context) respond(code string) {
	ctx.respondBytes(StringToBytesUnsafe(code))
}

// respondBytes responds either with raw code or gzipped if the
// code length is greater than the gzip threshold. Requires a byte slice.
func (ctx *Context) respondBytes(b []byte) {
	response := ctx.response
	header := response.Header()
	contentType := header.Get(contentTypeHeader)
	isMedia := IsMediaType(contentType)

	// Push
	if contentType == contentTypeHTML {
		header.Set(serverHeader, server)
		ctx.pushResources()
	}

	// Cache control header
	if isMedia {
		header.Set(cacheControlHeader, cacheControlMedia)
	} else {
		header.Set(cacheControlHeader, cacheControlAlwaysValidate)
	}

	// Small response
	if len(b) < gzipThreshold {
		header.Set(contentLengthHeader, strconv.Itoa(len(b)))
		response.WriteHeader(ctx.StatusCode)
		response.Write(b)
		return
	}

	// ETag generation
	h := xxhash.NewS64(0)
	h.Write(b)
	etag := strconv.FormatUint(h.Sum64(), 16)

	// If client cache is up to date, send 304 with no response body.
	clientETag := ctx.request.Header.Get(ifNoneMatchHeader)

	if etag == clientETag {
		response.WriteHeader(304)
		return
	}

	// Set ETag
	header.Set(etagHeader, etag)

	// No GZip?
	if !ctx.App.Config.GZip || isMedia {
		header.Set(contentLengthHeader, strconv.Itoa(len(b)))
		response.WriteHeader(ctx.StatusCode)
		response.Write(b)
		return
	}

	// GZip
	header.Set(contentEncodingHeader, contentEncodingGzip)

	if ctx.App.Config.GZipCache {
		cachedResponse, found := ctx.App.gzipCache.Get(etag)

		if found {
			cachedResponseBytes := cachedResponse.([]byte)
			header.Set(contentLengthHeader, strconv.Itoa(len(cachedResponseBytes)))
			response.WriteHeader(ctx.StatusCode)
			response.Write(cachedResponseBytes)
			return
		}
	}

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	fasthttp.WriteGzipLevel(writer, b, 9)
	writer.Flush()
	gzippedBytes := buffer.Bytes()

	header.Set(contentLengthHeader, strconv.Itoa(len(gzippedBytes)))
	response.WriteHeader(ctx.StatusCode)
	response.Write(gzippedBytes)

	if ctx.App.Config.GZipCache {
		ctx.App.gzipCache.Set(etag, gzippedBytes, cache.DefaultExpiration)
	}
}
