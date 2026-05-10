package stores

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestDBGamingStore_CreateUpdateCountPlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gamingStore := NewDBGamingStore(db)
		for i := range 10 {
			playerEmail := fmt.Sprintf("user%d@example.com", i)
			player, err := gamingStore.CreatePlayer(ctx, &models.Player{
				Email: playerEmail,
			})
			if err != nil {
				t.Fatalf("CreatePlayer() error = %v", err)
			}
			if player.ID == uuid.Nil {
				t.Errorf("CreatePlayer() = %v, want id not nil", player)
			}
			if player.Email != playerEmail {
				t.Errorf("CreatePlayer() = %v, want email %v", player, playerEmail)
			}
			if player.DisplayName != nil {
				t.Errorf("CreatePlayer() = %v, want display name nil", player)
			}
			displayName := fmt.Sprintf("display name %d", i)
			player.DisplayName = &displayName
			newPlayer, err := gamingStore.UpdatePlayer(ctx, player)
			if err != nil {
				t.Fatalf("UpdatePlayer() error = %v", err)
			}
			if *newPlayer.DisplayName != displayName {
				t.Errorf("UpdatePlayer() = %v, want display name %v", *newPlayer.DisplayName, displayName)
			}
			count, err := gamingStore.CountPlayers(ctx, nil)
			if err != nil {
				t.Fatalf("CountPlayers() error = %v", err)
			}
			if count != 1 {
				t.Errorf("CountPlayers() = %v, want %v", count, 1)
			}
			deleted, err := gamingStore.DeletePlayers(ctx, &PlayersFilter{
				Ids: []uuid.UUID{player.ID},
			})
			if err != nil {
				t.Fatalf("DeletePlayers() error = %v", err)
			}
			if deleted != 1 {
				t.Errorf("DeletePlayers() = %v, want %v", deleted, 1)
			}
			res, err := gamingStore.FindPlayers(ctx, nil)
			if err != nil {
				t.Fatalf("FindPlayers() error = %v", err)
			}
			if len(res) != 0 {
				t.Errorf("FindPlayers() = %v, want %v", len(res), 0)
			}
			newCount, err := gamingStore.CountPlayers(ctx, nil)
			if err != nil {
				t.Fatalf("CountPlayers() error = %v", err)
			}
			if newCount != 0 {
				t.Errorf("CountPlayers() = %v, want %v", count, 0)
			}
		}
	})
}

func TestDBGamingStore_CountPlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gamingStore := NewDBGamingStore(db)
		players := []*models.Player{}
		for i := range 10 {
			playerEmail := fmt.Sprintf("user%d@example.com", i)
			player, err := gamingStore.CreatePlayer(ctx, &models.Player{
				Email: playerEmail,
			})
			if err != nil {
				t.Fatalf("CreatePlayer() error = %v", err)
			}
			if player.ID == uuid.Nil {
				t.Errorf("CreatePlayer() = %v, want id not nil", player)
			}
			if player.Email != playerEmail {
				t.Errorf("CreatePlayer() = %v, want email %v", player, playerEmail)
			}
			if player.DisplayName != nil {
				t.Errorf("CreatePlayer() = %v, want display name nil", player)
			}
			displayName := fmt.Sprintf("display name %d", i)
			player.DisplayName = &displayName
			newPlayer, err := gamingStore.UpdatePlayer(ctx, player)
			if err != nil {
				t.Fatalf("UpdatePlayer() error = %v", err)
			}
			if *newPlayer.DisplayName != displayName {
				t.Errorf("UpdatePlayer() = %v, want display name %v", *newPlayer.DisplayName, displayName)
			}
			players = append(players, newPlayer)
		}
		res, err := gamingStore.FindPlayers(ctx, nil)
		if err != nil {
			t.Fatalf("FindPlayers() error = %v", err)
		}
		if len(res) != 10 {
			t.Errorf("FindPlayers() = %v, want %v", len(res), 10)
		}
		count, err := gamingStore.CountPlayers(ctx, nil)
		if err != nil {
			t.Fatalf("CountPlayers() error = %v", err)
		}
		if count != 10 {
			t.Errorf("CountPlayers() = %v, want %v", count, 10)
		}
	})
}

