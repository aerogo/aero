package aero

import (
	"net/http"
)

// Response is the interface for an HTTP response.
type Response interface {
	Header(string) string
	SetHeader(string, string)
}

// response represents the HTTP response used in the given context.
type response struct {
	inner http.ResponseWriter
}

// Header returns the header value for the given key.
func (res *response) Header(key string) string {
	return res.inner.Header().Get(key)
}

// SetHeader sets the header value for the given key.
func (res *response) SetHeader(key string, value string) {
	res.inner.Header().Set(key, value)
}
