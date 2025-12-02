package stores

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
)

func TestDBGamingStore_CreateRpsParticipant(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		rpsGameStatusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		moveSelector := test.NewRandomeSelector(models.RpsParticipantMoveRock, models.RpsParticipantMovePaper, models.RpsParticipantMoveScissors)
		typeSelector := test.NewRandomeSelector(models.RpsParticipantTypeGuest, models.RpsParticipantTypeHost)
		rpsParticipantStatusSelector := test.NewRandomeSelector(models.RpsParticipantStatusPending, models.RpsParticipantStatusDeclined, models.RpsParticipantStatusCompleted)
		rpsParticipantResultSelector := test.NewRandomeSelector(models.RpsParticipantResultWin, models.RpsParticipantResultTie, models.RpsParticipantResultLose)
		for i := range 10 {
			game := &models.RpsGame{
				Status:    rpsGameStatusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			player, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d@example.com", i),
			})
			if err != nil {
				t.Errorf("CreatePlayer error: %v", err)
			}
			participant := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     typeSelector.Select(),
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
				// RespondedAt: ,
			}
			createdParticipant, err := gameStore.CreateRpsParticipant(ctx, participant)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			if createdParticipant.ID == uuid.Nil {
				t.Errorf("CreateRpsParticipant.ID() = %v, want id not nil", createdParticipant.ID)
			}
			if createdParticipant.GameID != participant.GameID {
				t.Errorf("CreateRpsParticipant.GameID() = %v, want game id %v", createdParticipant.GameID, participant.GameID)
			}
			if createdParticipant.PlayerID != participant.PlayerID {
				t.Errorf("CreateRpsParticipant.PlayerID() = %v, want player id %v", createdParticipant.PlayerID, participant.PlayerID)
			}
			if createdParticipant.Move != participant.Move {
				t.Errorf("CreateRpsParticipant.Move() = %v, want move %v", createdParticipant.Move, participant.Move)
			}
			if createdParticipant.Type != participant.Type {
				t.Errorf("CreateRpsParticipant.Type() = %v, want type %v", createdParticipant.Type, participant.Type)
			}
			if createdParticipant.Status != participant.Status {
				t.Errorf("CreateRpsParticipant.Status() = %v, want status %v", createdParticipant.Status, participant.Status)
			}
			if createdParticipant.Result != participant.Result {
				t.Errorf("CreateRpsParticipant.Result() = %v, want result %v", createdParticipant.Result, participant.Result)
			}
		}
	})
}

func TestDBGamingStore_CreateRpsParticipant_ParticipantTypeConstraint(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		rpsGameStatusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		moveSelector := test.NewRandomeSelector(models.RpsParticipantMoveRock, models.RpsParticipantMovePaper, models.RpsParticipantMoveScissors)
		rpsParticipantStatusSelector := test.NewRandomeSelector(models.RpsParticipantStatusPending, models.RpsParticipantStatusDeclined, models.RpsParticipantStatusCompleted)
		rpsParticipantResultSelector := test.NewRandomeSelector(models.RpsParticipantResultWin, models.RpsParticipantResultTie, models.RpsParticipantResultLose)
		for i := range 1 {
			game := &models.RpsGame{
				Status:    rpsGameStatusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Fatalf("CreateRpsGame error: %v", err)
			}
			player, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d@example.com", i),
			})
			if err != nil {
				t.Fatalf("CreatePlayer error: %v", err)
			}
			player2, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d-2@example.com", i),
			})
			if err != nil {
				t.Fatalf("CreatePlayer error: %v", err)
			}
			player3, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d-3@example.com", i),
			})
			if err != nil {
				t.Fatalf("CreatePlayer error: %v", err)
			}
			participant := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeGuest,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			_, err = gameStore.CreateRpsParticipant(ctx, participant)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			participant2 := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player2.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeHost,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			_, err = gameStore.CreateRpsParticipant(ctx, participant2)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			participant3 := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player3.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeHost,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			_, err = gameStore.CreateRpsParticipant(ctx, participant3)
			if err == nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			if !strings.Contains(err.Error(), "violates unique constraint") {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
		}
	})
}
func TestDBGamingStore_CreateRpsParticipant_ParticipantPlayerIDConstraint(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		rpsGameStatusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		moveSelector := test.NewRandomeSelector(models.RpsParticipantMoveRock, models.RpsParticipantMovePaper, models.RpsParticipantMoveScissors)
		rpsParticipantStatusSelector := test.NewRandomeSelector(models.RpsParticipantStatusPending, models.RpsParticipantStatusDeclined, models.RpsParticipantStatusCompleted)
		rpsParticipantResultSelector := test.NewRandomeSelector(models.RpsParticipantResultWin, models.RpsParticipantResultTie, models.RpsParticipantResultLose)
		for i := range 1 {
			game := &models.RpsGame{
				Status:    rpsGameStatusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Fatalf("CreateRpsGame error: %v", err)
			}
			player, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d@example.com", i),
			})
			if err != nil {
				t.Fatalf("CreatePlayer error: %v", err)
			}

			participant := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeGuest,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			_, err = gameStore.CreateRpsParticipant(ctx, participant)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			participant2 := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeHost,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			_, err = gameStore.CreateRpsParticipant(ctx, participant2)
			if err == nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			if !strings.Contains(err.Error(), "violates unique constraint") {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
		}
	})
}

