package aero

import (
	"bytes"
	"compress/gzip"
	"strconv"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
)

const (
	gzipThreshold         = 1450
	responseCacheTime     = 30 * time.Second
	contentEncodingHeader = "Content-Encoding"
	contentEncoding       = "gzip"
	contentTypeHeader     = "Content-Type"
	contentType           = "text/html;charset=utf-8"
	etagHeader            = "ETag"
	serverHeader          = "Server"
	server                = "Aero"
	ifNoneMatchHeader     = "If-None-Match: "
	ifNoneMatchOffset     = len("If-None-Match: ")
)

var ifNoneMatchHeaderBytes []byte
var etagToResponse *cache.Cache

func init() {
	ifNoneMatchHeaderBytes = []byte(ifNoneMatchHeader)
	etagToResponse = cache.New(responseCacheTime, responseCacheTime)
}

func sendCachedResponse(ctx *fasthttp.RequestCtx, etag string) bool {
	// Headers
	ctx.Response.Header.Set(contentTypeHeader, contentType)
	ctx.Response.Header.Set(serverHeader, server)

	// Is the client cache up to date?
	headers := ctx.Request.Header.Header()
	index := bytes.Index(headers, ifNoneMatchHeaderBytes)

	if index != -1 {
		var clientETag []byte
		for i := index; i < len(headers); i++ {
			if headers[i] == '\r' {
				clientETag = headers[index+ifNoneMatchOffset : i]

				// Send short 304 response if the ETags match
				// if bytes.Compare([]byte(etag), clientETag) == 0 {
				if etag == string(clientETag) {
					ctx.SetStatusCode(304)
					return true
				}

				return false
			}
		}
	}

	return false
}

// Respond responds either with raw code or gzipped if the
// code length is greater than the gzip threshold.
func Respond(ctx *fasthttp.RequestCtx, code string) {
	RespondBytes(ctx, []byte(code))
}

// RespondBytes responds either with raw code or gzipped if the
// code length is greater than the gzip threshold. Requires a byte slice.
func RespondBytes(ctx *fasthttp.RequestCtx, b []byte) {
	// ETag generation
	h := xxhash.NewS64(0)
	h.Write(b)
	etag := strconv.FormatUint(h.Sum64(), 10)
	ctx.Response.Header.Set(etagHeader, etag)

	if sendCachedResponse(ctx, etag) {
		return
	}

	// Body
	if len(b) >= gzipThreshold {
		ctx.Response.Header.Set(contentEncodingHeader, contentEncoding)

		cachedResponse, found := etagToResponse.Get(etag)

		if false && found {
			ctx.Write(cachedResponse.([]byte))
		} else {
			// TODO: This needs optimization by reusing gzip writers
			var buffer bytes.Buffer
			gz, _ := gzip.NewWriterLevel(&buffer, 1)
			gz.Write(b)
			gz.Flush()
			defer gz.Close()
			response := buffer.Bytes()
			etagToResponse.Set(etag, response, cache.DefaultExpiration)
			ctx.Write(response)
			// fasthttp.WriteGzipLevel(ctx.Response.BodyWriter(), b, 1)
		}
	} else {
		ctx.Write(b)
	}
}
