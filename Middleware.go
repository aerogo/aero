package aero

// Middleware is a function that accepts a handler
// and transforms it into a different handler.
type Middleware func(Handler) Handler
