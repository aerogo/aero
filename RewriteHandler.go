package aero

import "net/http"

type rewriteHandler struct {
	rewrite func(*RewriteContext)
	router  http.Handler
}

func (r *rewriteHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	r.rewrite(&RewriteContext{
		request: request,
	})
	r.router.ServeHTTP(response, request)
}
