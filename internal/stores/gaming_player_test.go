package stores

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

func TestDBGamingStore_CreatePlayer(t *testing.T) {
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
			count, err = gamingStore.CountPlayers(ctx, nil)
			if err != nil {
				t.Fatalf("CountPlayers() error = %v", err)
			}
			if count != 0 {
				t.Errorf("CountPlayers() = %v, want %v", count, 0)
			}
		}
	})
}
