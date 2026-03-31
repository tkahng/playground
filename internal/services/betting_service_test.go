package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// mustFundWallet credits a user wallet with amount points for test setup.
func mustFundWallet(t *testing.T, ctx context.Context, adapter stores.StorageAdapterInterface, ledger LedgerService, userID uuid.UUID, amount int64) {
	t.Helper()
	issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
	if err != nil {
		t.Fatalf("mustFundWallet: get issuance account: %v", err)
	}
	wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
	if err != nil {
		t.Fatalf("mustFundWallet: get wallet: %v", err)
	}
	_, err = ledger.PostTransfer(ctx, PostTransferInput{
		LedgerCode:      "POINTS",
		DebitAccountID:  issuance.ID,
		CreditAccountID: wallet.ID,
		Amount:          amount,
		TransferCode:    models.TransferCodePurchase,
	})
	if err != nil {
		t.Fatalf("mustFundWallet: post transfer: %v", err)
	}
}

// --- FulfillPointsPurchase tests ---

func TestFulfillPointsPurchase_CreditsUserWallet(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		err := FulfillPointsPurchase(ctx, adapter, ledger, PointsPurchaseFulfillInput{
			UserID:          userID,
			PointsAmount:    500,
			StripeSessionID: "cs_test_abc123",
		})
		if err != nil {
			t.Fatalf("FulfillPointsPurchase() error = %v", err)
		}

		balance, err := ledger.GetUserBalance(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserBalance() error = %v", err)
		}
		if balance != 500 {
			t.Errorf("balance = %d, want 500", balance)
		}
	})
}

func TestFulfillPointsPurchase_Idempotent_DuplicateWebhook(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()
		sessionID := "cs_test_idempotent_xyz"

		// First delivery.
		if err := FulfillPointsPurchase(ctx, adapter, ledger, PointsPurchaseFulfillInput{
			UserID:          userID,
			PointsAmount:    200,
			StripeSessionID: sessionID,
		}); err != nil {
			t.Fatalf("first FulfillPointsPurchase() error = %v", err)
		}

		// Duplicate delivery with the same session ID.
		if err := FulfillPointsPurchase(ctx, adapter, ledger, PointsPurchaseFulfillInput{
			UserID:          userID,
			PointsAmount:    200,
			StripeSessionID: sessionID,
		}); err != nil {
			t.Fatalf("duplicate FulfillPointsPurchase() error = %v", err)
		}

		balance, err := ledger.GetUserBalance(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserBalance() error = %v", err)
		}
		// Should only have been credited once.
		if balance != 200 {
			t.Errorf("balance after duplicate = %d, want 200", balance)
		}
	})
}

func TestFulfillPointsPurchase_DifferentSessions_BothCredited(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		if err := FulfillPointsPurchase(ctx, adapter, ledger, PointsPurchaseFulfillInput{
			UserID:          userID,
			PointsAmount:    100,
			StripeSessionID: "cs_test_first",
		}); err != nil {
			t.Fatalf("first FulfillPointsPurchase() error = %v", err)
		}
		if err := FulfillPointsPurchase(ctx, adapter, ledger, PointsPurchaseFulfillInput{
			UserID:          userID,
			PointsAmount:    300,
			StripeSessionID: "cs_test_second",
		}); err != nil {
			t.Fatalf("second FulfillPointsPurchase() error = %v", err)
		}

		balance, err := ledger.GetUserBalance(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserBalance() error = %v", err)
		}
		if balance != 400 {
			t.Errorf("balance after two purchases = %d, want 400", balance)
		}
	})
}

// --- BettingService direct tests ---

func TestDbBettingService_PlaceHostBet_InsufficientBalance(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		hostUserID := uuid.New()
		gameID := uuid.New()

		// Wallet has 0 balance — bet should be rejected.
		_, err := betting.PlaceHostBet(ctx, gameID, hostUserID, 100)
		if err == nil {
			t.Fatal("expected error for insufficient balance, got nil")
		}
	})
}

func TestDbBettingService_PlaceHostBet_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		hostUserID := uuid.New()
		gameID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)

		pending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, 100)
		if err != nil {
			t.Fatalf("PlaceHostBet() error = %v", err)
		}
		if pending.Status != models.LedgerTransferStatusPending {
			t.Errorf("transfer status = %v, want pending", pending.Status)
		}

		// Available balance should be reduced by the hold.
		avail, err := ledger.GetUserAvailableBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance() error = %v", err)
		}
		if avail != 400 {
			t.Errorf("available balance = %d, want 400", avail)
		}
		// Settled balance unchanged.
		bal, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance() error = %v", err)
		}
		if bal != 500 {
			t.Errorf("settled balance = %d, want 500", bal)
		}
	})
}

func TestDbBettingService_PlaceGuestAndSettle_HostWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, 100)
		if err != nil {
			t.Fatalf("PlaceHostBet() error = %v", err)
		}

		err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             100,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultWin,
			GuestResult:           models.RpsParticipantResultLose,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle() error = %v", err)
		}

		hostBal, _ := ledger.GetUserBalance(ctx, hostUserID)
		guestBal, _ := ledger.GetUserBalance(ctx, guestUserID)
		if hostBal != 600 {
			t.Errorf("host balance = %d, want 600 (won pot of 200)", hostBal)
		}
		if guestBal != 400 {
			t.Errorf("guest balance = %d, want 400 (lost bet of 100)", guestBal)
		}
	})
}

func TestDbBettingService_PlaceGuestAndSettle_GuestWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, 100)
		if err != nil {
			t.Fatalf("PlaceHostBet() error = %v", err)
		}

		err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             100,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultLose,
			GuestResult:           models.RpsParticipantResultWin,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle() error = %v", err)
		}

		hostBal, _ := ledger.GetUserBalance(ctx, hostUserID)
		guestBal, _ := ledger.GetUserBalance(ctx, guestUserID)
		if hostBal != 400 {
			t.Errorf("host balance = %d, want 400", hostBal)
		}
		if guestBal != 600 {
			t.Errorf("guest balance = %d, want 600", guestBal)
		}
	})
}

func TestDbBettingService_PlaceGuestAndSettle_Tie(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, 100)
		if err != nil {
			t.Fatalf("PlaceHostBet() error = %v", err)
		}

		err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             100,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultTie,
			GuestResult:           models.RpsParticipantResultTie,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle() tie error = %v", err)
		}

		hostBal, _ := ledger.GetUserBalance(ctx, hostUserID)
		guestBal, _ := ledger.GetUserBalance(ctx, guestUserID)
		if hostBal != 500 {
			t.Errorf("host balance = %d, want 500 (refunded)", hostBal)
		}
		if guestBal != 500 {
			t.Errorf("guest balance = %d, want 500 (refunded)", guestBal)
		}
	})
}

func TestDbBettingService_RefundHostBet_VoidsPendingHold(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		gameID := uuid.New()
		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, 100)
		if err != nil {
			t.Fatalf("PlaceHostBet() error = %v", err)
		}

		if err := betting.RefundHostBet(ctx, hostPending.ID); err != nil {
			t.Fatalf("RefundHostBet() error = %v", err)
		}

		// Hold released — available balance is back to 500.
		avail, _ := ledger.GetUserAvailableBalance(ctx, hostUserID)
		if avail != 500 {
			t.Errorf("available balance after refund = %d, want 500", avail)
		}
	})
}
