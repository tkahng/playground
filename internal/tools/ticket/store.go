package ticket

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
)

// Storer is the interface implemented by both the in-memory Store and the
// database-backed DbStore. The interface is intentionally context-free to
// avoid breaking existing call sites; DbStore uses context.Background()
// internally since tickets are short-lived (≤60 s) auth tokens.
type Storer interface {
	Issue(userID, resourceID uuid.UUID) string
	Validate(ticket string) (userID, resourceID uuid.UUID, ok bool)
	Revoke(ticket string)
}

// DbStore is a PostgreSQL-backed ticket store. Unlike the in-memory Store it
// survives server restarts, making it safe for multi-instance deployments.
type DbStore struct {
	db  database.Dbx
	ttl time.Duration
}

var _ Storer = (*DbStore)(nil)

func NewDbStore(db database.Dbx, ttl time.Duration) *DbStore {
	return &DbStore{db: db, ttl: ttl}
}

func (s *DbStore) Issue(userID, resourceID uuid.UUID) string {
	ctx := context.Background()
	s.purge(ctx)
	id := uuid.New().String()
	_, _ = s.db.Exec(ctx, `
		INSERT INTO app.sse_tickets (id, user_id, resource_id, expires_at)
		VALUES ($1, $2, $3, $4)
	`, id, userID, resourceID, time.Now().Add(s.ttl))
	return id
}

func (s *DbStore) Validate(ticket string) (uuid.UUID, uuid.UUID, bool) {
	ctx := context.Background()
	row := s.db.QueryRow(ctx, `
		SELECT user_id, resource_id
		FROM app.sse_tickets
		WHERE id = $1 AND expires_at > clock_timestamp()
	`, ticket)
	var userID, resourceID uuid.UUID
	if err := row.Scan(&userID, &resourceID); err != nil {
		return uuid.Nil, uuid.Nil, false
	}
	return userID, resourceID, true
}

func (s *DbStore) Revoke(ticket string) {
	ctx := context.Background()
	_, _ = s.db.Exec(ctx, `DELETE FROM app.sse_tickets WHERE id = $1`, ticket)
}

func (s *DbStore) purge(ctx context.Context) {
	_, _ = s.db.Exec(ctx, `DELETE FROM app.sse_tickets WHERE expires_at <= clock_timestamp()`)
}

type entry struct {
	UserID     uuid.UUID
	ResourceID uuid.UUID // team-member ID or player ID depending on the SSE endpoint
	expiresAt  time.Time
}

var _ Storer = (*Store)(nil)

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
