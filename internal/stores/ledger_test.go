package stores

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

// makeIssuanceAccount creates a system issuance account (credits only, no overdraft constraint).
func makeIssuanceAccount(t *testing.T, ctx context.Context, s *DBLedgerStore) *models.LedgerAccount {
	t.Helper()
	acc, err := s.CreateAccount(ctx, &models.LedgerAccount{
		Code:       "system:issuance:" + uuid.NewString(),
		EntityType: "system",
		LedgerCode: "POINTS",
		Flags:      0,
		Metadata:   []byte("{}"),
	})
	if err != nil {
		t.Fatalf("makeIssuanceAccount: %v", err)
	}
	return acc
}

// makeWalletAccount creates a user wallet account (no-overdraft constraint).
func makeWalletAccount(t *testing.T, ctx context.Context, s *DBLedgerStore) *models.LedgerAccount {
	t.Helper()
	userID := uuid.New()
	acc, err := s.CreateAccount(ctx, &models.LedgerAccount{
		Code:       models.UserWalletCode(userID),
		EntityType: "user",
		EntityID:   &userID,
		LedgerCode: "POINTS",
		Flags:      models.AccountFlagDebitsMustNotExceedCredits,
		Metadata:   []byte("{}"),
	})
	if err != nil {
		t.Fatalf("makeWalletAccount: %v", err)
	}
	return acc
}

// --- Account tests ---

func TestDBLedgerStore_CreateAccount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		userID := uuid.New()
		acc, err := s.CreateAccount(ctx, &models.LedgerAccount{
			Code:       models.UserWalletCode(userID),
			EntityType: "user",
			EntityID:   &userID,
			LedgerCode: "POINTS",
			Flags:      models.AccountFlagDebitsMustNotExceedCredits,
			Metadata:   []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateAccount() error = %v", err)
		}
		if acc.ID == uuid.Nil {
			t.Error("CreateAccount() returned nil ID")
		}
		if acc.Balance() != 0 {
			t.Errorf("new account Balance() = %d, want 0", acc.Balance())
		}
		if acc.AvailableBalance() != 0 {
			t.Errorf("new account AvailableBalance() = %d, want 0", acc.AvailableBalance())
		}
	})
}

func TestDBLedgerStore_FindAccountByCode(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		acc := makeWalletAccount(t, ctx, s)

		found, err := s.FindAccountByCode(ctx, acc.Code)
		if err != nil {
			t.Fatalf("FindAccountByCode() error = %v", err)
		}
		if found == nil {
			t.Fatal("FindAccountByCode() returned nil")
		}
		if found.ID != acc.ID {
			t.Errorf("FindAccountByCode() ID = %v, want %v", found.ID, acc.ID)
		}

		missing, err := s.FindAccountByCode(ctx, "does-not-exist")
		if err != nil {
			t.Fatalf("FindAccountByCode(missing) error = %v", err)
		}
		if missing != nil {
			t.Error("FindAccountByCode(missing) expected nil, got account")
		}
	})
}

func TestDBLedgerStore_FindAccount_ByEntityID(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		acc := makeWalletAccount(t, ctx, s)

		found, err := s.FindAccount(ctx, &LedgerAccountFilter{
			EntityIds: []uuid.UUID{*acc.EntityID},
		})
		if err != nil {
			t.Fatalf("FindAccount() error = %v", err)
		}
		if found == nil || found.ID != acc.ID {
			t.Errorf("FindAccount() by entity_id = %v, want %v", found, acc.ID)
		}
	})
}

func TestDBLedgerStore_AtomicUpdateAccountBalances(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		acc := makeWalletAccount(t, ctx, s)

		// Credit 500 (simulate issuance).
		updated, err := s.AtomicUpdateAccountBalances(ctx, acc.ID, 0, 500, 0, 0)
		if err != nil {
			t.Fatalf("AtomicUpdateAccountBalances() credit error = %v", err)
		}
		if updated.Balance() != 500 {
			t.Errorf("Balance() after credit = %d, want 500", updated.Balance())
		}

		// Hold 100 as pending debit.
		updated, err = s.AtomicUpdateAccountBalances(ctx, acc.ID, 0, 0, 100, 0)
		if err != nil {
			t.Fatalf("AtomicUpdateAccountBalances() pending error = %v", err)
		}
		if updated.AvailableBalance() != 400 {
			t.Errorf("AvailableBalance() after pending = %d, want 400", updated.AvailableBalance())
		}

		// Post the debit.
		updated, err = s.AtomicUpdateAccountBalances(ctx, acc.ID, 100, 0, -100, 0)
		if err != nil {
			t.Fatalf("AtomicUpdateAccountBalances() post error = %v", err)
		}
		if updated.Balance() != 400 {
			t.Errorf("Balance() after post = %d, want 400", updated.Balance())
		}
		if updated.AvailableBalance() != 400 {
			t.Errorf("AvailableBalance() after post = %d, want 400", updated.AvailableBalance())
		}
	})
}

// --- Transfer tests ---

