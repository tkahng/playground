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

func TestDbRpsGameService_ExpireGamesAndRefundBets_RefundsBothEscrows(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("both_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "both_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("both_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "both_guest@example.com").ID),
		)

		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

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

		// Manually create a guest pending escrow (normally impossible on a pending game)
		// and inject it onto the game row to simulate the edge-case state.
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount() error = %v", err)
		}
		guestWallet, err := ledger.GetOrCreateUserWallet(ctx, *guest.UserID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet() error = %v", err)
		}
		refType := models.ReferenceTypeRpsGame
		guestPending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  guestWallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          betAmount,
			TransferCode:    models.TransferCodeBetEscrow,
			ReferenceType:   &refType,
			ReferenceID:     &game.RpsGame.ID,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer(guest) error = %v", err)
		}

		// Inject GuestBetTransferID onto the game row directly.
		game.RpsGame.GuestBetTransferID = &guestPending.ID
		if _, err := adapter.Gaming().UpdateRpsGame(ctx, game.RpsGame); err != nil {
			t.Fatalf("UpdateRpsGame() inject guest transfer error = %v", err)
		}

		// Confirm both escrows are held.
		hostAvail, _ := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		guestAvail, _ := ledger.GetUserAvailableBalance(ctx, *guest.UserID)
		if hostAvail != 400 {
			t.Fatalf("host available before sweep = %d, want 400", hostAvail)
		}
		if guestAvail != 400 {
			t.Fatalf("guest available before sweep = %d, want 400", guestAvail)
		}

		time.Sleep(2 * time.Second)

		processed, err := rpsService.ExpireGamesAndRefundBets(ctx)
		if err != nil {
			t.Fatalf("ExpireGamesAndRefundBets() error = %v", err)
		}
		if processed != 1 {
			t.Errorf("processed = %d, want 1", processed)
		}

		// Both escrows should be released.
		hostAvail, _ = ledger.GetUserAvailableBalance(ctx, *host.UserID)
		guestAvail, _ = ledger.GetUserAvailableBalance(ctx, *guest.UserID)
		if hostAvail != 500 {
			t.Errorf("host available after sweep = %d, want 500", hostAvail)
		}
		if guestAvail != 500 {
			t.Errorf("guest available after sweep = %d, want 500", guestAvail)
		}

		// Game should be cancelled.
		updated, _ := adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{
			Ids: []uuid.UUID{game.RpsGame.ID},
		})
		if updated.Status != models.RpsGameStatusCancelled {
			t.Errorf("game status = %v, want cancelled", updated.Status)
		}
	})
}

func TestDbRpsGameService_ExpireGamesAndRefundBets_ContinuesOnPerGameError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		hostA := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("sweep_hosta@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "sweep_hosta@example.com").ID),
		)
		hostB := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("sweep_hostb@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "sweep_hostb@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("sweep_guest@example.com"),
		)

		mustFundPlayerWallet(t, ctx, adapter, ledger, hostA.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, hostB.UserID, 500)

		betAmount := int64(100)

		gameA, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   hostA.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      1,
			BetAmount:            &betAmount,
			HostUserID:           hostA.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame A error = %v", err)
		}

		gameB, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   hostB.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      1,
			BetAmount:            &betAmount,
			HostUserID:           hostB.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame B error = %v", err)
		}

		// Pre-void game A's escrow to force a per-game sweep failure.
		if err := betting.RefundHostBet(ctx, *gameA.RpsGame.HostBetTransferID); err != nil {
			t.Fatalf("RefundHostBet (pre-void) error = %v", err)
		}

		time.Sleep(2 * time.Second)

		processed, sweepErr := rpsService.ExpireGamesAndRefundBets(ctx)
		if sweepErr == nil {
			t.Error("expected error from failed game A, got nil")
		}
		if processed != 1 {
			t.Errorf("processed = %d, want 1 (only game B should succeed)", processed)
		}

		// Game A: sweep tx rolled back — still pending.
		updatedA, _ := adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{
			Ids: []uuid.UUID{gameA.RpsGame.ID},
		})
		if updatedA.Status != models.RpsGameStatusPending {
			t.Errorf("game A status = %v, want pending (sweep error should roll back)", updatedA.Status)
		}

		// Game B: swept successfully — cancelled, escrow released.
		updatedB, _ := adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{
			Ids: []uuid.UUID{gameB.RpsGame.ID},
		})
		if updatedB.Status != models.RpsGameStatusCancelled {
			t.Errorf("game B status = %v, want cancelled", updatedB.Status)
		}
		availB, _ := ledger.GetUserAvailableBalance(ctx, *hostB.UserID)
		if availB != 500 {
			t.Errorf("host B available balance = %d, want 500 (escrow released)", availB)
		}
	})
}

