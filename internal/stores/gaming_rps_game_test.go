package stores

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
)

func TestDBGamingStore_CreateRpsGame(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		statusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		for i := range 10 {
			game := &models.RpsGame{
				Status:    statusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			if createdGame.ID == uuid.Nil {
				t.Errorf("CreateRpsGame.ID() = %v, want id not nil", createdGame.ID)
			}
			if createdGame.Status != game.Status {
				t.Errorf("CreateRpsGame.Status() = %v, want status %v", createdGame.Status, game.Status)
			}
			if !createdGame.ExpiresAt.UTC().Equal(game.ExpiresAt.UTC()) {
				t.Errorf("CreateRpsGame.ExpiresAt() = %v, want expires at %v", createdGame.ExpiresAt, game.ExpiresAt)
			}
			if string(createdGame.Metadata) != string(game.Metadata) {
				t.Errorf("CreateRpsGame.Metadata() = %v, want metadata %v", createdGame.Metadata, game.Metadata)
			}
		}
	})
}

func TestDBGamingStore_UpdateRpsGame(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		statusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		for i := range 10 {
			game := &models.RpsGame{
				Status:    statusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			if createdGame.ID == uuid.Nil {
				t.Errorf("CreateRpsGame.ID() = %v, want id not nil", createdGame.ID)
			}
			if createdGame.Status != game.Status {
				t.Errorf("CreateRpsGame.Status() = %v, want status %v", createdGame.Status, game.Status)
			}
			if !createdGame.ExpiresAt.UTC().Equal(game.ExpiresAt.UTC()) {
				t.Errorf("CreateRpsGame.ExpiresAt() = %v, want expires at %v", createdGame.ExpiresAt, game.ExpiresAt)
			}
			if string(createdGame.Metadata) != string(game.Metadata) {
				t.Errorf("CreateRpsGame.Metadata() = %v, want metadata %v", createdGame.Metadata, game.Metadata)
			}
			createdGame.Status = statusSelector.Select()
			createdGame.Metadata = fmt.Appendf(nil, `{"idx": %d, "updated": true}`, i)
			createdGame.ExpiresAt = time.Now().Add(time.Minute * 3600).UTC()
			updatedGame, err := gameStore.UpdateRpsGame(ctx, createdGame)
			if err != nil {
				t.Errorf("UpdateRpsGame error: %v", err)
			}
			if updatedGame.ID != createdGame.ID {
				t.Errorf("UpdateRpsGame.ID() = %v, want id %v", updatedGame.ID, createdGame.ID)
			}
			if updatedGame.Status != createdGame.Status {
				t.Errorf("UpdateRpsGame.Status() = %v, want status %v", updatedGame.Status, createdGame.Status)
			}
			if !updatedGame.ExpiresAt.UTC().Equal(createdGame.ExpiresAt.UTC()) {
				t.Errorf("UpdateRpsGame.ExpiresAt() = %v, want expires at %v", updatedGame.ExpiresAt, createdGame.ExpiresAt)
			}
			if string(updatedGame.Metadata) != string(createdGame.Metadata) {
				t.Errorf("UpdateRpsGame.Metadata() = %v, want metadata %v", updatedGame.Metadata, createdGame.Metadata)
			}
		}
	})
}
