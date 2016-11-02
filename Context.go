package aero

import (
	"encoding/json"
	"strconv"
	"time"
	"unsafe"

	"github.com/OneOfOne/xxhash"
	"github.com/buaazp/fasthttprouter"
	cache "github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
)

const (
	gzipThreshold              = 1450
	responseCacheDuration      = 5 * time.Minute
	responseCacheCleanup       = 1 * time.Minute
	contentEncodingHeader      = "Content-Encoding"
	contentEncodingGzip        = "gzip"
	contentTypeHeader          = "Content-Type"
	contentType                = "text/html;charset=utf-8"
	contentTypeJSON            = "application/json"
	contentTypePlainText       = "text/plain;charset=utf-8"
	etagHeader                 = "ETag"
	cacheControlHeader         = "Cache-Control"
	cacheControlAlwaysValidate = "no-cache"
	responseTimeHeader         = "X-Response-Time"
	serverHeader               = "Server"
	server                     = "Aero"
	ifNoneMatchHeader          = "If-None-Match"
)

var ifNoneMatchHeaderBytes []byte
var etagToResponse *cache.Cache

func init() {
	ifNoneMatchHeaderBytes = []byte(ifNoneMatchHeader)
	etagToResponse = cache.New(responseCacheDuration, responseCacheCleanup)
}

// Context ...
type Context struct {
	// Keep this as the first parameter for quick pointer acquisition.
	requestCtx *fasthttp.RequestCtx

	// A pointer to the application this request occured on.
	App *Application

	// Parameters used in this request.
	Params fasthttprouter.Params

	// Start time
	start time.Time
}

// Handle ...
type Handle func(*Context)

// JSON encodes the object to a JSON strings and responds.
func (ctx *Context) JSON(value interface{}) {
	bytes, _ := json.Marshal(value)

	ctx.requestCtx.Response.Header.Set(contentTypeHeader, contentTypeJSON)
	ctx.RespondBytes(bytes)
}

// HTML sends a HTML string.
func (ctx *Context) HTML(html string) {
	ctx.requestCtx.Response.Header.Set(contentTypeHeader, contentType)
	ctx.Respond(html)
}

// Text sends a plain text string.
func (ctx *Context) Text(html string) {
	ctx.requestCtx.Response.Header.Set(contentTypeHeader, contentTypePlainText)
	ctx.Respond(html)
}

// SetHeader sets header to value.
func (ctx *Context) SetHeader(header string, value string) {
	ctx.requestCtx.Response.Header.Set(header, value)
}

// Respond responds either with raw code or gzipped if the
// code length is greater than the gzip threshold.
func (ctx *Context) Respond(code string) {
	ctx.RespondBytes([]byte(code))
}

// RespondBytes responds either with raw code or gzipped if the
// code length is greater than the gzip threshold. Requires a byte slice.
func (ctx *Context) RespondBytes(b []byte) {
	http := ctx.requestCtx

	// ETag generation
	h := xxhash.NewS64(0)
	h.Write(b)
	etag := strconv.FormatUint(h.Sum64(), 16)

	// Headers
	http.Response.Header.Set(etagHeader, etag)
	http.Response.Header.Set(cacheControlHeader, cacheControlAlwaysValidate)
	http.Response.Header.Set(serverHeader, server)
	http.Response.Header.Set(responseTimeHeader, strconv.FormatInt(time.Since(ctx.start).Nanoseconds()/1000, 10)+" us")

	// If client cache is up to date, send 304 with no response body.
	clientETag := http.Request.Header.Peek(ifNoneMatchHeader)

	if etag == *(*string)(unsafe.Pointer(&clientETag)) {
		http.SetStatusCode(304)
		return
	}

	// Body
	if ctx.App.Config.GZip && len(b) >= gzipThreshold {
		http.Response.Header.Set(contentEncodingHeader, contentEncodingGzip)

		if ctx.App.Config.GZipCache {
			cachedResponse, found := etagToResponse.Get(etag)

			if found {
				http.Write(cachedResponse.([]byte))
				return
			}
		}

		fasthttp.WriteGzipLevel(http.Response.BodyWriter(), b, 1)

		if ctx.App.Config.GZipCache {
			body := http.Response.Body()
			gzipped := make([]byte, len(body))
			copy(gzipped, body)
			etagToResponse.Set(etag, gzipped, cache.DefaultExpiration)
		}
	} else {
		http.Write(b)
	}
}
