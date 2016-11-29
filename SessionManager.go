package aero

// SessionManager ...
type SessionManager struct {
	Store SessionStore
}

// NewSession creates a new session.
func (manager *SessionManager) New() *Session {
	session := &Session{
		id:   RandomBytes(32),
		data: make(map[string]interface{}),
	}

	manager.Store.Set(BytesToStringUnsafe(session.id), session)

	return session
}
