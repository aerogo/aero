package aero

import "net/http"

// rewriteHandler is the replacement handler we use in case we have a Rewrite function.
type rewriteHandler struct {
	rewrite func(*RewriteContext)
	router  http.Handler
}

// ServeHTTP deals with the request in case we have a Rewrite function.
func (r *rewriteHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	r.rewrite(&RewriteContext{
		Request: request,
	})
	r.router.ServeHTTP(response, request)
}
