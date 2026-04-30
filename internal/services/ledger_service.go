package services

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// PostTransferInput holds the parameters for creating a ledger transfer.
type PostTransferInput struct {
	LedgerCode      string
	DebitAccountID  uuid.UUID
	CreditAccountID uuid.UUID
	Amount          int64
	TransferCode    string
	ReferenceType   *string
	ReferenceID     *uuid.UUID
	PendingID       *uuid.UUID
	Metadata        []byte
}

// LedgerService manages double-entry ledger operations.
// All balance-mutating methods must be called inside a database transaction
// (via adapter.RunInTx or adapter.RunInTxCtx) to maintain atomicity.
type LedgerService interface {
	// GetUserWallet returns the points wallet for a user, or nil if it doesn't exist.
	GetUserWallet(ctx context.Context, userID uuid.UUID) (*models.LedgerAccount, error)

	// GetOrCreateUserWallet returns (or lazily creates) the points wallet for a user.
	GetOrCreateUserWallet(ctx context.Context, userID uuid.UUID) (*models.LedgerAccount, error)

	// GetSystemAccount returns a pre-seeded system account by its code.
	GetSystemAccount(ctx context.Context, code string) (*models.LedgerAccount, error)

	// GetUserBalance returns the settled balance for a user's wallet.
	GetUserBalance(ctx context.Context, userID uuid.UUID) (int64, error)

	// GetUserAvailableBalance returns balance minus pending holds.
	GetUserAvailableBalance(ctx context.Context, userID uuid.UUID) (int64, error)

	// PostTransfer creates an immediately-posted (settled) transfer.
	// Enforces account flag constraints (e.g., no overdraft).
	// Must be called within a transaction.
	PostTransfer(ctx context.Context, input PostTransferInput) (*models.LedgerTransfer, error)

	// CreatePendingTransfer creates a pending hold on the debit account.
	// Reduces AvailableBalance without affecting settled Balance.
	// Must be called within a transaction.
	CreatePendingTransfer(ctx context.Context, input PostTransferInput) (*models.LedgerTransfer, error)

	// PostPendingTransfer converts a pending transfer to posted (commits the hold).
	// Must be called within a transaction.
	PostPendingTransfer(ctx context.Context, pendingID uuid.UUID) (*models.LedgerTransfer, error)

	// VoidPendingTransfer releases a pending hold with no net balance effect.
	// Must be called within a transaction.
	VoidPendingTransfer(ctx context.Context, pendingID uuid.UUID) (*models.LedgerTransfer, error)

	// FindTransfers returns transfers matching the filter, ordered newest-first.
	FindTransfers(ctx context.Context, filter *stores.LedgerTransferFilter) ([]*models.LedgerTransfer, error)

	// CountTransfers returns the number of transfers matching the filter.
	CountTransfers(ctx context.Context, filter *stores.LedgerTransferFilter) (int64, error)
}

type DbLedgerService struct {
	adapter stores.StorageAdapterInterface
}

var _ LedgerService = (*DbLedgerService)(nil)

func NewDbLedgerService(adapter stores.StorageAdapterInterface) *DbLedgerService {
	return &DbLedgerService{adapter: adapter}
}

func (s *DbLedgerService) ledger() stores.LedgerStore {
	return s.adapter.Ledger()
}

// GetUserWallet returns the wallet for a user, or nil if it doesn't exist.
func (s *DbLedgerService) GetUserWallet(ctx context.Context, userID uuid.UUID) (*models.LedgerAccount, error) {
	return s.ledger().FindAccountByCode(ctx, models.UserWalletCode(userID))
}

// GetOrCreateUserWallet returns the wallet for a user, creating it if it doesn't exist.
func (s *DbLedgerService) GetOrCreateUserWallet(ctx context.Context, userID uuid.UUID) (*models.LedgerAccount, error) {
	code := models.UserWalletCode(userID)
	existing, err := s.ledger().FindAccountByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	return s.ledger().CreateAccount(ctx, &models.LedgerAccount{
		Code:       code,
		EntityType: "user",
		EntityID:   &userID,
		LedgerCode: "POINTS",
		Constraints: []models.AccountConstraint{models.AccountConstraintDebitsMustNotExceedCredits},
		Metadata:   []byte("{}"),
	})
}

// GetSystemAccount returns a pre-seeded system account (must exist in DB).
func (s *DbLedgerService) GetSystemAccount(ctx context.Context, code string) (*models.LedgerAccount, error) {
	acc, err := s.ledger().FindAccountByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("system ledger account %q not found", code)
	}
	return acc, nil
}

// GetUserBalance returns the settled balance of a user's wallet, or 0 if no wallet exists.
func (s *DbLedgerService) GetUserBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	wallet, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return 0, err
	}
	if wallet == nil {
		return 0, nil
	}
	return wallet.Balance(), nil
}

