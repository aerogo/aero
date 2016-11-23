package aero

// SessionStore ...
type SessionStore interface {
	Get(string) *Session
	Set(string, *Session)
}
