package aero

// Modifier is a function that modifies the
// response body before it is sent to the client.
type Modifier = func([]byte) []byte