// GetUserAvailableBalance returns the balance minus any pending holds, or 0 if no wallet exists.
func (s *DbLedgerService) GetUserAvailableBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	wallet, err := s.GetUserWallet(ctx, userID)
	if err != nil {
		return 0, err
	}
	if wallet == nil {
		return 0, nil
	}
	return wallet.AvailableBalance(), nil
}

// checkDebitConstraint verifies that posting a debit to `account` by `amount` will not violate constraints.
func checkDebitConstraint(account *models.LedgerAccount, amount int64) error {
	if slices.Contains(account.Constraints, models.AccountConstraintDebitsMustNotExceedCredits) {
		// After posting: debits_posted + amount must not exceed credits_posted.
		if account.DebitsPosted+amount > account.CreditsPosted {
			return fmt.Errorf("insufficient balance in account %s: available %d, requested %d",
				account.Code, account.Balance(), amount)
		}
	}
	return nil
}

// checkAvailableBalanceConstraint verifies available balance before creating a pending hold.
func checkAvailableBalanceConstraint(account *models.LedgerAccount, amount int64) error {
	if slices.Contains(account.Constraints, models.AccountConstraintDebitsMustNotExceedCredits) {
		if account.AvailableBalance() < amount {
			return fmt.Errorf("insufficient available balance in account %s: available %d, requested %d",
				account.Code, account.AvailableBalance(), amount)
		}
	}
	return nil
}

// PostTransfer creates a settled (posted) transfer atomically.
// This must be called within a transaction (RunInTx or RunInTxCtx).
func (s *DbLedgerService) PostTransfer(ctx context.Context, input PostTransferInput) (*models.LedgerTransfer, error) {
	if input.Amount <= 0 {
		return nil, errors.New("transfer amount must be positive")
	}
	ledger := s.ledger()

	// Lock both accounts for update.
	debitAcc, err := ledger.FindAccountByIDForUpdate(ctx, input.DebitAccountID)
	if err != nil {
		return nil, fmt.Errorf("find debit account: %w", err)
	}
	if debitAcc == nil {
		return nil, fmt.Errorf("debit account %s not found", input.DebitAccountID)
	}
	creditAcc, err := ledger.FindAccountByIDForUpdate(ctx, input.CreditAccountID)
	if err != nil {
		return nil, fmt.Errorf("find credit account: %w", err)
	}
	if creditAcc == nil {
		return nil, fmt.Errorf("credit account %s not found", input.CreditAccountID)
	}

	// Check constraints.
	if err := checkDebitConstraint(debitAcc, input.Amount); err != nil {
		return nil, err
	}

	ledgerCode := input.LedgerCode
	if ledgerCode == "" {
		ledgerCode = "POINTS"
	}
	meta := input.Metadata
	if meta == nil {
		meta = []byte("{}")
	}

	// Insert the transfer.
	transfer, err := ledger.CreateTransfer(ctx, &models.LedgerTransfer{
		LedgerCode:      ledgerCode,
		DebitAccountID:  input.DebitAccountID,
		CreditAccountID: input.CreditAccountID,
		Amount:          input.Amount,
		PendingID:       input.PendingID,
		Status:          models.LedgerTransferStatusPosted,
		TransferCode:    input.TransferCode,
		ReferenceType:   input.ReferenceType,
		ReferenceID:     input.ReferenceID,
		Metadata:        meta,
	})
	if err != nil {
		return nil, fmt.Errorf("create transfer: %w", err)
	}

	// Update balances atomically.
	if _, err = ledger.AtomicUpdateAccountBalances(ctx, input.DebitAccountID, input.Amount, 0, 0, 0); err != nil {
		return nil, fmt.Errorf("update debit account: %w", err)
	}
	if _, err = ledger.AtomicUpdateAccountBalances(ctx, input.CreditAccountID, 0, input.Amount, 0, 0); err != nil {
		return nil, fmt.Errorf("update credit account: %w", err)
	}

	return transfer, nil
}

