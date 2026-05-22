//go:build integration

package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestBetting_NoPendingTransfers_AfterCancel(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("cancel_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "cancel_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("cancel_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "cancel_guest@example.com").ID),
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
		if game.RpsGame.HostBetTransferID == nil {
			t.Fatal("expected HostBetTransferID to be set after RequestGame with bet")
		}

		availMid, err := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance (mid): %v", err)
		}
		if availMid != 400 {
			t.Errorf("host available balance after bet escrow = %d, want 400", availMid)
		}

		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCancelled,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest(cancel) error = %v", err)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after cancel = %d, want 0", pendingCount)
		}

		avail, err := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance: %v", err)
		}
		if avail != 500 {
			t.Errorf("host available balance = %d, want 500", avail)
		}
	})
}

func TestBetting_NoPendingTransfers_AfterExpiry(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("expiry_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "expiry_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("expiry_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "expiry_guest@example.com").ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)

		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      0,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}
		if game.RpsGame.HostBetTransferID == nil {
			t.Fatal("expected HostBetTransferID to be set after RequestGame with bet")
		}

		processed, err := rpsService.ExpireGamesAndRefundBets(ctx)
		if err != nil {
			t.Fatalf("ExpireGamesAndRefundBets() error = %v", err)
		}
		if processed != 1 {
			t.Errorf("processed = %d, want 1", processed)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after expiry = %d, want 0", pendingCount)
		}

		avail, err := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance: %v", err)
		}
		if avail != 500 {
			t.Errorf("host available balance after sweep = %d, want 500", avail)
		}

		updated, err := adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{Ids: []uuid.UUID{game.RpsGame.ID}})
		if err != nil {
			t.Fatalf("FindRpsGame: %v", err)
		}
		if updated.Status != models.RpsGameStatusCancelled {
			t.Errorf("game status = %v, want cancelled", updated.Status)
		}
	})
}

func TestBetting_NoPendingTransfers_AfterComplete_HostWins(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("hwins_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "hwins_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("hwins_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "hwins_guest@example.com").ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		betAmount := int64(100)
		// Host plays Rock; guest plays Scissors → host wins.
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

		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveScissors,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after host win = %d, want 0", pendingCount)
		}

		hostBal, err := ledger.GetUserBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBal, err := ledger.GetUserBalance(ctx, *guest.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}
		if hostBal != 600 {
			t.Errorf("host balance = %d, want 600", hostBal)
		}
		if guestBal != 400 {
			t.Errorf("guest balance = %d, want 400", guestBal)
		}
	})
}

func TestBetting_NoPendingTransfers_AfterComplete_GuestWins(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("gwins_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "gwins_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("gwins_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "gwins_guest@example.com").ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		betAmount := int64(100)
		// Host plays Rock; guest plays Paper → guest wins.
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

		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMovePaper,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after guest win = %d, want 0", pendingCount)
		}

		hostBal, err := ledger.GetUserBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBal, err := ledger.GetUserBalance(ctx, *guest.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}
		if hostBal != 400 {
			t.Errorf("host balance = %d, want 400", hostBal)
		}
		if guestBal != 600 {
			t.Errorf("guest balance = %d, want 600", guestBal)
		}
	})
}

func TestBetting_NoPendingTransfers_AfterComplete_Tie(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("tie_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "tie_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("tie_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "tie_guest@example.com").ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		betAmount := int64(100)
		// Rock vs Rock → tie.
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

		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveRock,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after tie = %d, want 0", pendingCount)
		}

		hostBal, err := ledger.GetUserBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBal, err := ledger.GetUserBalance(ctx, *guest.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}
		if hostBal != 500 {
			t.Errorf("host balance after tie = %d, want 500", hostBal)
		}
		if guestBal != 500 {
			t.Errorf("guest balance after tie = %d, want 500", guestBal)
		}
	})
}

func TestBetting_GuestCanRetry_AfterInsufficientFunds(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("retry_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "retry_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("retry_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "retry_guest@example.com").ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		// Guest starts with 0 balance — no funding here.

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

		// First attempt: must fail — guest has no funds.
		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveScissors,
		})
		if err == nil {
			t.Fatal("expected error when guest has no funds, got nil")
		}

		// Game must still be pending: RespondToGameRequest checks balance before calling
		// updateGame, so a balance error aborts before any state is persisted.
		currentGame, err := adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{Ids: []uuid.UUID{game.RpsGame.ID}})
		if err != nil {
			t.Fatalf("FindRpsGame: %v", err)
		}
		if currentGame.Status != models.RpsGameStatusPending {
			t.Errorf("game status after failed response = %v, want pending", currentGame.Status)
		}

		// Host escrow is still held.
		hostAvail, err := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance: %v", err)
		}
		if hostAvail != 400 {
			t.Errorf("host available after failed attempt = %d, want 400 (escrow still held)", hostAvail)
		}

		// Fund the guest, then retry — must succeed.
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		// Host plays Rock, guest plays Scissors → host wins.
		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveScissors,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest after funding guest: %v", err)
		}

		finalGame, err := adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{Ids: []uuid.UUID{game.RpsGame.ID}})
		if err != nil {
			t.Fatalf("FindRpsGame after retry: %v", err)
		}
		if finalGame.Status != models.RpsGameStatusCompleted {
			t.Errorf("game status after retry = %v, want completed", finalGame.Status)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after retry success = %d, want 0", pendingCount)
		}

		// Host wins: 500 - 100 (escrow posted) + 200 (pot) = 600
		hostBal, err := ledger.GetUserBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBal, err := ledger.GetUserBalance(ctx, *guest.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}
		if hostBal != 600 {
			t.Errorf("host balance = %d, want 600", hostBal)
		}
		if guestBal != 400 {
			t.Errorf("guest balance = %d, want 400 (500 funded - 100 bet)", guestBal)
		}
	})
}
