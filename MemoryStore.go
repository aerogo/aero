package aero

import "sync"

// MemoryStore is the default session store.
// You should use it for prototyping, not for production.
type MemoryStore struct {
	sessions map[string]*Session
	lock     sync.RWMutex
}

// NewMemoryStore creates a session memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*Session),
	}
}

// Get a session by its ID.
func (store *MemoryStore) Get(id string) *Session {
	store.lock.RLock()
	session := store.sessions[id]
	store.lock.RUnlock()
	return session
}

// Set saves a session so it can be retrieved by its ID.
func (store *MemoryStore) Set(id string, session *Session) {
	store.lock.Lock()
	store.sessions[id] = session
	store.lock.Unlock()
}
