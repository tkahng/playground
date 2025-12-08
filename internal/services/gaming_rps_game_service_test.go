package services

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestDbRpsGameService_RequestGame_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		rpsService := NewDbRpsGameService(adapter)
		player1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("player1@gmail.com"))
		player2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("player2@gmail.com"))
		requestInput := &RpsGameRequestInput{
			RequestingPlayerID:   player1.ID,
			InvitedPlayerID:      player2.ID,
			RequestingPlayerMove: models.RpsParticipantMovePaper,
			DurationSeconds:      60 * 60 * 24 * 7,
		}
		game, err := rpsService.RequestGame(ctx, requestInput)
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}
		if game.CompletedAt != nil {
			t.Errorf("Expected game.CompletedAt to be nil, got %v", game.CompletedAt)
		}
		// if game.ExpiresAt !=
		if game.RequestingParticipant.PlayerID != player1.ID {
			t.Errorf("Expected game.RequestingParticipant.PlayerID to be %s, got %s", player1.ID, game.RequestingParticipant.PlayerID)
		}
		if game.RequestingParticipant.Move != models.RpsParticipantMovePaper {
			t.Errorf("Expected game.RequestingParticipant.Move to be %s, got %s", models.RpsParticipantMovePaper, game.RequestingParticipant.Move)
		}
		if game.InvitedParticipant.PlayerID != player2.ID {
			t.Errorf("Expected game.InvitedParticipant.PlayerID to be %s, got %s", player2.ID, game.InvitedParticipant.PlayerID)
		}
		if game.Status != models.RpsGameStatusPending {
			t.Errorf("Expected game.Status to be %s, got %s", models.RpsGameStatusPending, game.Status)
		}
	})
}

func TestDbRpsGameService_RespondToGameRequest_Success_InvitedPlayer_Win(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		rpsService := NewDbRpsGameService(adapter)
		player1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("player1@gmail.com"))
		player2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("player2@gmail.com"))
		requestInput := &RpsGameRequestInput{
			RequestingPlayerID:   player1.ID,
			InvitedPlayerID:      player2.ID,
			RequestingPlayerMove: models.RpsParticipantMovePaper,
			DurationSeconds:      60 * 60 * 24 * 7,
		}
		game, err := rpsService.RequestGame(ctx, requestInput)
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		respondInput := &GameRequestResponse{
			InvitedPlayerID: player2.ID,
			GameID:          game.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveScissors,
		}
		respondedGame, err := rpsService.RespondToGameRequest(ctx, respondInput)
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}
		if respondedGame.ID != game.ID {
			t.Errorf("Expected respondedGame.ID to be %s, got %s", game.ID, respondedGame.ID)
		}
		if respondedGame.Status != models.RpsGameStatusCompleted {
			t.Errorf("Expected respondedGame.Status to be %s, got %s", models.RpsGameStatusCompleted, respondedGame.Status)
		}
		if respondedGame.RequestingParticipant.Result != models.RpsParticipantResultLose {
			t.Errorf("Expected respondedGame.RequestingParticipant.Result to be %s, got %s", models.RpsParticipantResultLose, respondedGame.RequestingParticipant.Result)
		}
		if respondedGame.InvitedParticipant.Result != models.RpsParticipantResultWin {
			t.Errorf("Expected respondedGame.InvitedParticipant.Result to be %s, got %s", models.RpsParticipantResultWin, respondedGame.InvitedParticipant.Result)
		}
	})
}
func TestDbRpsGameService_RespondToGameRequest_Success_InvitedPlayer_Lose(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		rpsService := NewDbRpsGameService(adapter)
		player1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("player1@gmail.com"))
		player2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("player2@gmail.com"))
		requestInput := &RpsGameRequestInput{
			RequestingPlayerID:   player1.ID,
			InvitedPlayerID:      player2.ID,
			RequestingPlayerMove: models.RpsParticipantMovePaper,
			DurationSeconds:      60 * 60 * 24 * 7,
		}
		game, err := rpsService.RequestGame(ctx, requestInput)
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		respondInput := &GameRequestResponse{
			InvitedPlayerID: player2.ID,
			GameID:          game.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveRock,
		}
		respondedGame, err := rpsService.RespondToGameRequest(ctx, respondInput)
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}
		if respondedGame.ID != game.ID {
			t.Errorf("Expected respondedGame.ID to be %s, got %s", game.ID, respondedGame.ID)
		}
		if respondedGame.Status != models.RpsGameStatusCompleted {
			t.Errorf("Expected respondedGame.Status to be %s, got %s", models.RpsGameStatusCompleted, respondedGame.Status)
		}
		if respondedGame.RequestingParticipant.Result != models.RpsParticipantResultWin {
			t.Errorf("Expected respondedGame.RequestingParticipant.Result to be %s, got %s", models.RpsParticipantResultWin, respondedGame.RequestingParticipant.Result)
		}
		if respondedGame.InvitedParticipant.Result != models.RpsParticipantResultLose {
			t.Errorf("Expected respondedGame.InvitedParticipant.Result to be %s, got %s", models.RpsParticipantResultLose, respondedGame.InvitedParticipant.Result)
		}
	})
}
func TestDbRpsGameService_RespondToGameRequest_Success_InvitedPlayer_Tie(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		rpsService := NewDbRpsGameService(adapter)
		player1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("player1@gmail.com"))
		player2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("player2@gmail.com"))
		requestInput := &RpsGameRequestInput{
			RequestingPlayerID:   player1.ID,
			InvitedPlayerID:      player2.ID,
			RequestingPlayerMove: models.RpsParticipantMovePaper,
			DurationSeconds:      60 * 60 * 24 * 7,
		}
		game, err := rpsService.RequestGame(ctx, requestInput)
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		respondInput := &GameRequestResponse{
			InvitedPlayerID: player2.ID,
			GameID:          game.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMovePaper,
		}
		respondedGame, err := rpsService.RespondToGameRequest(ctx, respondInput)
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}
		if respondedGame.ID != game.ID {
			t.Errorf("Expected respondedGame.ID to be %s, got %s", game.ID, respondedGame.ID)
		}
		if respondedGame.Status != models.RpsGameStatusCompleted {
			t.Errorf("Expected respondedGame.Status to be %s, got %s", models.RpsGameStatusCompleted, respondedGame.Status)
		}
		if respondedGame.RequestingParticipant.Result != models.RpsParticipantResultTie {
			t.Errorf("Expected respondedGame.RequestingParticipant.Result to be %s, got %s", models.RpsParticipantResultTie, respondedGame.RequestingParticipant.Result)
		}
		if respondedGame.InvitedParticipant.Result != models.RpsParticipantResultTie {
			t.Errorf("Expected respondedGame.InvitedParticipant.Result to be %s, got %s", models.RpsParticipantResultTie, respondedGame.InvitedParticipant.Result)
		}
	})
}
