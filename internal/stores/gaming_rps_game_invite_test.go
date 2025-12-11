package stores

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

func TestDBGamingStore_CreateRpsGameInvitation(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		game, err := gameStore.CreateRpsGame(ctx, &models.RpsGame{
			ExpiresAt: time.Now().UTC(),
			Status:    models.RpsGameStatusPending,
			Metadata:  []byte("{}"),
		})
		player := MustCreatePlayer(t, ctx, gameStore, WithPlayerEmail("player1@gmail.com"))
		otherPlayer := MustCreatePlayer(t, ctx, gameStore, WithPlayerEmail("player2@gmail.com"))
		if err != nil {
			t.Errorf("CreateRpsGame error: %v", err)
		}
		for i := range 5 {
			invite := &models.RpsGameInvite{
				GameID:             game.ID,
				RequestingPlayerID: player.ID,
				InvitedPlayerID:    otherPlayer.ID,
				Token:              uuid.NewString(),
				ExpiresAt:          time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:           fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdInvite, err := gameStore.CreateRpsGameInvite(ctx, invite)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			if createdInvite.ID == uuid.Nil {
				t.Errorf("CreateRpsGameInvite.ID() = %v, want id not nil", createdInvite.ID)
			}
			if !createdInvite.ExpiresAt.UTC().Equal(invite.ExpiresAt.UTC()) {
				t.Errorf("CreateRpsGameInvite.ExpiresAt() = %v, want expires at %v", createdInvite.ExpiresAt, invite.ExpiresAt)
			}
			if string(createdInvite.Metadata) != string(invite.Metadata) {
				t.Errorf("CreateRpsGameInvite.Metadata() = %v, want metadata %v", createdInvite.Metadata, invite.Metadata)
			}
			if createdInvite.RequestingPlayerID != invite.RequestingPlayerID {
				t.Errorf("CreateRpsGameInvite.RequestingPlayerID() = %v, want requesting player id %v", createdInvite.RequestingPlayerID, invite.RequestingPlayerID)
			}
			if createdInvite.InvitedPlayerID != invite.InvitedPlayerID {
				t.Errorf("CreateRpsGameInvite.InvitedPlayerID() = %v, want invited player id %v", createdInvite.InvitedPlayerID, invite.InvitedPlayerID)
			}
			if createdInvite.Token != invite.Token {
				t.Errorf("CreateRpsGameInvite.Token() = %v, want token %v", createdInvite.Token, invite.Token)
			}
		}
	})
}
func TestDBGamingStore_UpdateRpsGameInvitation(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		game, err := gameStore.CreateRpsGame(ctx, &models.RpsGame{
			ExpiresAt: time.Now().UTC(),
			Status:    models.RpsGameStatusPending,
			Metadata:  []byte("{}"),
		})
		player := MustCreatePlayer(t, ctx, gameStore, WithPlayerEmail("player1@gmail.com"))
		otherPlayer := MustCreatePlayer(t, ctx, gameStore, WithPlayerEmail("player2@gmail.com"))
		if err != nil {
			t.Errorf("CreateRpsGame error: %v", err)
		}
		for i := range 5 {
			invite := &models.RpsGameInvite{
				GameID:             game.ID,
				RequestingPlayerID: player.ID,
				InvitedPlayerID:    otherPlayer.ID,
				Token:              uuid.NewString(),
				ExpiresAt:          time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:           fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdInvite, err := gameStore.CreateRpsGameInvite(ctx, invite)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			createdInvite.Metadata = fmt.Appendf(nil, `{"idx": %d, "updated": true}`, i)
			createdInvite.ExpiresAt = time.Now().Add(time.Minute * 3600).UTC()
			createdInvite.Token = uuid.NewString()
			updatedInvite, err := gameStore.UpdateRpsGameInvite(ctx, createdInvite)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			if updatedInvite.ID == uuid.Nil {
				t.Errorf("CreateRpsGameInvite.ID() = %v, want id not nil", createdInvite.ID)
			}
			if !updatedInvite.ExpiresAt.UTC().Equal(createdInvite.ExpiresAt.UTC()) {
				t.Errorf("CreateRpsGameInvite.ExpiresAt() = %v, want expires at %v", updatedInvite.ExpiresAt, createdInvite.ExpiresAt)
			}
			if string(updatedInvite.Metadata) != string(createdInvite.Metadata) {
				t.Errorf("CreateRpsGameInvite.Metadata() = %v, want metadata %v", updatedInvite.Metadata, createdInvite.Metadata)
			}
			if updatedInvite.RequestingPlayerID != createdInvite.RequestingPlayerID {
				t.Errorf("CreateRpsGameInvite.RequestingPlayerID() = %v, want requesting player id %v", updatedInvite.RequestingPlayerID, createdInvite.RequestingPlayerID)
			}
			if updatedInvite.InvitedPlayerID != createdInvite.InvitedPlayerID {
				t.Errorf("CreateRpsGameInvite.InvitedPlayerID() = %v, want invited player id %v", updatedInvite.InvitedPlayerID, createdInvite.InvitedPlayerID)
			}
			if updatedInvite.Token != createdInvite.Token {
				t.Errorf("CreateRpsGameInvite.Token() = %v, want token %v", updatedInvite.Token, createdInvite.Token)
			}
		}
	})
}
func TestDBGamingStore_DeleteRpsGameInvitation(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		game, err := gameStore.CreateRpsGame(ctx, &models.RpsGame{
			ExpiresAt: time.Now().UTC(),
			Status:    models.RpsGameStatusPending,
			Metadata:  []byte("{}"),
		})
		player := MustCreatePlayer(t, ctx, gameStore, WithPlayerEmail("player1@gmail.com"))
		otherPlayer := MustCreatePlayer(t, ctx, gameStore, WithPlayerEmail("player2@gmail.com"))
		if err != nil {
			t.Errorf("CreatePlayer error: %v", err)
		}
		for i := range 5 {
			invite := &models.RpsGameInvite{
				GameID:             game.ID,
				RequestingPlayerID: player.ID,
				InvitedPlayerID:    otherPlayer.ID,
				Token:              uuid.NewString(),
				ExpiresAt:          time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:           fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdInvite, err := gameStore.CreateRpsGameInvite(ctx, invite)
			if err != nil {
				t.Errorf("CreateRpsGameInvite error: %v", err)
			}
			otherInvite := &models.RpsGameInvite{
				GameID:             game.ID,
				RequestingPlayerID: player.ID,
				InvitedPlayerID:    otherPlayer.ID,
				Token:              uuid.NewString(),
				ExpiresAt:          time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:           fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdOtherInvite, err := gameStore.CreateRpsGameInvite(ctx, otherInvite)
			if err != nil {
				t.Errorf("CreateRpsGameInvite error: %v", err)
			}
			deletedCount, err := gameStore.DeleteRpGameInvites(ctx, &RpsGameInviteFilter{
				Ids: []uuid.UUID{createdInvite.ID},
			})
			if err != nil {
				t.Errorf("deleteRpsGame error: %v", err)
			}
			if deletedCount != 1 {
				t.Errorf("deletedCount = %v, want 1", deletedCount)
			}
			otherInviteFound, err := gameStore.FindRpsGameInvite(ctx, &RpsGameInviteFilter{
				Ids: []uuid.UUID{createdOtherInvite.ID},
			})
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			if otherInviteFound.ID != createdOtherInvite.ID {
				t.Errorf("otherInviteFound.ID() = %v, want id %v", otherInviteFound.ID, createdOtherInvite.ID)
			}
		}
	})
}
