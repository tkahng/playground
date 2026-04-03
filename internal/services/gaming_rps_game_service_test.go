package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestDbRpsGameService_RequestGame_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)
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
		if game.RpsGame.CompletedAt != nil {
			t.Errorf("Expected game.CompletedAt to be nil, got %v", game.RpsGame.CompletedAt)
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
		if game.RpsGame.Status != models.RpsGameStatusPending {
			t.Errorf("Expected game.RpsGame.Status to be %s, got %s", models.RpsGameStatusPending, game.RpsGame.Status)
		}
	})
}

func TestDbRpsGameService_RespondToGameRequest_Success_InvitedPlayer_Win(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

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
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveScissors,
		}
		respondedGame, err := rpsService.RespondToGameRequest(ctx, respondInput)
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}
		if respondedGame.RpsGame.ID != game.RpsGame.ID {
			t.Errorf("Expected respondedgame.RpsGame.ID to be %s, got %s", game.RpsGame.ID, respondedGame.RpsGame.ID)
		}
		if respondedGame.RpsGame.Status != models.RpsGameStatusCompleted {
			t.Errorf("Expected respondedGame.RpsGame.Status to be %s, got %s", models.RpsGameStatusCompleted, respondedGame.RpsGame.Status)
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
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)
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
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveRock,
		}
		respondedGame, err := rpsService.RespondToGameRequest(ctx, respondInput)
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}
		if respondedGame.RpsGame.ID != game.RpsGame.ID {
			t.Errorf("Expected respondedgame.RpsGame.ID to be %s, got %s", game.RpsGame.ID, respondedGame.RpsGame.ID)
		}
		if respondedGame.RpsGame.Status != models.RpsGameStatusCompleted {
			t.Errorf("Expected respondedGame.RpsGame.Status to be %s, got %s", models.RpsGameStatusCompleted, respondedGame.RpsGame.Status)
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
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)
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
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMovePaper,
		}
		respondedGame, err := rpsService.RespondToGameRequest(ctx, respondInput)
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}
		if respondedGame.RpsGame.ID != game.RpsGame.ID {
			t.Errorf("Expected respondedgame.RpsGame.ID to be %s, got %s", game.RpsGame.ID, respondedGame.RpsGame.ID)
		}
		if respondedGame.RpsGame.Status != models.RpsGameStatusCompleted {
			t.Errorf("Expected respondedGame.RpsGame.Status to be %s, got %s", models.RpsGameStatusCompleted, respondedGame.RpsGame.Status)
		}
		if respondedGame.RequestingParticipant.Result != models.RpsParticipantResultTie {
			t.Errorf("Expected respondedGame.RequestingParticipant.Result to be %s, got %s", models.RpsParticipantResultTie, respondedGame.RequestingParticipant.Result)
		}
		if respondedGame.InvitedParticipant.Result != models.RpsParticipantResultTie {
			t.Errorf("Expected respondedGame.InvitedParticipant.Result to be %s, got %s", models.RpsParticipantResultTie, respondedGame.InvitedParticipant.Result)
		}
	})
}

func TestDbRpsGameService_RespondToGameRequest_Fail_Expired(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)
		player1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("player1@gmail.com"))
		player2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("player2@gmail.com"))
		requestInput := &RpsGameRequestInput{
			RequestingPlayerID:   player1.ID,
			InvitedPlayerID:      player2.ID,
			RequestingPlayerMove: models.RpsParticipantMovePaper,
			DurationSeconds:      1,
		}
		game, err := rpsService.RequestGame(ctx, requestInput)
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}
		time.Sleep(time.Second * 1)
		respondInput := &GameRequestResponse{
			InvitedPlayerID: player2.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMovePaper,
		}
		_, err = rpsService.RespondToGameRequest(ctx, respondInput)
		if err == nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}
		if err.Error() != "game expired" {
			t.Errorf("Expected error to be 'game expired', got %s", err.Error())
		}
	})
}

