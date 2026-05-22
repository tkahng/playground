//go:build integration

package services

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// setupPendingTransfer funds a wallet and creates a pending transfer of 50 points
// from the user wallet to the game escrow account. Returns the pending transfer and userID.
func setupPendingTransfer(t *testing.T, ctx context.Context, adapter stores.StorageAdapterInterface, ledger LedgerService) (*models.LedgerTransfer, uuid.UUID) {
	t.Helper()
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
	return pending, userID
}

func TestLedgerService_PostPendingTransfer_AlreadyPosted_ReturnsError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		pending, userID := setupPendingTransfer(t, ctx, adapter, ledger)

		// First post: should succeed
		_, err := ledger.PostPendingTransfer(ctx, pending.ID)
		if err != nil {
			t.Fatalf("PostPendingTransfer (first): %v", err)
		}

		// Second post: must return error
		_, err = ledger.PostPendingTransfer(ctx, pending.ID)
		if err == nil {
			t.Fatal("PostPendingTransfer (second): expected error, got nil")
		}
		if !strings.Contains(err.Error(), "pending") && !strings.Contains(err.Error(), "posted") && !strings.Contains(err.Error(), "status") && !strings.Contains(err.Error(), "already") && !strings.Contains(err.Error(), "terminal") && !strings.Contains(err.Error(), "invalid") {
			t.Logf("PostPendingTransfer (second) error (acceptable): %v", err)
		}

		// Final balance assertions: 200 - 50 = 150
		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 150 {
			t.Errorf("Balance() = %d, want 150", wallet.Balance())
		}
		if wallet.AvailableBalance() != 150 {
			t.Errorf("AvailableBalance() = %d, want 150", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_PostPendingTransfer_AlreadyVoided_ReturnsError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		pending, userID := setupPendingTransfer(t, ctx, adapter, ledger)

		// Void first: should succeed
		_, err := ledger.VoidPendingTransfer(ctx, pending.ID)
		if err != nil {
			t.Fatalf("VoidPendingTransfer: %v", err)
		}

		// Post after void: must return error
		_, err = ledger.PostPendingTransfer(ctx, pending.ID)
		if err == nil {
			t.Fatal("PostPendingTransfer after void: expected error, got nil")
		}
		if !strings.Contains(err.Error(), "pending") && !strings.Contains(err.Error(), "voided") && !strings.Contains(err.Error(), "status") && !strings.Contains(err.Error(), "already") && !strings.Contains(err.Error(), "terminal") && !strings.Contains(err.Error(), "invalid") {
			t.Logf("PostPendingTransfer after void error (acceptable): %v", err)
		}

		// Final balance assertions: void releases hold, so balance stays 200
		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 200 {
			t.Errorf("Balance() = %d, want 200", wallet.Balance())
		}
		if wallet.AvailableBalance() != 200 {
			t.Errorf("AvailableBalance() = %d, want 200", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_VoidPendingTransfer_AlreadyVoided_ReturnsError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		pending, userID := setupPendingTransfer(t, ctx, adapter, ledger)

		// First void: should succeed
		_, err := ledger.VoidPendingTransfer(ctx, pending.ID)
		if err != nil {
			t.Fatalf("VoidPendingTransfer (first): %v", err)
		}

		// Second void: must return error
		_, err = ledger.VoidPendingTransfer(ctx, pending.ID)
		if err == nil {
			t.Fatal("VoidPendingTransfer (second): expected error, got nil")
		}
		if !strings.Contains(err.Error(), "pending") && !strings.Contains(err.Error(), "voided") && !strings.Contains(err.Error(), "status") && !strings.Contains(err.Error(), "already") && !strings.Contains(err.Error(), "terminal") && !strings.Contains(err.Error(), "invalid") {
			t.Logf("VoidPendingTransfer (second) error (acceptable): %v", err)
		}

		// Final balance assertions: void releases hold, balance stays 200
		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 200 {
			t.Errorf("Balance() = %d, want 200", wallet.Balance())
		}
		if wallet.AvailableBalance() != 200 {
			t.Errorf("AvailableBalance() = %d, want 200", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_VoidPendingTransfer_AlreadyPosted_ReturnsError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		pending, userID := setupPendingTransfer(t, ctx, adapter, ledger)

		// Post first: should succeed
		_, err := ledger.PostPendingTransfer(ctx, pending.ID)
		if err != nil {
			t.Fatalf("PostPendingTransfer: %v", err)
		}

		// Void after post: must return error
		_, err = ledger.VoidPendingTransfer(ctx, pending.ID)
		if err == nil {
			t.Fatal("VoidPendingTransfer after post: expected error, got nil")
		}
		if !strings.Contains(err.Error(), "pending") && !strings.Contains(err.Error(), "posted") && !strings.Contains(err.Error(), "status") && !strings.Contains(err.Error(), "already") && !strings.Contains(err.Error(), "terminal") && !strings.Contains(err.Error(), "invalid") {
			t.Logf("VoidPendingTransfer after post error (acceptable): %v", err)
		}

		// Final balance assertions: post committed the 50-point debit, balance = 150
		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 150 {
			t.Errorf("Balance() = %d, want 150", wallet.Balance())
		}
		if wallet.AvailableBalance() != 150 {
			t.Errorf("AvailableBalance() = %d, want 150", wallet.AvailableBalance())
		}
	})
}
