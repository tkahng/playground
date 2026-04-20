package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestLedgerService_AvailableBalance_DecreasesWithPendingHold(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 200)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		_, err = ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		// Re-fetch wallet to get updated counters.
		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 200 {
			t.Errorf("Balance() = %d, want 200", wallet.Balance())
		}
		if wallet.AvailableBalance() != 150 {
			t.Errorf("AvailableBalance() = %d, want 150", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_AvailableBalance_RestoresAfterVoid(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 200)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		pending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		if _, err = ledger.VoidPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("VoidPendingTransfer: %v", err)
		}

		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 200 {
			t.Errorf("Balance() = %d, want 200 after void", wallet.Balance())
		}
		if wallet.AvailableBalance() != 200 {
			t.Errorf("AvailableBalance() = %d, want 200 after void", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_AvailableBalance_DecreasesAfterPendingPost(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 200)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		pending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		if _, err = ledger.PostPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("PostPendingTransfer: %v", err)
		}

		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 150 {
			t.Errorf("Balance() = %d, want 150 after post", wallet.Balance())
		}
		if wallet.AvailableBalance() != 150 {
			t.Errorf("AvailableBalance() = %d, want 150 after post", wallet.AvailableBalance())
		}
	})
}
