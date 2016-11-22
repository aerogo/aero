package aero

import "github.com/valyala/fasthttp"

// RewriteContext is used for the URI rewrite ability.
type RewriteContext struct {
	requestCtx *fasthttp.RequestCtx
}

// URIBytes returns the relative path, e.g. /blog/post/123 as a byte slice.
func (ctx *RewriteContext) URIBytes() []byte {
	return ctx.requestCtx.RequestURI()
}

// SetURIBytes returns the relative path, e.g. /blog/post/123 as a byte slice.
func (ctx *RewriteContext) SetURIBytes(b []byte) {
	ctx.requestCtx.Request.SetRequestURIBytes(b)
}