func TestDBLedgerStore_CreateTransfer_Posted(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		src := makeIssuanceAccount(t, ctx, s)
		dst := makeWalletAccount(t, ctx, s)

		ref := uuid.New()
		refType := models.ReferenceTypeStripeCheckout
		tr, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: dst.ID,
			Amount:          100,
			Status:          models.LedgerTransferStatusPosted,
			TransferCode:    models.TransferCodePurchase,
			ReferenceType:   &refType,
			ReferenceID:     &ref,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer() error = %v", err)
		}
		if tr.ID == uuid.Nil {
			t.Error("CreateTransfer() returned nil ID")
		}
		if tr.Status != models.LedgerTransferStatusPosted {
			t.Errorf("Status = %v, want posted", tr.Status)
		}
		if tr.Amount != 100 {
			t.Errorf("Amount = %d, want 100", tr.Amount)
		}
	})
}

func TestDBLedgerStore_CreateTransfer_Pending(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		src := makeWalletAccount(t, ctx, s)
		escrow := makeIssuanceAccount(t, ctx, s)

		tr, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			Flags:           models.TransferFlagPending,
			Status:          models.LedgerTransferStatusPending,
			TransferCode:    models.TransferCodeBetEscrow,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer() pending error = %v", err)
		}
		if tr.Status != models.LedgerTransferStatusPending {
			t.Errorf("Status = %v, want pending", tr.Status)
		}
	})
}

func TestDBLedgerStore_UpdateTransferStatus(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		src := makeWalletAccount(t, ctx, s)
		dst := makeIssuanceAccount(t, ctx, s)

		tr, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: dst.ID,
			Amount:          75,
			Flags:           models.TransferFlagPending,
			Status:          models.LedgerTransferStatusPending,
			TransferCode:    models.TransferCodeBetEscrow,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer() error = %v", err)
		}

		posted, err := s.UpdateTransferStatus(ctx, tr.ID, models.LedgerTransferStatusPosted)
		if err != nil {
			t.Fatalf("UpdateTransferStatus() posted error = %v", err)
		}
		if posted.Status != models.LedgerTransferStatusPosted {
			t.Errorf("Status after post = %v, want posted", posted.Status)
		}

		// Create another pending transfer to void.
		tr2, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: dst.ID,
			Amount:          25,
			Flags:           models.TransferFlagPending,
			Status:          models.LedgerTransferStatusPending,
			TransferCode:    models.TransferCodeBetEscrow,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer() 2 error = %v", err)
		}
		voided, err := s.UpdateTransferStatus(ctx, tr2.ID, models.LedgerTransferStatusVoided)
		if err != nil {
			t.Fatalf("UpdateTransferStatus() void error = %v", err)
		}
		if voided.Status != models.LedgerTransferStatusVoided {
			t.Errorf("Status after void = %v, want voided", voided.Status)
		}
	})
}

func TestDBLedgerStore_FindTransfers_ByStatus(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		src := makeIssuanceAccount(t, ctx, s)
		dst := makeWalletAccount(t, ctx, s)

		posted, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: dst.ID,
			Amount:          100,
			Status:          models.LedgerTransferStatusPosted,
			TransferCode:    models.TransferCodePurchase,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer posted: %v", err)
		}
		pending, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: dst.ID,
			Amount:          50,
			Flags:           models.TransferFlagPending,
			Status:          models.LedgerTransferStatusPending,
			TransferCode:    models.TransferCodeBetEscrow,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer pending: %v", err)
		}

		results, err := s.FindTransfers(ctx, &LedgerTransferFilter{
			Statuses: []models.LedgerTransferStatus{models.LedgerTransferStatusPosted},
			AccountIds: []uuid.UUID{dst.ID},
		})
		if err != nil {
			t.Fatalf("FindTransfers posted: %v", err)
		}
		if len(results) != 1 || results[0].ID != posted.ID {
			t.Errorf("FindTransfers by posted status: got %d results, want 1 with ID %v", len(results), posted.ID)
		}

		results, err = s.FindTransfers(ctx, &LedgerTransferFilter{
			Statuses: []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
			AccountIds: []uuid.UUID{dst.ID},
		})
		if err != nil {
			t.Fatalf("FindTransfers pending: %v", err)
		}
		if len(results) != 1 || results[0].ID != pending.ID {
			t.Errorf("FindTransfers by pending status: got %d results, want 1 with ID %v", len(results), pending.ID)
		}
	})
}

func TestDBLedgerStore_FindTransfers_ByTransferCode(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		src := makeIssuanceAccount(t, ctx, s)
		dst := makeWalletAccount(t, ctx, s)

		_, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: dst.ID,
			Amount:          200,
			Status:          models.LedgerTransferStatusPosted,
			TransferCode:    models.TransferCodePurchase,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer purchase: %v", err)
		}
		_, err = s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: dst.ID,
			Amount:          50,
			Status:          models.LedgerTransferStatusPosted,
			TransferCode:    models.TransferCodeBetWin,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer bet_win: %v", err)
		}

		results, err := s.FindTransfers(ctx, &LedgerTransferFilter{
			TransferCodes: []string{models.TransferCodePurchase},
			AccountIds:    []uuid.UUID{dst.ID},
		})
		if err != nil {
			t.Fatalf("FindTransfers by transfer_code: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("FindTransfers by transfer_code: got %d, want 1", len(results))
		}
		if results[0].TransferCode != models.TransferCodePurchase {
			t.Errorf("TransferCode = %v, want %v", results[0].TransferCode, models.TransferCodePurchase)
		}
	})
}

