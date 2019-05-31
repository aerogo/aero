package aero

import (
	stdContext "context"
	"net/http"
)

// Request is an interface for HTTP requests.
type Request interface {
	Body() Body
	Context() stdContext.Context
	Header(string) string
	Method() string
	Protocol() string
	Host() string
	Path() string
}

// request represents the HTTP request used in the given context.
type request struct {
	inner *http.Request
}

// Body represents the request body.
func (req request) Body() Body {
	return Body{
		reader: req.inner.Body,
	}
}

// Context returns the request context.
func (req request) Context() stdContext.Context {
	return req.inner.Context()
}

// Header returns the header value for the given key.
func (req request) Header(key string) string {
	return req.inner.Header.Get(key)
}

// Method returns the request method.
func (req request) Method() string {
	return req.inner.Method
}

// Protocol returns the request protocol.
func (req request) Protocol() string {
	return req.inner.Proto
}

// Host returns the requested host.
func (req request) Host() string {
	return req.inner.Host
}

// Path returns the requested path.
func (req request) Path() string {
	return req.inner.URL.Path
}
