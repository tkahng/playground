package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestBettingService_PotConservation_HostWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		escrowBefore := escrow.Balance()

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, 100)
		if err != nil {
			t.Fatalf("PlaceHostBet() error = %v", err)
		}

		_, err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
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

		hostBal, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance(host): %v", err)
		}
		guestBal, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance(guest): %v", err)
		}
		if hostBal != 600 {
			t.Errorf("host balance = %d, want 600", hostBal)
		}
		if guestBal != 400 {
			t.Errorf("guest balance = %d, want 400", guestBal)
		}

		escrow, err = ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("re-fetch escrow: %v", err)
		}
		if escrow.Balance() != escrowBefore {
			t.Errorf("escrow balance = %d, want %d", escrow.Balance(), escrowBefore)
		}
	})
}

func TestBettingService_PotConservation_GuestWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		escrowBefore := escrow.Balance()

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, 100)
		if err != nil {
			t.Fatalf("PlaceHostBet() error = %v", err)
		}

		_, err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
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

		hostBal, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance(host): %v", err)
		}
		guestBal, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance(guest): %v", err)
		}
		if hostBal != 400 {
			t.Errorf("host balance = %d, want 400", hostBal)
		}
		if guestBal != 600 {
			t.Errorf("guest balance = %d, want 600", guestBal)
		}

		escrow, err = ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("re-fetch escrow: %v", err)
		}
		if escrow.Balance() != escrowBefore {
			t.Errorf("escrow balance = %d, want %d", escrow.Balance(), escrowBefore)
		}
	})
}

func TestBettingService_PotConservation_Tie(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		escrowBefore := escrow.Balance()

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, 100)
		if err != nil {
			t.Fatalf("PlaceHostBet() error = %v", err)
		}

		_, err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             100,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultTie,
			GuestResult:           models.RpsParticipantResultTie,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle() error = %v", err)
		}

		hostBal, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance(host): %v", err)
		}
		guestBal, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance(guest): %v", err)
		}
		if hostBal != 500 {
			t.Errorf("host balance = %d, want 500", hostBal)
		}
		if guestBal != 500 {
			t.Errorf("guest balance = %d, want 500", guestBal)
		}

		escrow, err = ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("re-fetch escrow: %v", err)
		}
		if escrow.Balance() != escrowBefore {
			t.Errorf("escrow balance = %d, want %d", escrow.Balance(), escrowBefore)
		}
	})
}

func TestBettingService_EnsureGuestCanAffordBet_UsesAvailableBalance(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		guestUserID := uuid.New()

		// Fund guest with 100 points.
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 100)

		// Create an 80-point pending hold on the guest wallet (debit guest → escrow).
		guestWallet, err := ledger.GetOrCreateUserWallet(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		_, err = ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  guestWallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          80,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		// Available balance is now 20 (100 - 80 pending hold).
		// Asking for 30 should fail.
		err = betting.EnsureGuestCanAffordBet(ctx, guestUserID, 30)
		if err == nil {
			t.Fatal("expected error from EnsureGuestCanAffordBet when available=20 and need=30, got nil")
		}
	})
}

func TestBettingService_PlaceGuestAndSettle_RejectsDoubleSettlement(t *testing.T) {

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

		input := PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             100,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultWin,
			GuestResult:           models.RpsParticipantResultLose,
		}

		// First call should succeed.
		_, err = betting.PlaceGuestAndSettle(ctx, input)
		if err != nil {
			t.Fatalf("first PlaceGuestAndSettle() error = %v", err)
		}

		// Second call with same hostPending.ID should be rejected.
		_, err = betting.PlaceGuestAndSettle(ctx, input)
		if err == nil {
			t.Fatal("expected error on double-settlement, got nil")
		}
	})
}
