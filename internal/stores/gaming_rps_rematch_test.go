//go:build integration

package stores

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

func mustCreateCompletedGame(t testing.TB, ctx context.Context, store GamingStore) (*models.RpsGame, *models.Player, *models.Player) {
	t.Helper()
	host := MustCreatePlayer(t, ctx, store)
	guest := MustCreatePlayer(t, ctx, store)
	game := MustCreateRpsGame(t, store, WithRpsGameStatus(models.RpsGameStatusCompleted))
	_, err := store.CreateRpsParticipants(ctx, []*models.RpsParticipant{
		{GameID: game.ID, PlayerID: host.ID, Type: models.RpsParticipantTypeHost, Status: models.RpsParticipantStatusCompleted, Move: models.RpsParticipantMoveRock, Result: models.RpsParticipantResultWin},
		{GameID: game.ID, PlayerID: guest.ID, Type: models.RpsParticipantTypeGuest, Status: models.RpsParticipantStatusCompleted, Move: models.RpsParticipantMoveScissors, Result: models.RpsParticipantResultLose},
	})
	require.NoError(t, err)
	return game, host, guest
}

func TestDBGamingStore_CreateRpsRematchRequest(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDBGamingStore(db)
		game, host, guest := mustCreateCompletedGame(t, ctx, store)

		req := &models.RpsRematchRequest{
			OriginalGameID:     game.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
			Status:             models.RpsRematchStatusPending,
			ExpiresAt:          time.Now().UTC().Add(45 * time.Second),
		}
		created, err := store.CreateRpsRematchRequest(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, game.ID, created.OriginalGameID)
		assert.Equal(t, host.ID, created.RequestingPlayerID)
		assert.Equal(t, guest.ID, created.InvitedPlayerID)
		assert.Equal(t, models.RpsRematchStatusPending, created.Status)
		assert.Nil(t, created.NewGameID)
		assert.NotEmpty(t, created.Metadata)
	})
}

func TestDBGamingStore_FindRpsRematchRequest_ByID(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDBGamingStore(db)
		game, host, guest := mustCreateCompletedGame(t, ctx, store)

		created, err := store.CreateRpsRematchRequest(ctx, &models.RpsRematchRequest{
			OriginalGameID:     game.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
			Status:             models.RpsRematchStatusPending,
			ExpiresAt:          time.Now().UTC().Add(45 * time.Second),
		})
		require.NoError(t, err)

		found, err := store.FindRpsRematchRequest(ctx, &RpsRematchFilter{Ids: []uuid.UUID{created.ID}})
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
	})
}

func TestDBGamingStore_FindRpsRematchRequest_ByOriginalGame(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDBGamingStore(db)
		game, host, guest := mustCreateCompletedGame(t, ctx, store)

		created, err := store.CreateRpsRematchRequest(ctx, &models.RpsRematchRequest{
			OriginalGameID:     game.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
			Status:             models.RpsRematchStatusPending,
			ExpiresAt:          time.Now().UTC().Add(45 * time.Second),
		})
		require.NoError(t, err)

		found, err := store.FindRpsRematchRequest(ctx, &RpsRematchFilter{
			OriginalGameIDs: []uuid.UUID{game.ID},
			Statuses:        []models.RpsRematchStatus{models.RpsRematchStatusPending},
		})
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
	})
}

func TestDBGamingStore_FindRpsRematchRequest_NotFound(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDBGamingStore(db)
		game, host, guest := mustCreateCompletedGame(t, ctx, store)

		_, err := store.CreateRpsRematchRequest(ctx, &models.RpsRematchRequest{
			OriginalGameID:     game.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
			Status:             models.RpsRematchStatusAccepted,
			ExpiresAt:          time.Now().UTC().Add(45 * time.Second),
		})
		require.NoError(t, err)

		// Filter for pending — should find nothing since we stored accepted.
		found, err := store.FindRpsRematchRequest(ctx, &RpsRematchFilter{
			OriginalGameIDs: []uuid.UUID{game.ID},
			Statuses:        []models.RpsRematchStatus{models.RpsRematchStatusPending},
		})
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestDBGamingStore_UpdateRpsRematchRequest(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDBGamingStore(db)
		game, host, guest := mustCreateCompletedGame(t, ctx, store)

		created, err := store.CreateRpsRematchRequest(ctx, &models.RpsRematchRequest{
			OriginalGameID:     game.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
			Status:             models.RpsRematchStatusPending,
			ExpiresAt:          time.Now().UTC().Add(45 * time.Second),
		})
		require.NoError(t, err)

		created.Status = models.RpsRematchStatusDeclined
		updated, err := store.UpdateRpsRematchRequest(ctx, created)
		require.NoError(t, err)
		assert.Equal(t, models.RpsRematchStatusDeclined, updated.Status)
	})
}

func TestDBGamingStore_FindExpiredPendingRpsRematches(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDBGamingStore(db)
		game, host, guest := mustCreateCompletedGame(t, ctx, store)

		// One already-expired pending request.
		_, err := store.CreateRpsRematchRequest(ctx, &models.RpsRematchRequest{
			OriginalGameID:     game.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
			Status:             models.RpsRematchStatusPending,
			ExpiresAt:          time.Now().UTC().Add(-1 * time.Second),
		})
		require.NoError(t, err)

		// One not-yet-expired pending request — needs a different game.
		game2, host2, guest2 := mustCreateCompletedGame(t, ctx, store)
		_, err = store.CreateRpsRematchRequest(ctx, &models.RpsRematchRequest{
			OriginalGameID:     game2.ID,
			RequestingPlayerID: host2.ID,
			InvitedPlayerID:    guest2.ID,
			Status:             models.RpsRematchStatusPending,
			ExpiresAt:          time.Now().UTC().Add(time.Hour),
		})
		require.NoError(t, err)

		expired, err := store.FindExpiredPendingRpsRematches(ctx)
		require.NoError(t, err)
		assert.Len(t, expired, 1)
		assert.Equal(t, game.ID, expired[0].OriginalGameID)
	})
}
