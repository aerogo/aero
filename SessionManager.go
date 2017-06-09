package aero

// SessionManager ...
type SessionManager struct {
	Store SessionStore
}

// New creates a new session.
func (manager *SessionManager) New() *Session {
	sessionID := GenerateUUID()

	// Session data is not allowed to be empty.
	// Therefore we are adding the session ID as dummy data.
	sessionData := map[string]interface{}{
		"sid": sessionID,
	}

	session := NewSession(sessionID, sessionData)
	manager.Store.Set(session.id, session)
	return session
}
