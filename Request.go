package aero

import (
	"net/http"
	"net/url"
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

// Method returns the request method.
func (request Request) Method() string {
	return request.inner.Method
}

// Protocol returns the request protocol.
func (request Request) Protocol() string {
	return request.inner.Proto
}

// Host returns the requested host.
func (request Request) Host() string {
	return request.inner.Host
}

// URL returns the request URL.
func (request Request) URL() *url.URL {
	return request.inner.URL
}