// CreatePendingTransfer creates a pending (held) transfer.
// Increases debit_pending on the debit account, credit_pending on the credit account.
// This must be called within a transaction.
func (s *DbLedgerService) CreatePendingTransfer(ctx context.Context, input PostTransferInput) (*models.LedgerTransfer, error) {
	if input.Amount <= 0 {
		return nil, errors.New("transfer amount must be positive")
	}
	ledger := s.ledger()

	debitAcc, err := ledger.FindAccountByIDForUpdate(ctx, input.DebitAccountID)
	if err != nil {
		return nil, fmt.Errorf("find debit account: %w", err)
	}
	if debitAcc == nil {
		return nil, fmt.Errorf("debit account %s not found", input.DebitAccountID)
	}

	if err := checkAvailableBalanceConstraint(debitAcc, input.Amount); err != nil {
		return nil, err
	}

	ledgerCode := input.LedgerCode
	if ledgerCode == "" {
		ledgerCode = "POINTS"
	}
	meta := input.Metadata
	if meta == nil {
		meta = []byte("{}")
	}

	transfer, err := ledger.CreateTransfer(ctx, &models.LedgerTransfer{
		LedgerCode:      ledgerCode,
		DebitAccountID:  input.DebitAccountID,
		CreditAccountID: input.CreditAccountID,
		Amount:          input.Amount,
		Status:          models.LedgerTransferStatusPending,
		TransferCode:    input.TransferCode,
		ReferenceType:   input.ReferenceType,
		ReferenceID:     input.ReferenceID,
		Metadata:        meta,
	})
	if err != nil {
		return nil, fmt.Errorf("create pending transfer: %w", err)
	}

	// Increase pending counters.
	if _, err = ledger.AtomicUpdateAccountBalances(ctx, input.DebitAccountID, 0, 0, input.Amount, 0); err != nil {
		return nil, fmt.Errorf("update debit account pending: %w", err)
	}
	if _, err = ledger.AtomicUpdateAccountBalances(ctx, input.CreditAccountID, 0, 0, 0, input.Amount); err != nil {
		return nil, fmt.Errorf("update credit account pending: %w", err)
	}

	return transfer, nil
}

// PostPendingTransfer converts a pending transfer to posted.
// Moves amounts from pending counters to posted counters.
// This must be called within a transaction.
func (s *DbLedgerService) PostPendingTransfer(ctx context.Context, pendingID uuid.UUID) (*models.LedgerTransfer, error) {
	ledger := s.ledger()

	pending, err := ledger.FindTransferByIDForUpdate(ctx, pendingID)
	if err != nil {
		return nil, fmt.Errorf("find pending transfer: %w", err)
	}
	if pending == nil {
		return nil, fmt.Errorf("pending transfer %s not found", pendingID)
	}
	if pending.Status != models.LedgerTransferStatusPending {
		return nil, fmt.Errorf("transfer %s is not pending (status: %s)", pendingID, pending.Status)
	}

	posted, err := ledger.UpdateTransferStatus(ctx, pendingID, models.LedgerTransferStatusPosted)
	if err != nil {
		return nil, fmt.Errorf("update transfer status: %w", err)
	}

	// Move from pending to posted on debit account.
	if _, err = ledger.AtomicUpdateAccountBalances(ctx, pending.DebitAccountID, pending.Amount, 0, -pending.Amount, 0); err != nil {
		return nil, fmt.Errorf("post debit account: %w", err)
	}
	// Move from pending to posted on credit account.
	if _, err = ledger.AtomicUpdateAccountBalances(ctx, pending.CreditAccountID, 0, pending.Amount, 0, -pending.Amount); err != nil {
		return nil, fmt.Errorf("post credit account: %w", err)
	}

	return posted, nil
}

// VoidPendingTransfer releases a pending hold with no net balance effect.
// This must be called within a transaction.
func (s *DbLedgerService) VoidPendingTransfer(ctx context.Context, pendingID uuid.UUID) (*models.LedgerTransfer, error) {
	ledger := s.ledger()

	pending, err := ledger.FindTransferByIDForUpdate(ctx, pendingID)
	if err != nil {
		return nil, fmt.Errorf("find pending transfer: %w", err)
	}
	if pending == nil {
		return nil, fmt.Errorf("pending transfer %s not found", pendingID)
	}
	if pending.Status != models.LedgerTransferStatusPending {
		return nil, fmt.Errorf("transfer %s is not pending (status: %s)", pendingID, pending.Status)
	}

	voided, err := ledger.UpdateTransferStatus(ctx, pendingID, models.LedgerTransferStatusVoided)
	if err != nil {
		return nil, fmt.Errorf("update transfer status: %w", err)
	}

	// Release the holds.
	if _, err = ledger.AtomicUpdateAccountBalances(ctx, pending.DebitAccountID, 0, 0, -pending.Amount, 0); err != nil {
		return nil, fmt.Errorf("void debit account: %w", err)
	}
	if _, err = ledger.AtomicUpdateAccountBalances(ctx, pending.CreditAccountID, 0, 0, 0, -pending.Amount); err != nil {
		return nil, fmt.Errorf("void credit account: %w", err)
	}

	return voided, nil
}

func (s *DbLedgerService) FindTransfers(ctx context.Context, filter *stores.LedgerTransferFilter) ([]*models.LedgerTransfer, error) {
	return s.ledger().FindTransfers(ctx, filter)
}

func (s *DbLedgerService) CountTransfers(ctx context.Context, filter *stores.LedgerTransferFilter) (int64, error) {
	return s.ledger().CountTransfers(ctx, filter)
}
