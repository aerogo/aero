package aero

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/OneOfOne/xxhash"
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
	cacheControlMedia             = "max-age=864000"
	contentTypeOptionsHeader      = "X-Content-Type-Options"
	contentTypeOptions            = "nosniff"
	xssProtectionHeader           = "X-XSS-Protection"
	xssProtection                 = "1; mode=block"
	etagHeader                    = "ETag"
	contentTypeHeader             = "Content-Type"
	contentTypeHTML               = "text/html; charset=utf-8"
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

	// Start time
	start time.Time

	// User session
	session *Session
}

// Handle ...
type Handle func(*Context) string

// Session returns the session of the context or creates and caches a new session.
func (ctx *Context) Session() *Session {
	// Return cached session if available.
	if ctx.session != nil {
		return ctx.session
	}

	// Check if the client has a session cookie already.
	cookie, err := ctx.request.Cookie("sid")

	if err == nil {
		sid := cookie.Value

		if sid != "" {
			ctx.session = ctx.App.Sessions.Store.Get(sid)

			if ctx.session != nil {
				return ctx.session
			}
		}
	}

	// Create a new session
	ctx.session = ctx.App.Sessions.New()

	sessionCookie := http.Cookie{
		Name:     "sid",
		Value:    ctx.session.id,
		HttpOnly: true,
		Secure:   true,
	}

	http.SetCookie(ctx.response, &sessionCookie)

	return ctx.session
}

// HasSession indicates whether a session has been constructed for this context or not.
func (ctx *Context) HasSession() bool {
	return ctx.session != nil
}

// JSON encodes the object to a JSON string and responds.
func (ctx *Context) JSON(value interface{}) string {
	bytes, _ := json.Marshal(value)

	ctx.SetResponseHeader(contentTypeHeader, contentTypeJSON)
	return string(bytes)
}

// HTML sends a HTML string.
func (ctx *Context) HTML(html string) string {
	ctx.SetResponseHeader(contentTypeHeader, contentTypeHTML)
	ctx.SetResponseHeader(contentTypeOptionsHeader, contentTypeOptions)
	ctx.SetResponseHeader(xssProtectionHeader, xssProtection)
	ctx.SetResponseHeader(xFrameOptionsHeader, xFrameOptions)
	ctx.SetResponseHeader(referrerPolicyHeader, referrerPolicySameOrigin)

	if ctx.App.Security.Certificate != "" {
		ctx.SetResponseHeader(strictTransportSecurityHeader, strictTransportSecurity)
		ctx.SetResponseHeader(contentSecurityPolicyHeader, ctx.App.contentSecurityPolicy)
	}

	return html
}

// Text sends a plain text string.
func (ctx *Context) Text(text string) string {
	ctx.SetResponseHeader(contentTypeHeader, contentTypePlainText)
	return text
}

// File sends the contents of a local file and determines its mime type by extension.
func (ctx *Context) File(file string) string {
	extension := filepath.Ext(file)
	mimeType := mime.TypeByExtension(extension)
	data, _ := ioutil.ReadFile(file)

	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	ctx.SetResponseHeader(contentTypeHeader, mimeType)
	return string(data)
}

// Error should be used for sending error messages to the user.
func (ctx *Context) Error(statusCode int, explanation string, err error) string {
	ctx.StatusCode = statusCode
	ctx.SetResponseHeader(contentTypeHeader, contentTypeHTML)
	// ctx.App.Logger.Error(
	// 	color.RedString(explanation),
	// 	zap.String("error", err.Error()),
	// 	zap.String("url", ctx.request.RequestURI),
	// )
	color.Red(err.Error())
	return explanation
}

// GetRequestHeader retrieves the value for the request header.
func (ctx *Context) GetRequestHeader(header string) string {
	return ctx.request.Header.Get(header)
}

// SetResponseHeader sets response header to value.
func (ctx *Context) SetResponseHeader(header string, value string) {
	ctx.response.Header().Set(header, value)
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

// RequestBody returns the request body as a string.
func (ctx *Context) RequestBody() []byte {
	body, err := ioutil.ReadAll(ctx.request.Body)

	if err != nil {
		panic(err)
	}

	return body
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
	ctx.SetResponseHeader("Location", url)
	return ""
}

// RedirectPermanently redirects to the given URL and indicates that this is a permanent change using status code 301.
func (ctx *Context) RedirectPermanently(url string) string {
	ctx.StatusCode = http.StatusPermanentRedirect
	ctx.SetResponseHeader("Location", url)
	return ""
}

// CanUseWebP checks the Accept header to find out if WebP is supported by the client's browser.
func (ctx *Context) CanUseWebP() bool {
	accept := ctx.GetRequestHeader("Accept")

	if strings.Index(accept, "image/webp") != -1 {
		return true
	}

	return false
}

// Respond responds either with raw code or gzipped if the
// code length is greater than the gzip threshold.
func (ctx *Context) Respond(code string) {
	ctx.RespondBytes(StringToBytesUnsafe(code))
}

// RespondBytes responds either with raw code or gzipped if the
// code length is greater than the gzip threshold. Requires a byte slice.
func (ctx *Context) RespondBytes(b []byte) {
	response := ctx.response
	header := response.Header()
	contentType := ctx.response.Header().Get(contentTypeHeader)
	isMedia := strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/")

	// Headers
	if isMedia {
		header.Set(cacheControlHeader, cacheControlMedia)
	} else {
		header.Set(cacheControlHeader, cacheControlAlwaysValidate)
		header.Set(serverHeader, server)
		header.Set(responseTimeHeader, strconv.FormatInt(time.Since(ctx.start).Nanoseconds()/1000, 10)+" us")
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