func TestDBGamingStore_UpdateRpsParticipant(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		rpsGameStatusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		moveSelector := test.NewRandomeSelector(models.RpsParticipantMoveRock, models.RpsParticipantMovePaper, models.RpsParticipantMoveScissors)
		typeSelector := test.NewRandomeSelector(models.RpsParticipantTypeGuest, models.RpsParticipantTypeHost)
		rpsParticipantStatusSelector := test.NewRandomeSelector(models.RpsParticipantStatusPending, models.RpsParticipantStatusDeclined, models.RpsParticipantStatusCompleted)
		rpsParticipantResultSelector := test.NewRandomeSelector(models.RpsParticipantResultWin, models.RpsParticipantResultTie, models.RpsParticipantResultLose)
		for i := range 10 {
			game := &models.RpsGame{
				Status:    rpsGameStatusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			player, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d@example.com", i),
			})
			if err != nil {
				t.Errorf("CreatePlayer error: %v", err)
			}
			participant := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     typeSelector.Select(),
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
				// RespondedAt: ,
			}
			createdParticipant, err := gameStore.CreateRpsParticipant(ctx, participant)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			if createdParticipant.ID == uuid.Nil {
				t.Errorf("CreateRpsParticipant.ID() = %v, want id not nil", createdParticipant.ID)
			}
			if createdParticipant.GameID != participant.GameID {
				t.Errorf("CreateRpsParticipant.GameID() = %v, want game id %v", createdParticipant.GameID, participant.GameID)
			}
			if createdParticipant.PlayerID != participant.PlayerID {
				t.Errorf("CreateRpsParticipant.PlayerID() = %v, want player id %v", createdParticipant.PlayerID, participant.PlayerID)
			}
			if createdParticipant.Move != participant.Move {
				t.Errorf("CreateRpsParticipant.Move() = %v, want move %v", createdParticipant.Move, participant.Move)
			}
			if createdParticipant.Type != participant.Type {
				t.Errorf("CreateRpsParticipant.Type() = %v, want type %v", createdParticipant.Type, participant.Type)
			}
			if createdParticipant.Status != participant.Status {
				t.Errorf("CreateRpsParticipant.Status() = %v, want status %v", createdParticipant.Status, participant.Status)
			}
			if createdParticipant.Result != participant.Result {
				t.Errorf("CreateRpsParticipant.Result() = %v, want result %v", createdParticipant.Result, participant.Result)
			}

			// update
			createdParticipant.Move = moveSelector.Select()
			createdParticipant.Type = typeSelector.Select()
			createdParticipant.Status = rpsParticipantStatusSelector.Select()
			createdParticipant.Result = rpsParticipantResultSelector.Select()
			updatedParticipant, err := gameStore.UpdateRpsParticipant(ctx, createdParticipant)
			if err != nil {
				t.Errorf("UpdateRpsParticipant error: %v", err)
			}
			if updatedParticipant.ID != createdParticipant.ID {
				t.Errorf("UpdateRpsParticipant.ID() = %v, want id %v", updatedParticipant.ID, createdParticipant.ID)
			}
			if updatedParticipant.GameID != createdParticipant.GameID {
				t.Errorf("UpdateRpsParticipant.GameID() = %v, want game id %v", updatedParticipant.GameID, createdParticipant.GameID)
			}
			if updatedParticipant.PlayerID != createdParticipant.PlayerID {
				t.Errorf("UpdateRpsParticipant.PlayerID() = %v, want player id %v", updatedParticipant.PlayerID, createdParticipant.PlayerID)
			}
			if updatedParticipant.Move != createdParticipant.Move {
				t.Errorf("UpdateRpsParticipant.Move() = %v, want move %v", updatedParticipant.Move, createdParticipant.Move)
			}
			if updatedParticipant.Type != createdParticipant.Type {
				t.Errorf("UpdateRpsParticipant.Type() = %v, want type %v", updatedParticipant.Type, createdParticipant.Type)
			}
			if updatedParticipant.Status != createdParticipant.Status {
				t.Errorf("UpdateRpsParticipant.Status() = %v, want status %v", updatedParticipant.Status, createdParticipant.Status)
			}
			if updatedParticipant.Result != createdParticipant.Result {
				t.Errorf("UpdateRpsParticipant.Result() = %v, want result %v", updatedParticipant.Result, createdParticipant.Result)
			}
		}
	})
}
