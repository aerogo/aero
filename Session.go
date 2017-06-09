package aero

import "sync"

// Session ...
type Session struct {
	id       string
	data     map[string]interface{}
	lock     sync.RWMutex
	modified bool
}

// NewSession creates a new session with the given ID and data.
func NewSession(sid string, baseData map[string]interface{}) *Session {
	return &Session{
		id:   sid,
		data: baseData,
	}
}

// ID returns the session ID.
func (session *Session) ID() string {
	return session.id
}

// Get returns the value for the key in this session.
func (session *Session) Get(key string) interface{} {
	session.lock.RLock()
	value := session.data[key]
	session.lock.RUnlock()
	return value
}

// GetString returns the string value for the key in this session.
func (session *Session) GetString(key string) string {
	value := session.Get(key)

	if value != nil {
		str, ok := value.(string)

		if ok {
			return str
		}

		return ""
	}

	return ""
}

// Set sets the value for the key in this session.
func (session *Session) Set(key string, value interface{}) {
	session.lock.Lock()
	session.data[key] = value
	session.lock.Unlock()

	session.modified = true
}

// Modified indicates whether the session has been modified since it's been retrieved.
func (session *Session) Modified() bool {
	return session.modified
}

// Data returns the underlying session data.
// READING OR WRITING DATA IS NOT THREAD-SAFE.
// Use Set() and Get() to modify data safely.
func (session *Session) Data() map[string]interface{} {
	return session.data
}
