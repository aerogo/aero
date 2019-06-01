package aero

import "net/http/httptest"

// Handler is a function that deals with the given request/response context.
type Handler func(Context) error

// Simulate simulates the request with the given handler
// and context and returns the recorded response.
func (handler Handler) Simulate(ctx Context) (*httptest.ResponseRecorder, error) {
	// Set up fake state
	originalResponse := ctx.Response().Internal()
	request := ctx.Request().Internal()
	originalAcceptEncoding := request.Header.Get("Accept-Encoding")
	request.Header.Set("Accept-Encoding", "")

	// Record the response
	response := httptest.NewRecorder()
	ctx.Response().SetInternal(response)
	err := handler(ctx)

	// Restore old state
	request.Header.Set("Accept-Encoding", originalAcceptEncoding)
	ctx.Response().SetInternal(originalResponse)

	return response, err
}
