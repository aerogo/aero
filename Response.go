package aero

import (
	"net/http"
)

// Response is the interface for an HTTP response.
type Response interface {
	Header(string) string
	Internal() http.ResponseWriter
	SetHeader(string, string)
	SetInternal(http.ResponseWriter)
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

// Internal returns the underlying http.ResponseWriter.
// This method should be avoided unless absolutely necessary
// because Aero doesn't guarantee that the underlying framework
// will always stay net/http based in the future.
func (res *response) Internal() http.ResponseWriter {
	return res.inner
}

// SetInternal sets the underlying http.ResponseWriter.
// This method should be avoided unless absolutely necessary
// because Aero doesn't guarantee that the underlying framework
// will always stay net/http based in the future.
func (res *response) SetInternal(writer http.ResponseWriter) {
	res.inner = writer
}
