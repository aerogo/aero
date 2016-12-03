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
	serverHeader               = "Server"
	server                     = "Aero"
	cacheControlHeader         = "Cache-Control"
	cacheControlAlwaysValidate = "must-revalidate"
	contentTypeOptionsHeader   = "X-Content-Type-Options"
	contentTypeOptions         = "nosniff"
	xssProtectionHeader        = "X-XSS-Protection"
	xssProtection              = "1; mode=block"
	etagHeader                 = "ETag"
	contentTypeHeader          = "Content-Type"
	contentTypeHTML            = "text/html; charset=utf-8"
	contentTypeJSON            = "application/json"
	contentTypePlainText       = "text/plain; charset=utf-8"
	contentEncodingHeader      = "Content-Encoding"
	contentEncodingGzip        = "gzip"
	responseTimeHeader         = "X-Response-Time"
	ifNoneMatchHeader          = "If-None-Match"
)

var ifNoneMatchHeaderBytes []byte

func init() {
	ifNoneMatchHeaderBytes = []byte(ifNoneMatchHeader)
}

// Context ...
type Context struct {
	// net/http
	request  *http.Request
	response http.ResponseWriter
	params   httprouter.Params

	// A pointer to the application this request occured on.
	App *Application

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
		Value:    BytesToStringUnsafe(ctx.session.id),
		HttpOnly: true,
		Secure:   true,
	}

	http.SetCookie(ctx.response, &sessionCookie)

	return ctx.session
}

// JSON encodes the object to a JSON string and responds.
func (ctx *Context) JSON(value interface{}) string {
	bytes, _ := json.Marshal(value)

	ctx.response.Header().Set(contentTypeHeader, contentTypeJSON)
	return string(bytes)
}

// HTML sends a HTML string.
func (ctx *Context) HTML(html string) string {
	ctx.response.Header().Set(contentTypeHeader, contentTypeHTML)
	return html
}

// Text sends a plain text string.
func (ctx *Context) Text(text string) string {
	ctx.response.Header().Set(contentTypeHeader, contentTypePlainText)
	return text
}

// Error should be used for sending error messages to the user.
func (ctx *Context) Error(statusCode int, explanation string, err error) string {
	ctx.SetStatusCode(statusCode)
	ctx.response.Header().Set(contentTypeHeader, contentTypeHTML)

	// fmt.Println("{")
	// color.Blue("\t" + ctx.requestCtx.Request.URI().String())
	// color.Yellow("\t" + explanation)
	// color.Red("\t" + err.Error())
	// fmt.Println("}")

	return explanation
}

// SetStatusCode sets the status code of the request.
func (ctx *Context) SetStatusCode(status int) {
	ctx.response.WriteHeader(status)
}

// SetHeader sets header to value.
func (ctx *Context) SetHeader(header string, value string) {
	ctx.response.Header().Set(header, value)
}

// Get retrieves an URL parameter.
func (ctx *Context) Get(param string) string {
	return ctx.params.ByName(param)
}

// GetInt retrieves an URL parameter as an integer.
func (ctx *Context) GetInt(param string) (int, error) {
	return strconv.Atoi(ctx.Get(param))
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

	// Headers
	response.Header().Set(cacheControlHeader, cacheControlAlwaysValidate)
	response.Header().Set(serverHeader, server)
	response.Header().Set(contentTypeOptionsHeader, contentTypeOptions)
	response.Header().Set(xssProtectionHeader, xssProtection)
	// response.Header().Set(responseTimeHeader, strconv.FormatInt(time.Since(ctx.start).Nanoseconds()/1000, 10)+" us")

	if ctx.App.Security.Certificate != "" {
		response.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		response.Header().Set("Content-Security-Policy", "default-src 'none'; img-src https:; script-src 'self'; style-src 'sha256-"+ctx.App.cssHash+"'; font-src https:; frame-src https:; connect-src https: wss:")
	}

	response.Header().Set("X-Frame-Options", "SAMEORIGIN")

	// Body
	if ctx.App.Config.GZip && len(b) >= gzipThreshold {
		response.Header().Set(contentEncodingHeader, contentEncodingGzip)

		// ETag generation
		h := xxhash.NewS64(0)
		h.Write(b)
		etag := strconv.FormatUint(h.Sum64(), 16)

		// If client cache is up to date, send 304 with no response body.
		clientETag := ctx.request.Header.Get(ifNoneMatchHeader)

		if etag == clientETag {
			ctx.SetStatusCode(304)
			return
		}

		// Set ETag
		response.Header().Set(etagHeader, etag)

		if ctx.App.Config.GZipCache {
			cachedResponse, found := ctx.App.gzipCache.Get(etag)

			if found {
				response.Write(cachedResponse.([]byte))
				return
			}
		}

		var buffer bytes.Buffer
		writer := bufio.NewWriter(&buffer)
		fasthttp.WriteGzipLevel(writer, b, 1)

		if ctx.App.Config.GZipCache {
			writer.Flush()
			ctx.App.gzipCache.Set(etag, buffer.Bytes(), cache.DefaultExpiration)
		}
	} else {
		response.Write(b)
	}
}
