package services

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// mustRematchCompletedGame creates two players and a completed RPS game between them.
func mustRematchCompletedGame(t testing.TB, ctx context.Context, adapter stores.StorageAdapterInterface) *RpsGameWithParticipants {
	t.Helper()
	svc := NewDbRpsGameService(adapter, nil)
	p1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming())
	p2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming())
	game, err := svc.RequestGame(ctx, &RpsGameRequestInput{
		RequestingPlayerID:   p1.ID,
		InvitedPlayerID:      p2.ID,
		RequestingPlayerMove: models.RpsParticipantMoveRock,
		DurationSeconds:      3600,
	})
	require.NoError(t, err)
	completed, err := svc.RespondToGameRequest(ctx, &GameRequestResponse{
		InvitedPlayerID: p2.ID,
		GameID:          game.RpsGame.ID,
		Status:          models.RpsGameStatusCompleted,
		Move:            models.RpsParticipantMoveScissors,
	})
	require.NoError(t, err)
	return completed
}

func TestDbRpsGameService_RequestRematch_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)

		game := mustRematchCompletedGame(t, ctx, adapter)

		rematch, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)
		assert.Equal(t, game.RpsGame.ID, rematch.OriginalGameID)
		assert.Equal(t, models.RpsRematchStatusPending, rematch.Status)
		assert.True(t, rematch.ExpiresAt.After(time.Now()))
		assert.Nil(t, rematch.NewGameID)
	})
}

func TestDbRpsGameService_RequestRematch_RejectsNonCompletedGame(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		p1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("host2@rematch.test"))
		p2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("guest2@rematch.test"))
		game, err := svc.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   p1.ID,
			InvitedPlayerID:      p2.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      3600,
		})
		require.NoError(t, err)

		_, err = svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: p1.ID,
			InvitedPlayerID:    p2.ID,
		})
		require.Error(t, err)
	})
}

func TestDbRpsGameService_RequestRematch_RejectsDuplicate(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		_, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)

		// Second request for same game should conflict.
		_, err = svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.Error(t, err)
	})
}

func TestDbRpsGameService_AcceptRematch_CreatesNewGameWithSwappedRoles(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		rematch, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)

		accepted, err := svc.AcceptRematch(ctx, &RematchAcceptInput{
			RematchID:       rematch.ID,
			InvitedPlayerID: game.InvitedParticipant.PlayerID,
			HostMove:        models.RpsParticipantMovePaper,
		})
		require.NoError(t, err)
		assert.Equal(t, models.RpsRematchStatusAccepted, accepted.Status)
		require.NotNil(t, accepted.NewGameID)

		// Verify new game exists and roles are swapped.
		newGame, err := svc.FindRpsGameWithParticipants(ctx, *accepted.NewGameID)
		require.NoError(t, err)
		require.NotNil(t, newGame)
		assert.Equal(t, models.RpsGameStatusPending, newGame.RpsGame.Status)
		// Previous guest is now host — with status completed (move submitted during accept).
		assert.Equal(t, game.InvitedParticipant.PlayerID, newGame.RequestingParticipant.PlayerID)
		assert.Equal(t, models.RpsParticipantTypeHost, newGame.RequestingParticipant.Type)
		assert.Equal(t, models.RpsParticipantStatusCompleted, newGame.RequestingParticipant.Status)
		assert.Equal(t, models.RpsParticipantMovePaper, newGame.RequestingParticipant.Move)
		// Previous host is now guest — still pending.
		assert.Equal(t, game.RequestingParticipant.PlayerID, newGame.InvitedParticipant.PlayerID)
		assert.Equal(t, models.RpsParticipantTypeGuest, newGame.InvitedParticipant.Type)
		assert.Equal(t, models.RpsParticipantStatusPending, newGame.InvitedParticipant.Status)
	})
}

func TestDbRpsGameService_AcceptRematch_RejectsWrongPlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		rematch, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)

		// Host tries to accept their own request — must be rejected.
		_, err = svc.AcceptRematch(ctx, &RematchAcceptInput{
			RematchID:       rematch.ID,
			InvitedPlayerID: game.RequestingParticipant.PlayerID,
			HostMove:        models.RpsParticipantMoveRock,
		})
		require.Error(t, err)
	})
}

func TestDbRpsGameService_AcceptRematch_RejectsExpired(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		rematch, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)

		// Force-expire the rematch by writing past the TTL.
		rematch.ExpiresAt = time.Now().UTC().Add(-1 * time.Second)
		_, err = adapter.Gaming().UpdateRpsRematchRequest(context.Background(), rematch)
		require.NoError(t, err)

		_, err = svc.AcceptRematch(ctx, &RematchAcceptInput{
			RematchID:       rematch.ID,
			InvitedPlayerID: game.InvitedParticipant.PlayerID,
			HostMove:        models.RpsParticipantMoveRock,
		})
		require.Error(t, err)
	})
}

func TestDbRpsGameService_DeclineRematch_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		rematch, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)

		declined, err := svc.DeclineRematch(ctx, rematch.ID, game.InvitedParticipant.PlayerID)
		require.NoError(t, err)
		assert.Equal(t, models.RpsRematchStatusDeclined, declined.Status)
		assert.Nil(t, declined.NewGameID)
	})
}

