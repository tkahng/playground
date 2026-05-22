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

func TestDbLedgerService_FindTransfers_ReturnsMatchingTransfers(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		// Fund the wallet so we have a posted transfer.
		if err := FulfillPointsPurchase(ctx, adapter, ledger, PointsPurchaseFulfillInput{
			UserID:          userID,
			PointsAmount:    300,
			StripeSessionID: "cs_test_find_transfers",
		}); err != nil {
			t.Fatalf("FulfillPointsPurchase() error = %v", err)
		}

		wallet, err := ledger.GetUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserWallet() error = %v", err)
		}

		transfers, err := ledger.FindTransfers(ctx, &stores.LedgerTransferFilter{
			AccountIds:    []uuid.UUID{wallet.ID},
			TransferCodes: []string{models.TransferCodePurchase},
		})
		if err != nil {
			t.Fatalf("FindTransfers() error = %v", err)
		}
		if len(transfers) != 1 {
			t.Errorf("FindTransfers() len = %d, want 1", len(transfers))
		}
		if transfers[0].TransferCode != models.TransferCodePurchase {
			t.Errorf("FindTransfers()[0].TransferCode = %q, want %q", transfers[0].TransferCode, models.TransferCodePurchase)
		}
	})
}

func TestDbLedgerService_FindTransfers_EmptyForUnknownAccount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		transfers, err := ledger.FindTransfers(ctx, &stores.LedgerTransferFilter{
			AccountIds: []uuid.UUID{uuid.New()},
		})
		if err != nil {
			t.Fatalf("FindTransfers() error = %v", err)
		}
		if len(transfers) != 0 {
			t.Errorf("FindTransfers() len = %d, want 0", len(transfers))
		}
	})
}

func TestDbLedgerService_CountTransfers_ReturnsCorrectCount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		// Create 3 purchase transfers via idempotent sessions.
		for _, sessionID := range []string{"cs_count_1", "cs_count_2", "cs_count_3"} {
			if err := FulfillPointsPurchase(ctx, adapter, ledger, PointsPurchaseFulfillInput{
				UserID:          userID,
				PointsAmount:    100,
				StripeSessionID: sessionID,
			}); err != nil {
				t.Fatalf("FulfillPointsPurchase(%s) error = %v", sessionID, err)
			}
		}

		wallet, err := ledger.GetUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserWallet() error = %v", err)
		}

		count, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			AccountIds: []uuid.UUID{wallet.ID},
		})
		if err != nil {
			t.Fatalf("CountTransfers() error = %v", err)
		}
		if count != 3 {
			t.Errorf("CountTransfers() = %d, want 3", count)
		}

		// Filter by a code that produced no transfers.
		count, err = ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			AccountIds:    []uuid.UUID{wallet.ID},
			TransferCodes: []string{models.TransferCodeBetWin},
		})
		if err != nil {
			t.Fatalf("CountTransfers() by bet_win error = %v", err)
		}
		if count != 0 {
			t.Errorf("CountTransfers() by bet_win = %d, want 0", count)
		}
	})
}
