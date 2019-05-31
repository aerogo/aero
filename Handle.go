package aero

// Handle is a function that deals with the given request/response context.
type Handle = func(Context) error
