package aero

// Middleware is a function that accepts a context and the next function in the call chain.
type Middleware func(*Context, func())
