package aero

import (
	"net/http"
)

// Request represents the HTTP request used in the given context.
type Request struct {
	inner *http.Request
}

// Body represents the request body.
func (request Request) Body() BodyReader {
	return BodyReader{
		reader: request.inner.Body,
	}
}

// Header represents the request headers.
func (request Request) Header() http.Header {
	return request.inner.Header
}
