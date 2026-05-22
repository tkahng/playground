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

func TestLedgerService_PostTransfer_RejectsOverdraft(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 100)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		// Attempt to debit 101 from a wallet that only has 100.
		_, err = ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: issuance.ID,
			Amount:          101,
			TransferCode:    models.TransferCodePurchase,
		})
		if err == nil {
			t.Fatal("PostTransfer overdraft: want error, got nil")
		}
		if !strings.Contains(err.Error(), "insufficient balance") {
			t.Errorf("PostTransfer overdraft error = %q, want to contain \"insufficient balance\"", err.Error())
		}
	})
}

func TestLedgerService_CreatePendingTransfer_RejectsWhenAvailableBalanceInsufficient(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 100)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		// Place an 80-point pending hold.
		_, err = ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          80,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer 80: %v", err)
		}

		// Available is now 20. Attempt a 30-point pending hold.
		_, err = ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          30,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err == nil {
			t.Fatal("CreatePendingTransfer over available balance: want error, got nil")
		}
		if !strings.Contains(err.Error(), "insufficient available balance") {
			t.Errorf("error = %q, want to contain \"insufficient available balance\"", err.Error())
		}
	})
}

func TestLedgerService_PostTransfer_RejectsZeroAmount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		wallet, err := ledger.GetOrCreateUserWallet(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}

		_, err = ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  issuance.ID,
			CreditAccountID: wallet.ID,
			Amount:          0,
			TransferCode:    models.TransferCodePurchase,
		})
		if err == nil {
			t.Fatal("PostTransfer amount=0: want error, got nil")
		}
		if !strings.Contains(err.Error(), "must be positive") {
			t.Errorf("error = %q, want to contain \"must be positive\"", err.Error())
		}
	})
}

func TestLedgerService_PostTransfer_RejectsNegativeAmount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		wallet, err := ledger.GetOrCreateUserWallet(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}

		_, err = ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  issuance.ID,
			CreditAccountID: wallet.ID,
			Amount:          -1,
			TransferCode:    models.TransferCodePurchase,
		})
		if err == nil {
			t.Fatal("PostTransfer amount=-1: want error, got nil")
		}
		if !strings.Contains(err.Error(), "must be positive") {
			t.Errorf("error = %q, want to contain \"must be positive\"", err.Error())
		}
	})
}

func TestLedgerService_CreatePendingTransfer_RejectsZeroAmount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		userID := uuid.New()
		mustFundWallet(t, ctx, adapter, ledger, userID, 100)
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
			Amount:          0,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err == nil {
			t.Fatal("CreatePendingTransfer amount=0: want error, got nil")
		}
		if !strings.Contains(err.Error(), "must be positive") {
			t.Errorf("error = %q, want to contain \"must be positive\"", err.Error())
		}
	})
}

// TestLedgerService_DbConstraint_RejectsOverdraft verifies the DB-level CHECK constraint
// on ledger.accounts fires when a direct balance update would overdraft a wallet account,
// bypassing the application-layer checkDebitConstraint.
func TestLedgerService_DbConstraint_RejectsOverdraft(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 100)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}

		// Directly increment debits_posted by 101 — bypassing app-layer guard.
		// The DB CHECK constraint must reject this.
		_, err = adapter.Ledger().AtomicUpdateAccountBalances(ctx, wallet.ID, 101, 0, 0, 0)
		if err == nil {
			t.Fatal("AtomicUpdateAccountBalances overdraft: want DB constraint error, got nil")
		}
	})
}

// TestLedgerService_IssuanceAccount_CreditsMustNotExceedDebits_IsNotEnforced documents
// that the AccountConstraintCreditsMustNotExceedDebits constraint on the issuance account
// is never enforced by checkDebitConstraint (which only checks DebitsMustNotExceedCredits).
// The issuance account starts with CreditsPosted=0, so if enforced, no points could ever
// be issued without first crediting the issuance account.
// Remove the t.Skip when the enforcement bug is fixed.
func TestLedgerService_IssuanceAccount_CreditsMustNotExceedDebits_IsNotEnforced(t *testing.T) {
	t.Skip("known bug: AccountConstraintCreditsMustNotExceedDebits is not enforced in checkDebitConstraint")

	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		// Verify the constraint is present.
		found := false
		for _, c := range issuance.Constraints {
			if c == models.AccountConstraintCreditsMustNotExceedDebits {
				found = true
			}
		}
		if !found {
			t.Fatalf("issuance account missing CreditsMustNotExceedDebits constraint; test premise invalid")
		}

		wallet, err := ledger.GetOrCreateUserWallet(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}

		// Issuance account has CreditsPosted=0. If enforced, debiting it should fail.
		_, err = ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  issuance.ID,
			CreditAccountID: wallet.ID,
			Amount:          100,
			TransferCode:    models.TransferCodePurchase,
		})
		if err == nil {
			t.Error("PostTransfer from unconstrained issuance: want error (credits_must_not_exceed_debits), got nil — constraint is not enforced")
		}
	})
}