func TestDBGamingStore_FindPlayers_Registered(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := NewStorageAdapter(db)
		user := CreateUserWithOptions(t, adapter, UserWithEmail("testing@gmail.com"))
		player, err := adapter.Gaming().CreatePlayer(ctx, &models.Player{
			UserID: &user.User.ID,
			Email:  user.User.Email,
		})
		if err != nil {
			t.Fatalf("CreatePlayer() error = %v", err)
		}
		_, err = adapter.Gaming().CreatePlayer(ctx, &models.Player{
			Email: "no_user@example.com",
		})
		if err != nil {
			t.Fatalf("CreatePlayer() error = %v", err)
		}
		filter := &PlayersFilter{
			Registered: types.OptionalParam[bool]{
				Value: true,
				IsSet: true,
			},
		}
		foundPlayers, err := adapter.Gaming().FindPlayers(ctx, filter)
		if err != nil {
			t.Fatalf("FindPlayers() error = %v", err)
		}
		if len(foundPlayers) != 1 {
			t.Errorf("FindPlayers() got %d players, want 1", len(foundPlayers))
		}
		if foundPlayers[0].ID != player.ID {
			t.Errorf("FindPlayers() got player %v, want %v", foundPlayers[0].ID, player.ID)
		}
		if foundPlayers[0].UserID == nil {
			t.Errorf("FindPlayers() got player %v, want %v", foundPlayers[0].UserID, user.User.ID)
		}
	})
}
func TestDBGamingStore_FindPlayers_NotRegistered(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := NewStorageAdapter(db)
		user := CreateUserWithOptions(t, adapter, UserWithEmail("testing@gmail.com"))
		_, err := adapter.Gaming().CreatePlayer(ctx, &models.Player{
			UserID: &user.User.ID,
			Email:  user.User.Email,
		})
		if err != nil {
			t.Fatalf("CreatePlayer() error = %v", err)
		}
		noUserPlayer, err := adapter.Gaming().CreatePlayer(ctx, &models.Player{
			Email: "no_user@example.com",
		})
		if err != nil {
			t.Fatalf("CreatePlayer() error = %v", err)
		}
		filter := &PlayersFilter{
			Registered: types.OptionalParam[bool]{
				Value: false,
				IsSet: true,
			},
		}
		foundPlayers, err := adapter.Gaming().FindPlayers(ctx, filter)
		if err != nil {
			t.Fatalf("FindPlayers() error = %v", err)
		}
		if len(foundPlayers) != 1 {
			t.Errorf("FindPlayers() got %d players, want 1", len(foundPlayers))
		}
		if foundPlayers[0].ID != noUserPlayer.ID {
			t.Errorf("FindPlayers() got player %v, want %v", foundPlayers[0].ID, noUserPlayer.ID)
		}
		if foundPlayers[0].UserID != nil {
			t.Errorf("FindPlayers() got player %v, want %v", foundPlayers[0].UserID, nil)
		}
	})
}
func TestDBGamingStore_DeletePlayers_Registered(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := NewStorageAdapter(db)
		user := CreateUserWithOptions(t, adapter, UserWithEmail("testing@gmail.com"))
		_, err := adapter.Gaming().CreatePlayer(ctx, &models.Player{
			UserID: &user.User.ID,
			Email:  user.User.Email,
		})
		if err != nil {
			t.Fatalf("CreatePlayer() error = %v", err)
		}
		_, err = adapter.Gaming().CreatePlayer(ctx, &models.Player{
			Email: "no_user@example.com",
		})
		if err != nil {
			t.Fatalf("CreatePlayer() error = %v", err)
		}
		filter := &PlayersFilter{
			Registered: types.OptionalParam[bool]{
				Value: true,
				IsSet: true,
			},
		}
		deleted, err := adapter.Gaming().DeletePlayers(ctx, filter)
		if err != nil {
			t.Fatalf("FindPlayers() error = %v", err)
		}
		if deleted != 1 {
			t.Errorf("FindPlayers() got %d players, want 1", deleted)
		}
		playerCount, err := adapter.Gaming().CountPlayers(ctx, nil)
		if err != nil {
			t.Fatalf("FindPlayers() error = %v", err)
		}
		if playerCount != 1 {
			t.Errorf("FindPlayers() got %d players, want 1", playerCount)
		}
	})
}

func TestDBGamingStore_UpdatePlayerLastSeen(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDBGamingStore(db)
		player, err := store.CreatePlayer(ctx, &models.Player{Email: "lastseen@test.example"})
		require.NoError(t, err)
		assert.Nil(t, player.LastSeenAt)

		before := time.Now().UTC().Add(-time.Second)
		require.NoError(t, store.UpdatePlayerLastSeen(ctx, player.ID))
		after := time.Now().UTC().Add(time.Second)

		updated, err := store.FindPlayer(ctx, &PlayersFilter{Ids: []uuid.UUID{player.ID}})
		require.NoError(t, err)
		require.NotNil(t, updated.LastSeenAt)
		assert.True(t, updated.LastSeenAt.After(before), "last_seen_at should be after test start")
		assert.True(t, updated.LastSeenAt.Before(after), "last_seen_at should be before test end")
	})
}

func TestDBGamingStore_UpdatePlayerLastSeen_IdempotentOnRepeatCalls(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		store := NewDBGamingStore(db)
		player, err := store.CreatePlayer(ctx, &models.Player{Email: "lastseen2@test.example"})
		require.NoError(t, err)

		require.NoError(t, store.UpdatePlayerLastSeen(ctx, player.ID))
		require.NoError(t, store.UpdatePlayerLastSeen(ctx, player.ID))

		updated, err := store.FindPlayer(ctx, &PlayersFilter{Ids: []uuid.UUID{player.ID}})
		require.NoError(t, err)
		assert.NotNil(t, updated.LastSeenAt)
	})
}
