package aero

import "net/http"

// RewriteContext is used for the URI rewrite ability.
type RewriteContext struct {
	request *http.Request
}

// URI returns the relative path, e.g. /blog/post/123.
func (ctx *RewriteContext) URI() string {
	return ctx.request.URL.Path
}

// SetURI sets the relative path, e.g. /blog/post/123.
func (ctx *RewriteContext) SetURI(b string) {
	ctx.request.URL.Path = b
}