func TestDBLedgerStore_FindTransfers_ByReferenceID(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		src := makeIssuanceAccount(t, ctx, s)
		dst := makeWalletAccount(t, ctx, s)

		gameID := uuid.New()
		otherID := uuid.New()
		refType := models.ReferenceTypeRpsGame

		_, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: dst.ID,
			Amount:          100,
			Status:          models.LedgerTransferStatusPosted,
			TransferCode:    models.TransferCodeBetWin,
			ReferenceType:   &refType,
			ReferenceID:     &gameID,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer for gameID: %v", err)
		}
		_, err = s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: dst.ID,
			Amount:          50,
			Status:          models.LedgerTransferStatusPosted,
			TransferCode:    models.TransferCodeBetWin,
			ReferenceType:   &refType,
			ReferenceID:     &otherID,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer for otherID: %v", err)
		}

		results, err := s.FindTransfers(ctx, &LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{gameID},
		})
		if err != nil {
			t.Fatalf("FindTransfers by reference_id: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("FindTransfers by reference_id: got %d, want 1", len(results))
		}
		if *results[0].ReferenceID != gameID {
			t.Errorf("ReferenceID = %v, want %v", results[0].ReferenceID, gameID)
		}
	})
}

func TestDBLedgerStore_FindTransfers_ByAccountID(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		src := makeIssuanceAccount(t, ctx, s)
		walletA := makeWalletAccount(t, ctx, s)
		walletB := makeWalletAccount(t, ctx, s)

		_, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: walletA.ID,
			Amount:          100,
			Status:          models.LedgerTransferStatusPosted,
			TransferCode:    models.TransferCodePurchase,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer to walletA: %v", err)
		}
		_, err = s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: walletB.ID,
			Amount:          200,
			Status:          models.LedgerTransferStatusPosted,
			TransferCode:    models.TransferCodePurchase,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer to walletB: %v", err)
		}

		results, err := s.FindTransfers(ctx, &LedgerTransferFilter{
			AccountIds: []uuid.UUID{walletA.ID},
		})
		if err != nil {
			t.Fatalf("FindTransfers by account_id: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("FindTransfers by walletA: got %d, want 1", len(results))
		}
	})
}

func TestDBLedgerStore_CountTransfers(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		src := makeIssuanceAccount(t, ctx, s)
		dst := makeWalletAccount(t, ctx, s)

		for range 3 {
			_, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
				LedgerCode:      "POINTS",
				DebitAccountID:  src.ID,
				CreditAccountID: dst.ID,
				Amount:          10,
				Status:          models.LedgerTransferStatusPosted,
				TransferCode:    models.TransferCodePurchase,
				Metadata:        []byte("{}"),
			})
			if err != nil {
				t.Fatalf("CreateTransfer: %v", err)
			}
		}

		count, err := s.CountTransfers(ctx, &LedgerTransferFilter{
			AccountIds: []uuid.UUID{dst.ID},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if count != 3 {
			t.Errorf("CountTransfers = %d, want 3", count)
		}

		count, err = s.CountTransfers(ctx, &LedgerTransferFilter{
			AccountIds:    []uuid.UUID{dst.ID},
			TransferCodes: []string{models.TransferCodeBetWin},
		})
		if err != nil {
			t.Fatalf("CountTransfers by code: %v", err)
		}
		if count != 0 {
			t.Errorf("CountTransfers by bet_win = %d, want 0", count)
		}
	})
}

func TestDBLedgerStore_FindTransferByIDForUpdate(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		s := NewDBLedgerStore(db)
		src := makeIssuanceAccount(t, ctx, s)
		dst := makeWalletAccount(t, ctx, s)

		tr, err := s.CreateTransfer(ctx, &models.LedgerTransfer{
			LedgerCode:      "POINTS",
			DebitAccountID:  src.ID,
			CreditAccountID: dst.ID,
			Amount:          42,
			Status:          models.LedgerTransferStatusPending,
			TransferCode:    models.TransferCodeBetEscrow,
			Metadata:        []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateTransfer: %v", err)
		}

		locked, err := s.FindTransferByIDForUpdate(ctx, tr.ID)
		if err != nil {
			t.Fatalf("FindTransferByIDForUpdate: %v", err)
		}
		if locked == nil || locked.ID != tr.ID {
			t.Errorf("FindTransferByIDForUpdate ID = %v, want %v", locked, tr.ID)
		}

		missing, err := s.FindTransferByIDForUpdate(ctx, uuid.New())
		if err != nil {
			t.Fatalf("FindTransferByIDForUpdate missing: %v", err)
		}
		if missing != nil {
			t.Error("FindTransferByIDForUpdate(missing) expected nil")
		}
	})
}
