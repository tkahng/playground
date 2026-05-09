package ticket_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/tools/ticket"
)

func TestStore_ValidTicketAccepted(t *testing.T) {
	s := ticket.New(time.Minute)
	userID := uuid.New()
	memberID := uuid.New()

	tok := s.Issue(userID, memberID)
	gotUser, gotMember, ok := s.Validate(tok)

	assert.True(t, ok)
	assert.Equal(t, userID, gotUser)
	assert.Equal(t, memberID, gotMember)
}

func TestStore_UnknownTicketRejected(t *testing.T) {
	s := ticket.New(time.Minute)
	_, _, ok := s.Validate("nonexistent-ticket")
	assert.False(t, ok)
}

func TestStore_ExpiredTicketRejected(t *testing.T) {
	s := ticket.New(5 * time.Millisecond)
	tok := s.Issue(uuid.New(), uuid.New())
	time.Sleep(20 * time.Millisecond)
	_, _, ok := s.Validate(tok)
	assert.False(t, ok)
}

func TestStore_RevokedTicketRejected(t *testing.T) {
	s := ticket.New(time.Minute)
	tok := s.Issue(uuid.New(), uuid.New())
	s.Revoke(tok)
	_, _, ok := s.Validate(tok)
	assert.False(t, ok)
}

func TestStore_TicketReusableWithinTTL(t *testing.T) {
	s := ticket.New(time.Minute)
	userID := uuid.New()
	memberID := uuid.New()
	tok := s.Issue(userID, memberID)

	for range 3 {
		_, _, ok := s.Validate(tok)
		assert.True(t, ok, "ticket should remain valid within TTL")
	}
}

func TestStore_DifferentTicketsAreIndependent(t *testing.T) {
	s := ticket.New(time.Minute)
	user1, member1 := uuid.New(), uuid.New()
	user2, member2 := uuid.New(), uuid.New()

	tok1 := s.Issue(user1, member1)
	tok2 := s.Issue(user2, member2)

	u1, m1, ok1 := s.Validate(tok1)
	u2, m2, ok2 := s.Validate(tok2)

	assert.True(t, ok1)
	assert.Equal(t, user1, u1)
	assert.Equal(t, member1, m1)

	assert.True(t, ok2)
	assert.Equal(t, user2, u2)
	assert.Equal(t, member2, m2)
}
