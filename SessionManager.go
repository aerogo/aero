package aero

// SessionManager ...
type SessionManager struct {
	Store SessionStore

	// Automatically create sessions on every request.
	AutoCreate bool
}

// NewSession creates a new session.
func (manager *SessionManager) NewSession() *Session {
	session := &Session{
		id:   RandomBytes(32),
		data: make(map[string]interface{}),
	}

	manager.Store.Set(BytesToStringUnsafe(session.id), session)

	return session
}
