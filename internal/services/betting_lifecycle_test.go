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

func TestBetting_NoPendingTransfers_AfterCancel(t *testing.T) {
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
			DurationSeconds:      1,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		time.Sleep(2 * time.Second)

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
