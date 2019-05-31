package aero

// Middleware is a function that accepts a handler
// and transforms it into a different handler.
type Middleware func(Handler) Handler

// Bind chains the middleware to be called before the handler
// and returns the new handler.
func (m Middleware) Bind(handler Handler) Handler {
	return m(handler)
}
