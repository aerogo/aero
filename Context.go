package aero

import (
	"strconv"
	"time"
	"unsafe"

	"github.com/OneOfOne/xxhash"
	"github.com/buaazp/fasthttprouter"
	cache "github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
)

const (
	gzipThreshold         = 1450
	responseCacheDuration = 5 * time.Minute
	responseCacheCleanup  = 1 * time.Minute
	contentEncodingHeader = "Content-Encoding"
	contentEncoding       = "gzip"
	contentTypeHeader     = "Content-Type"
	contentType           = "text/html;charset=utf-8"
	etagHeader            = "ETag"
	serverHeader          = "Server"
	server                = "Aero"
	ifNoneMatchHeader     = "If-None-Match"
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
}

// Handle ...
type Handle func(*Context)

// Respond responds either with raw code or gzipped if the
// code length is greater than the gzip threshold.
func (aeroCtx *Context) Respond(code string) {
	// // Convert string to byte slice
	// stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&code))

	// if stringHeader != nil {
	// 	converted := (*[0x7fffffff]byte)(unsafe.Pointer(stringHeader.Data))[:len(code):len(code)]
	// 	if converted != nil {
	// 		aeroCtx.RespondBytes(converted)
	// 		return
	// 	}
	// }

	aeroCtx.RespondBytes([]byte(code))
}

// RespondBytes responds either with raw code or gzipped if the
// code length is greater than the gzip threshold. Requires a byte slice.
func (aeroCtx *Context) RespondBytes(b []byte) {
	ctx := aeroCtx.requestCtx

	// ETag generation
	h := xxhash.NewS64(0)
	h.Write(b)
	etag := strconv.FormatUint(h.Sum64(), 16)
	ctx.Response.Header.Set(etagHeader, etag)

	// Headers
	ctx.Response.Header.Set(contentTypeHeader, contentType)
	ctx.Response.Header.Set(serverHeader, server)

	// If client cache is up to date, send 304 with no response body.
	clientETag := ctx.Request.Header.Peek(ifNoneMatchHeader)

	if etag == *(*string)(unsafe.Pointer(&clientETag)) {
		ctx.SetStatusCode(304)
		return
	}

	// Body
	if aeroCtx.App.Config.GZip && len(b) >= gzipThreshold {
		ctx.Response.Header.Set(contentEncodingHeader, contentEncoding)

		if aeroCtx.App.Config.GZipCache {
			cachedResponse, found := etagToResponse.Get(etag)

			if found {
				ctx.Write(cachedResponse.([]byte))
				return
			}
		}

		fasthttp.WriteGzipLevel(ctx.Response.BodyWriter(), b, 1)

		if aeroCtx.App.Config.GZipCache {
			etagToResponse.Set(etag, ctx.Response.Body(), cache.DefaultExpiration)
		}
	} else {
		ctx.Write(b)
	}
}
