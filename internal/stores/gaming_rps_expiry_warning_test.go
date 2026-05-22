//go:build integration

package stores

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

func TestFindPendingGamesExpiringWithin(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDBGamingStore(db)
		now := time.Now().UTC()

		// Game expiring in 12 h — within 24 h window.
		gameExpiring, err := store.CreateRpsGame(ctx, &models.RpsGame{
			Status:    models.RpsGameStatusPending,
			ExpiresAt: now.Add(12 * time.Hour),
			Metadata:  []byte("{}"),
		})
		require.NoError(t, err)

		// Game expiring in 48 h — outside window.
		gameFar, err := store.CreateRpsGame(ctx, &models.RpsGame{
			Status:    models.RpsGameStatusPending,
			ExpiresAt: now.Add(48 * time.Hour),
			Metadata:  []byte("{}"),
		})
		require.NoError(t, err)

		// Completed game expiring in 12 h — must be excluded (only pending counted).
		gameCompleted, err := store.CreateRpsGame(ctx, &models.RpsGame{
			Status:    models.RpsGameStatusCompleted,
			ExpiresAt: now.Add(12 * time.Hour),
			Metadata:  []byte("{}"),
		})
		require.NoError(t, err)

		games, err := store.FindPendingGamesExpiringWithin(ctx, 24*time.Hour)
		require.NoError(t, err)

		ids := make(map[string]bool, len(games))
		for _, g := range games {
			ids[g.ID.String()] = true
		}
		assert.True(t, ids[gameExpiring.ID.String()], "soon-expiring pending game should be returned")
		assert.False(t, ids[gameFar.ID.String()], "far-future game should not be returned")
		assert.False(t, ids[gameCompleted.ID.String()], "completed game should not be returned")

		// After MarkRpsGameExpirySent, the game must not appear again.
		err = store.MarkRpsGameExpirySent(ctx, gameExpiring)
		require.NoError(t, err)

		games2, err := store.FindPendingGamesExpiringWithin(ctx, 24*time.Hour)
		require.NoError(t, err)
		for _, g := range games2 {
			assert.NotEqual(t, gameExpiring.ID, g.ID, "already-warned game must not be returned again")
		}
	})
}
