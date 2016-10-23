package aero

import (
	"strconv"
	"time"
	"unsafe"

	"github.com/OneOfOne/xxhash"
	"github.com/patrickmn/go-cache"
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

// Configuration ...
type Configuration struct {
	GZip      bool
	GZipCache bool
}

// Config ...
var Config Configuration

var ifNoneMatchHeaderBytes []byte
var etagToResponse *cache.Cache

func init() {
	Config.GZip = true
	Config.GZipCache = true

	ifNoneMatchHeaderBytes = []byte(ifNoneMatchHeader)
	etagToResponse = cache.New(responseCacheDuration, responseCacheCleanup)
}

// Respond responds either with raw code or gzipped if the
// code length is greater than the gzip threshold.
func Respond(ctx *fasthttp.RequestCtx, code string) {
	RespondBytes(ctx, *(*[]byte)(unsafe.Pointer(&code)))
}

// RespondBytes responds either with raw code or gzipped if the
// code length is greater than the gzip threshold. Requires a byte slice.
func RespondBytes(ctx *fasthttp.RequestCtx, b []byte) {
	// ETag generation
	h := xxhash.NewS64(0)
	h.Write(b)
	etag := strconv.FormatUint(h.Sum64(), 10)
	ctx.Response.Header.Set(etagHeader, etag)

	// Headers
	ctx.Response.Header.Set(contentTypeHeader, contentType)
	ctx.Response.Header.Set(serverHeader, server)

	// If client cache is up to date, send 304 with no response body
	clientETag := ctx.Request.Header.Peek(ifNoneMatchHeader)

	if etag == *(*string)(unsafe.Pointer(&clientETag)) {
		ctx.SetStatusCode(304)
		return
	}

	// Body
	if Config.GZip && len(b) >= gzipThreshold {
		ctx.Response.Header.Set(contentEncodingHeader, contentEncoding)

		if Config.GZipCache {
			cachedResponse, found := etagToResponse.Get(etag)

			if found {
				ctx.Write(cachedResponse.([]byte))
				return
			}
		}

		fasthttp.WriteGzipLevel(ctx.Response.BodyWriter(), b, 1)
		defer etagToResponse.Set(etag, ctx.Response.Body(), cache.DefaultExpiration)
	} else {
		ctx.Write(b)
	}
}
