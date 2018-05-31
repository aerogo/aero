package aero

import (
	"net/http"
)

// Response represents the HTTP response used in the given context.
type Response struct {
	inner http.ResponseWriter
}

// Header represents the response headers.
func (response Response) Header() http.Header {
	return response.inner.Header()
}