func TestDbRpsGameService_CreatePlayerByParams_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		user := mustCreateUser(t, ctx, adapter, "newplayer@example.com")
		player, err := rpsService.CreatePlayerByParams(ctx, &PlayerFindParams{
			Email:  "newplayer@example.com",
			UserID: &user.ID,
		})
		if err != nil {
			t.Fatalf("CreatePlayerByParams() error = %v", err)
		}
		if player == nil {
			t.Fatal("CreatePlayerByParams() returned nil player")
		}
		if player.Email != "newplayer@example.com" {
			t.Errorf("player.Email = %q, want %q", player.Email, "newplayer@example.com")
		}
		if player.UserID == nil || *player.UserID != user.ID {
			t.Errorf("player.UserID = %v, want %v", player.UserID, user.ID)
		}
	})
}

func TestDbRpsGameService_FindPlayerByParams_Found(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		// Create a player first.
		stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("findme@example.com"))

		found, err := rpsService.FindPlayerByParams(ctx, &PlayerFindParams{Email: "findme@example.com"})
		if err != nil {
			t.Fatalf("FindPlayerByParams() error = %v", err)
		}
		if found == nil {
			t.Fatal("FindPlayerByParams() returned nil, want player")
		}
		if found.Email != "findme@example.com" {
			t.Errorf("found.Email = %q, want %q", found.Email, "findme@example.com")
		}
	})
}

func TestDbRpsGameService_FindPlayerByParams_NotFound(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		found, err := rpsService.FindPlayerByParams(ctx, &PlayerFindParams{Email: "nobody@example.com"})
		if err != nil {
			t.Fatalf("FindPlayerByParams() error = %v", err)
		}
		if found != nil {
			t.Errorf("FindPlayerByParams() returned %v, want nil", found)
		}
	})
}

func TestDbRpsGameService_PlayerCanPlayWithPlayer_DeclinedFriendship(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		p1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("decline_p1@example.com"))
		p2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("decline_p2@example.com"))

		// MustCreateFriendship(store, invitedPlayer=p2, requestingPlayer=p1)
		// PlayerCanPlayWithPlayer filters by RequestingPlayerIds=[p1], InvitedPlayerIds=[p2]
		stores.MustCreateFriendship(t, ctx, adapter.Gaming(), p2, p1, stores.WithStatus(models.FriendshipStatusDeclined))

		canPlay, err := rpsService.PlayerCanPlayWithPlayer(ctx, p1.ID, p2.ID)
		if err != nil {
			t.Fatalf("PlayerCanPlayWithPlayer() error = %v", err)
		}
		if canPlay {
			t.Error("PlayerCanPlayWithPlayer() = true, want false for declined friendship")
		}
	})
}

func TestDbRpsGameService_PlayerCanPlayWithPlayer_AcceptedFriendship(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		p1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("accept_p1@example.com"))
		p2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("accept_p2@example.com"))

		stores.MustCreateFriendship(t, ctx, adapter.Gaming(), p2, p1, stores.WithStatus(models.FriendshipStatusAccepted))

		canPlay, err := rpsService.PlayerCanPlayWithPlayer(ctx, p1.ID, p2.ID)
		if err != nil {
			t.Fatalf("PlayerCanPlayWithPlayer() error = %v", err)
		}
		if !canPlay {
			t.Error("PlayerCanPlayWithPlayer() = false, want true for accepted friendship")
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
