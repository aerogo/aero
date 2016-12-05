package aero

// SessionManager ...
type SessionManager struct {
	Store SessionStore
}

// New creates a new session.
func (manager *SessionManager) New() *Session {
	session := &Session{
		id:   string(RandomBytes(32)),
		data: make(map[string]interface{}),
	}

	manager.Store.Set(session.id, session)

	return session
}
