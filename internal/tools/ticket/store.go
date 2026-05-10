package ticket

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type entry struct {
	UserID     uuid.UUID
	ResourceID uuid.UUID // team-member ID or player ID depending on the SSE endpoint
	expiresAt  time.Time
}

// Store is a short-lived in-memory ticket store used for SSE authentication.
// Tickets are issued via a normal authenticated POST and then exchanged for an
// SSE connection, avoiding the need to put long-lived JWTs in server-logged URLs.
type Store struct {
	mu      sync.Mutex
	entries map[string]entry
	ttl     time.Duration
}

func New(ttl time.Duration) *Store {
	return &Store{
		entries: make(map[string]entry),
		ttl:     ttl,
	}
}

// Issue creates a new ticket valid for the store's TTL and returns its ID.
// resourceID is the team-member or player ID depending on the SSE channel.
func (s *Store) Issue(userID, resourceID uuid.UUID) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purge()
	id := uuid.New().String()
	s.entries[id] = entry{
		UserID:     userID,
		ResourceID: resourceID,
		expiresAt:  time.Now().Add(s.ttl),
	}
	return id
}

// Validate returns the userID and resourceID associated with a ticket.
// The ticket remains valid (multi-use) until it expires, allowing EventSource
// auto-reconnects within the TTL window.
func (s *Store) Validate(ticket string) (userID, resourceID uuid.UUID, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, exists := s.entries[ticket]
	if !exists {
		return uuid.Nil, uuid.Nil, false
	}
	if time.Now().After(e.expiresAt) {
		delete(s.entries, ticket)
		return uuid.Nil, uuid.Nil, false
	}
	return e.UserID, e.ResourceID, true
}

// Revoke removes a ticket immediately.
func (s *Store) Revoke(ticket string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, ticket)
}

func (s *Store) purge() {
	now := time.Now()
	for k, v := range s.entries {
		if now.After(v.expiresAt) {
			delete(s.entries, k)
		}
	}
}