// mustFundPlayerWallet is a test helper that credits a user-linked player's wallet.
func mustFundPlayerWallet(t *testing.T, ctx context.Context, adapter stores.StorageAdapterInterface, ledger LedgerService, userID *uuid.UUID, amount int64) {
	t.Helper()
	issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
	if err != nil {
		t.Fatalf("mustFundPlayerWallet: get issuance account: %v", err)
	}
	wallet, err := ledger.GetOrCreateUserWallet(ctx, *userID)
	if err != nil {
		t.Fatalf("mustFundPlayerWallet: get wallet: %v", err)
	}
	_, err = ledger.PostTransfer(ctx, PostTransferInput{
		LedgerCode:      "POINTS",
		DebitAccountID:  issuance.ID,
		CreditAccountID: wallet.ID,
		Amount:          amount,
		TransferCode:    models.TransferCodePurchase,
	})
	if err != nil {
		t.Fatalf("mustFundPlayerWallet: post transfer: %v", err)
	}
}

func TestDbRpsGameService_Betting_HostWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		// Players must have linked user IDs for betting.
		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("bethost@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "bethost@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("betguest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "betguest@example.com").ID),
		)

		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      60 * 60 * 24,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}
		if game.RpsGame.HostBetTransferID == nil {
			t.Fatal("expected HostBetTransferID to be set after RequestGame with bet")
		}

		// Scissors loses to Rock → host wins.
		responded, err := rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveScissors,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}
		if responded.RequestingParticipant.Result != models.RpsParticipantResultWin {
			t.Errorf("host result = %v, want win", responded.RequestingParticipant.Result)
		}

		hostBal, _ := ledger.GetUserBalance(ctx, *host.UserID)
		guestBal, _ := ledger.GetUserBalance(ctx, *guest.UserID)
		if hostBal != 600 {
			t.Errorf("host balance = %d, want 600", hostBal)
		}
		if guestBal != 400 {
			t.Errorf("guest balance = %d, want 400", guestBal)
		}
	})
}

func TestDbRpsGameService_Betting_GuestWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("bethost2@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "bethost2@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("betguest2@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "betguest2@example.com").ID),
		)

		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      60 * 60 * 24,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		// Paper beats Rock → guest wins.
		responded, err := rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMovePaper,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}
		if responded.InvitedParticipant.Result != models.RpsParticipantResultWin {
			t.Errorf("guest result = %v, want win", responded.InvitedParticipant.Result)
		}

		hostBal, _ := ledger.GetUserBalance(ctx, *host.UserID)
		guestBal, _ := ledger.GetUserBalance(ctx, *guest.UserID)
		if hostBal != 400 {
			t.Errorf("host balance = %d, want 400", hostBal)
		}
		if guestBal != 600 {
			t.Errorf("guest balance = %d, want 600", guestBal)
		}
	})
}

func TestDbRpsGameService_Betting_Tie_BothRefunded(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("bethost3@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "bethost3@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("betguest3@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "betguest3@example.com").ID),
		)

		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      60 * 60 * 24,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		// Rock vs Rock → tie.
		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveRock,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}

		hostBal, _ := ledger.GetUserBalance(ctx, *host.UserID)
		guestBal, _ := ledger.GetUserBalance(ctx, *guest.UserID)
		if hostBal != 500 {
			t.Errorf("host balance after tie = %d, want 500", hostBal)
		}
		if guestBal != 500 {
			t.Errorf("guest balance after tie = %d, want 500", guestBal)
		}
	})
}

func TestDbRpsGameService_Betting_GuestCancels_HostRefunded(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("bethost4@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "bethost4@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("betguest4@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "betguest4@example.com").ID),
		)

		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)

		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      60 * 60 * 24,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCancelled,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest() cancel error = %v", err)
		}

		// Host's pending hold should have been released.
		avail, _ := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if avail != 500 {
			t.Errorf("host available balance after cancel = %d, want 500", avail)
		}
	})
}

func TestDbRpsGameService_Betting_RequestGame_WithoutHostUserID_Fails(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("nouser@example.com"))
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("nouser2@example.com"))

		betAmount := int64(50)
		_, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      3600,
			BetAmount:            &betAmount,
			HostUserID:           nil, // missing — should fail
		})
		if err == nil {
			t.Fatal("expected error when BetAmount set but HostUserID is nil")
		}
	})
}

