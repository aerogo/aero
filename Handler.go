package aero

import "net/http/httptest"

// Handler is a function that deals with the given request/response context.
type Handler func(Context) error

// Simulate simulates the request with the given handler
// and context and returns the recorded response.
func (handler Handler) Simulate(ctx Context) (*httptest.ResponseRecorder, error) {
	c := ctx.(*context)

	// Set up fake state
	originalResponse := c.response
	originalAcceptEncoding := c.request.Header.Get("Accept-Encoding")
	c.request.Header.Set("Accept-Encoding", "")

	// Record the response
	response := httptest.NewRecorder()
	c.response = response
	err := handler(ctx)

	// Restore old state
	c.request.Header.Set("Accept-Encoding", originalAcceptEncoding)
	c.response = originalResponse

	return response, err
}
