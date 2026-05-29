package ticket_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/tools/ticket"
)

func TestDbStore_ValidTicketAccepted(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := ticket.NewDbStore(db, time.Minute)
		userID, memberID := uuid.New(), uuid.New()

		tok := s.Issue(userID, memberID)
		gotUser, gotMember, ok := s.Validate(tok)

		assert.True(t, ok)
		assert.Equal(t, userID, gotUser)
		assert.Equal(t, memberID, gotMember)
	})
}

func TestDbStore_UnknownTicketRejected(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := ticket.NewDbStore(db, time.Minute)
		_, _, ok := s.Validate("nonexistent-ticket")
		assert.False(t, ok)
	})
}

func TestDbStore_RevokedTicketRejected(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := ticket.NewDbStore(db, time.Minute)
		tok := s.Issue(uuid.New(), uuid.New())
		s.Revoke(tok)
		_, _, ok := s.Validate(tok)
		assert.False(t, ok)
	})
}

func TestDbStore_TicketReusableWithinTTL(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := ticket.NewDbStore(db, time.Minute)
		userID, memberID := uuid.New(), uuid.New()
		tok := s.Issue(userID, memberID)

		for range 3 {
			_, _, ok := s.Validate(tok)
			require.True(t, ok, "ticket should remain valid within TTL")
		}
	})
}

func TestDbStore_DifferentTicketsAreIndependent(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := ticket.NewDbStore(db, time.Minute)
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
	})
}