func TestDbRpsGameService_RequestGame_SelfPlay_Rejected(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("selfplay@example.com"))

		_, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   player.ID,
			InvitedPlayerID:      player.ID, // same player
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      3600,
		})
		if err == nil {
			t.Fatal("expected error when player challenges themselves, got nil")
		}
		if err.Error() != "cannot challenge yourself" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestDbRpsGameService_Betting_GuestInsufficientBalance_Rejected(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("bhost_insuf@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "bhost_insuf@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("bguest_insuf@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "bguest_insuf@example.com").ID),
		)

		// Only fund the host; guest has 0 balance.
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)

		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      60 * 60 * 24,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		// Guest tries to respond but has no funds.
		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveScissors,
		})
		if err == nil {
			t.Fatal("expected error for guest insufficient balance, got nil")
		}

		// Host's escrow should still be pending (game did not settle).
		hostAvail, _ := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if hostAvail != 400 {
			t.Errorf("host available balance = %d, want 400 (escrow still held)", hostAvail)
		}
	})
}

func TestDbRpsGameService_Betting_GuestBetTransferID_SavedAfterSettle(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("bhost_gid@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "bhost_gid@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("bguest_gid@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "bguest_gid@example.com").ID),
		)

		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      60 * 60 * 24,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		responded, err := rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveRock, // tie
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}

		if responded.RpsGame.GuestBetTransferID == nil {
			t.Error("expected GuestBetTransferID to be set after bet settlement")
		}
	})
}

func TestDbRpsGameService_ExpireGamesAndRefundBets_RefundsHostEscrow(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("expire_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "expire_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("expire_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "expire_guest@example.com").ID),
		)

		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)

		// Create a game that expires in 1 second.
		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      1,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}
		if game.RpsGame.HostBetTransferID == nil {
			t.Fatal("expected HostBetTransferID to be set")
		}

		// Confirm escrow is held.
		availBefore, _ := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if availBefore != 400 {
			t.Fatalf("available balance before expiry = %d, want 400", availBefore)
		}

		// Wait for game to expire.
		time.Sleep(2 * time.Second)

		// Run expiry sweep.
		processed, err := rpsService.ExpireGamesAndRefundBets(ctx)
		if err != nil {
			t.Fatalf("ExpireGamesAndRefundBets() error = %v", err)
		}
		if processed != 1 {
			t.Errorf("processed = %d, want 1", processed)
		}

		// Escrow should be released.
		availAfter, _ := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if availAfter != 500 {
			t.Errorf("available balance after sweep = %d, want 500 (escrow released)", availAfter)
		}

		// Game should be cancelled.
		updatedGame, _ := adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{
			Ids: []uuid.UUID{game.RpsGame.ID},
		})
		if updatedGame.Status != models.RpsGameStatusCancelled {
			t.Errorf("game status = %v, want cancelled", updatedGame.Status)
		}
	})
}

func TestDbRpsGameService_ExpireGamesAndRefundBets_IgnoresNoBetGames(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		player1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("nobet_p1@example.com"))
		player2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("nobet_p2@example.com"))

		// Game with no bet, expired.
		_, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   player1.ID,
			InvitedPlayerID:      player2.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      1,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		time.Sleep(2 * time.Second)

		processed, err := rpsService.ExpireGamesAndRefundBets(ctx)
		if err != nil {
			t.Fatalf("ExpireGamesAndRefundBets() error = %v", err)
		}
		// No-bet games are not touched by the sweep.
		if processed != 0 {
			t.Errorf("processed = %d, want 0 (no-bet game should be ignored)", processed)
		}
	})
}

func mustCreateUser(t *testing.T, ctx context.Context, adapter stores.StorageAdapterInterface, email string) *models.User {
	t.Helper()
	user, err := adapter.User().CreateUser(ctx, &models.User{Email: email})
	if err != nil {
		t.Fatalf("mustCreateUser(%s): %v", email, err)
	}
	return user
}
