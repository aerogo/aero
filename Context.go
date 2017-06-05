package aero

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/OneOfOne/xxhash"
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
	contentTypeOptionsHeader      = "X-Content-Type-Options"
	contentTypeOptions            = "nosniff"
	xssProtectionHeader           = "X-XSS-Protection"
	xssProtection                 = "1; mode=block"
	etagHeader                    = "ETag"
	contentTypeHeader             = "Content-Type"
	contentTypeHTML               = "text/html; charset=utf-8"
	contentTypeJSON               = "application/json"
	contentTypePlainText          = "text/plain; charset=utf-8"
	contentEncodingHeader         = "Content-Encoding"
	contentEncodingGzip           = "gzip"
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
	// sid := ctx.requestCtx.Request.Header.CookieBytes(sidBytes)
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

// JSON encodes the object to a JSON string and responds.
func (ctx *Context) JSON(value interface{}) string {
	bytes, _ := json.Marshal(value)

	ctx.SetHeader(contentTypeHeader, contentTypeJSON)
	return string(bytes)
}

// HTML sends a HTML string.
func (ctx *Context) HTML(html string) string {
	ctx.SetHeader(contentTypeHeader, contentTypeHTML)
	ctx.SetHeader(contentTypeOptionsHeader, contentTypeOptions)
	ctx.SetHeader(xssProtectionHeader, xssProtection)
	ctx.SetHeader(xFrameOptionsHeader, xFrameOptions)
	ctx.SetHeader(referrerPolicyHeader, referrerPolicySameOrigin)

	if ctx.App.Security.Certificate != "" {
		ctx.SetHeader(strictTransportSecurityHeader, strictTransportSecurity)
		ctx.SetHeader(contentSecurityPolicyHeader, ctx.App.contentSecurityPolicy)
	}

	return html
}

// Text sends a plain text string.
func (ctx *Context) Text(text string) string {
	ctx.SetHeader(contentTypeHeader, contentTypePlainText)
	return text
}

// Error should be used for sending error messages to the user.
func (ctx *Context) Error(statusCode int, explanation string, err error) string {
	ctx.StatusCode = statusCode
	ctx.SetHeader(contentTypeHeader, contentTypeHTML)
	// ctx.App.Logger.Error(
	// 	color.RedString(explanation),
	// 	zap.String("error", err.Error()),
	// 	zap.String("url", ctx.request.RequestURI),
	// )
	return explanation
}

// SetHeader sets header to value.
func (ctx *Context) SetHeader(header string, value string) {
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

// RealIP tries to determine the real IP address of the request.
func (ctx *Context) RealIP() string {
	return realip.RealIP(ctx.request)
}

// UserAgent retrieves the user agent for the given request.
func (ctx *Context) UserAgent() string {
	return ctx.request.UserAgent()
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

	// Headers
	header.Set(serverHeader, server)
	header.Set(responseTimeHeader, strconv.FormatInt(time.Since(ctx.start).Nanoseconds()/1000, 10)+" us")
	header.Set(cacheControlHeader, cacheControlAlwaysValidate)

	// Body
	if ctx.App.Config.GZip && len(b) >= gzipThreshold {
		header.Set(contentEncodingHeader, contentEncodingGzip)

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

		if ctx.App.Config.GZipCache {
			cachedResponse, found := ctx.App.gzipCache.Get(etag)

			if found {
				response.WriteHeader(ctx.StatusCode)
				response.Write(cachedResponse.([]byte))
				return
			}
		}

		var buffer bytes.Buffer
		writer := bufio.NewWriter(&buffer)
		fasthttp.WriteGzipLevel(writer, b, 9)
		writer.Flush()
		gzippedBytes := buffer.Bytes()

		response.WriteHeader(ctx.StatusCode)
		response.Write(gzippedBytes)

		if ctx.App.Config.GZipCache {
			ctx.App.gzipCache.Set(etag, gzippedBytes, cache.DefaultExpiration)
		}
	} else {
		response.WriteHeader(ctx.StatusCode)
		response.Write(b)
	}
}