func TestDbRpsGameService_DeclineRematch_RejectsWrongPlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		rematch, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)

		_, err = svc.DeclineRematch(ctx, rematch.ID, game.RequestingParticipant.PlayerID)
		require.Error(t, err)
	})
}

func TestDbRpsGameService_DeclineRematch_RejectsAlreadyActioned(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		rematch, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)

		_, err = svc.DeclineRematch(ctx, rematch.ID, game.InvitedParticipant.PlayerID)
		require.NoError(t, err)

		// Declining again must fail.
		_, err = svc.DeclineRematch(ctx, rematch.ID, game.InvitedParticipant.PlayerID)
		require.Error(t, err)
	})
}

func TestDbRpsGameService_ExpireRematches(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		// Create a request that is already past its TTL.
		req, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)
		req.ExpiresAt = time.Now().UTC().Add(-1 * time.Second)
		_, err = adapter.Gaming().UpdateRpsRematchRequest(ctx, req)
		require.NoError(t, err)

		count, err := svc.ExpireRematches(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify status updated in DB.
		stored, err := adapter.Gaming().FindRpsRematchRequest(ctx, &stores.RpsRematchFilter{
			Ids: []uuid.UUID{req.ID},
		})
		require.NoError(t, err)
		require.NotNil(t, stored)
		assert.Equal(t, models.RpsRematchStatusExpired, stored.Status)
	})
}

func TestDbRpsGameService_ExpireRematches_SkipsNonExpired(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		// Create a still-valid request.
		_, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)

		count, err := svc.ExpireRematches(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

// TestDbRpsGameService_RequestRematch_ConcurrentDuplicate verifies that when two
// goroutines simultaneously call RequestRematch for the same game, exactly one
// succeeds and one gets a Conflict error.
func TestDbRpsGameService_RequestRematch_ConcurrentDuplicate(t *testing.T) {
	database.WithNewDatabase(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		const numGoroutines = 2
		type result struct{ err error }
		results := make(chan result, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := svc.RequestRematch(ctx, &RematchRequestInput{
					OriginalGameID:     game.RpsGame.ID,
					RequestingPlayerID: game.RequestingParticipant.PlayerID,
					InvitedPlayerID:    game.InvitedParticipant.PlayerID,
				})
				results <- result{err: err}
			}()
		}
		wg.Wait()
		close(results)

		var successes, failures int
		for r := range results {
			if r.err == nil {
				successes++
			} else {
				failures++
			}
		}
		if successes != 1 {
			t.Errorf("successes = %d, want exactly 1", successes)
		}
		if failures != 1 {
			t.Errorf("failures = %d, want exactly 1", failures)
		}

		// Exactly one pending rematch must exist.
		pending, err := adapter.Gaming().FindRpsRematchRequest(ctx, &stores.RpsRematchFilter{
			OriginalGameIDs: []uuid.UUID{game.RpsGame.ID},
			Statuses:        []models.RpsRematchStatus{models.RpsRematchStatusPending},
		})
		require.NoError(t, err)
		assert.NotNil(t, pending, "exactly one pending rematch must exist")
	})
}

// TestDbRpsGameService_AcceptRematch_Atomic verifies that a failure during
// participant creation does not leave an orphaned game row. The decorator
// intercepts CreateRpsParticipants; the transaction wrapping AcceptRematch
// must roll back the game row together with the failed participant write.
func TestDbRpsGameService_AcceptRematch_Atomic(t *testing.T) {
	database.WithNewDatabase(t, func(ctx context.Context, db database.Dbx) {
		// Use a gaming decorator so we can inject a fault.
		gamingDec := stores.NewDbGamingStoreDecorator(db)
		adapter := stores.NewDbAdapterDecorators(db)
		adapter.GamingFunc = gamingDec

		svc := NewDbRpsGameService(adapter, nil)
		game := mustRematchCompletedGame(t, ctx, adapter)

		rematch, err := svc.RequestRematch(ctx, &RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: game.RequestingParticipant.PlayerID,
			InvitedPlayerID:    game.InvitedParticipant.PlayerID,
		})
		require.NoError(t, err)

		before, err := adapter.Gaming().CountRpsGames(ctx, &stores.RpsGameFilter{})
		require.NoError(t, err)

		// Inject failure: participant creation fails after game is created.
		gamingDec.CreateRpsParticipantsFunc = func(_ context.Context, _ []*models.RpsParticipant) ([]*models.RpsParticipant, error) {
			return nil, fmt.Errorf("injected participant error")
		}

		_, err = svc.AcceptRematch(ctx, &RematchAcceptInput{
			RematchID:       rematch.ID,
			InvitedPlayerID: game.InvitedParticipant.PlayerID,
			HostMove:        models.RpsParticipantMoveRock,
		})
		require.Error(t, err, "AcceptRematch must return error when participant creation fails")

		gamingDec.CreateRpsParticipantsFunc = nil

		// Game count must be unchanged — the new game row was rolled back.
		after, err := adapter.Gaming().CountRpsGames(ctx, &stores.RpsGameFilter{})
		require.NoError(t, err)
		assert.Equal(t, before, after, "no orphaned game row should exist after partial failure")

		// Rematch status must still be pending.
		updated, err := adapter.Gaming().FindRpsRematchRequest(ctx, &stores.RpsRematchFilter{
			Ids: []uuid.UUID{rematch.ID},
		})
		require.NoError(t, err)
		assert.Equal(t, models.RpsRematchStatusPending, updated.Status)
	})
}
