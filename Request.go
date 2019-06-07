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
	Host() string
	Internal() *http.Request
	Method() string
	Path() string
	Protocol() string
	Scheme() string
}

// request represents the HTTP request used in the given context.
type request struct {
	inner *http.Request
}

// Body represents the request body.
func (req *request) Body() Body {
	return Body{
		reader: req.inner.Body,
	}
}

// Context returns the request context.
func (req *request) Context() stdContext.Context {
	return req.inner.Context()
}

// Header returns the header value for the given key.
func (req *request) Header(key string) string {
	return req.inner.Header.Get(key)
}

// Method returns the request method.
func (req *request) Method() string {
	return req.inner.Method
}

// Protocol returns the request protocol.
func (req *request) Protocol() string {
	return req.inner.Proto
}

// Host returns the requested host.
func (req *request) Host() string {
	return req.inner.Host
}

// Path returns the requested path.
func (req *request) Path() string {
	return req.inner.URL.Path
}

// Scheme returns http or https depending on what scheme has been used.
func (req *request) Scheme() string {
	scheme := req.inner.Header.Get("X-Forwarded-Proto")

	if scheme != "" {
		return scheme
	}

	if req.inner.TLS != nil {
		return "https"
	}

	return "http"
}

// Internal returns the underlying *http.Request.
// This method should be avoided unless absolutely necessary
// because Aero doesn't guarantee that the underlying framework
// will always stay net/http based in the future.
func (req *request) Internal() *http.Request {
	return req.inner
}
