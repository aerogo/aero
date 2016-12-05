package aero

import "sync"

// Session ...
type Session struct {
	id   string
	data map[string]interface{}
	lock sync.RWMutex
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
}
