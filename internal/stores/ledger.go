package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

// LedgerAccountFilter holds filter parameters for querying ledger accounts.
type LedgerAccountFilter struct {
	PaginatedInput
	Codes       []string    `query:"codes,omitempty"`
	Ids         []uuid.UUID `query:"ids,omitempty"`
	EntityTypes []string    `query:"entity_types,omitempty"`
	EntityIds   []uuid.UUID `query:"entity_ids,omitempty"`
	LedgerCodes []string    `query:"ledger_codes,omitempty"`
}

// LedgerTransferFilter holds filter parameters for querying ledger transfers.
type LedgerTransferFilter struct {
	PaginatedInput
	Ids              []uuid.UUID                    `query:"ids,omitempty"`
	Statuses         []models.LedgerTransferStatus  `query:"statuses,omitempty"`
	TransferCodes    []string                       `query:"transfer_codes,omitempty"`
	DebitAccountIds  []uuid.UUID                    `query:"debit_account_ids,omitempty"`
	CreditAccountIds []uuid.UUID                    `query:"credit_account_ids,omitempty"`
	AccountIds       []uuid.UUID                    `query:"account_ids,omitempty"` // debit OR credit
	ReferenceTypes   []string                       `query:"reference_types,omitempty"`
	ReferenceIds     []uuid.UUID                    `query:"reference_ids,omitempty"`
	LedgerCodes      []string                       `query:"ledger_codes,omitempty"`
}

// LedgerStore is the data-access interface for the double-entry ledger.
type LedgerStore interface {
	// Account operations
	CreateAccount(ctx context.Context, account *models.LedgerAccount) (*models.LedgerAccount, error)
	FindAccount(ctx context.Context, filter *LedgerAccountFilter) (*models.LedgerAccount, error)
	FindAccountByCode(ctx context.Context, code string) (*models.LedgerAccount, error)
	FindAccountByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.LedgerAccount, error)
	AtomicUpdateAccountBalances(ctx context.Context, id uuid.UUID, deltaDebitsPosted, deltaCreditsPosted, deltaDebitsPending, deltaCreditsPending int64) (*models.LedgerAccount, error)

	// Transfer operations
	CreateTransfer(ctx context.Context, transfer *models.LedgerTransfer) (*models.LedgerTransfer, error)
	FindTransfer(ctx context.Context, filter *LedgerTransferFilter) (*models.LedgerTransfer, error)
	FindTransferByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.LedgerTransfer, error)
	FindTransfers(ctx context.Context, filter *LedgerTransferFilter) ([]*models.LedgerTransfer, error)
	CountTransfers(ctx context.Context, filter *LedgerTransferFilter) (int64, error)
	UpdateTransferStatus(ctx context.Context, id uuid.UUID, status models.LedgerTransferStatus) (*models.LedgerTransfer, error)

	WithTx(db database.Dbx) *DBLedgerStore
}

// DBLedgerStore is the PostgreSQL implementation of LedgerStore.
type DBLedgerStore struct {
	db database.Dbx
}

func NewDBLedgerStore(db database.Dbx) *DBLedgerStore {
	return &DBLedgerStore{db: db}
}

func (s *DBLedgerStore) WithTx(db database.Dbx) *DBLedgerStore {
	return &DBLedgerStore{db: db}
}

var _ LedgerStore = (*DBLedgerStore)(nil)

// DbLedgerStoreDecorator enables mocking LedgerStore in tests.
type DbLedgerStoreDecorator struct {
	Delegate                          *DBLedgerStore
	CreateAccountFunc                 func(ctx context.Context, account *models.LedgerAccount) (*models.LedgerAccount, error)
	FindAccountFunc                   func(ctx context.Context, filter *LedgerAccountFilter) (*models.LedgerAccount, error)
	FindAccountByCodeFunc             func(ctx context.Context, code string) (*models.LedgerAccount, error)
	FindAccountByIDForUpdateFunc      func(ctx context.Context, id uuid.UUID) (*models.LedgerAccount, error)
	AtomicUpdateAccountBalancesFunc   func(ctx context.Context, id uuid.UUID, deltaDebitsPosted, deltaCreditsPosted, deltaDebitsPending, deltaCreditsPending int64) (*models.LedgerAccount, error)
	CreateTransferFunc                func(ctx context.Context, transfer *models.LedgerTransfer) (*models.LedgerTransfer, error)
	FindTransferFunc                  func(ctx context.Context, filter *LedgerTransferFilter) (*models.LedgerTransfer, error)
	FindTransferByIDForUpdateFunc     func(ctx context.Context, id uuid.UUID) (*models.LedgerTransfer, error)
	FindTransfersFunc                 func(ctx context.Context, filter *LedgerTransferFilter) ([]*models.LedgerTransfer, error)
	CountTransfersFunc                func(ctx context.Context, filter *LedgerTransferFilter) (int64, error)
	UpdateTransferStatusFunc          func(ctx context.Context, id uuid.UUID, status models.LedgerTransferStatus) (*models.LedgerTransfer, error)
	WithTxFunc                        func(db database.Dbx) *DBLedgerStore
}

func NewDbLedgerStoreDecorator(db database.Dbx) *DbLedgerStoreDecorator {
	return &DbLedgerStoreDecorator{Delegate: NewDBLedgerStore(db)}
}

var _ LedgerStore = (*DbLedgerStoreDecorator)(nil)

func (s *DbLedgerStoreDecorator) WithTx(db database.Dbx) *DBLedgerStore {
	if s.WithTxFunc != nil {
		return s.WithTxFunc(db)
	}
	return s.Delegate.WithTx(db)
}

func (s *DbLedgerStoreDecorator) delegate(op string) error {
	if s.Delegate == nil {
		return fmt.Errorf("LedgerStore decorator %s: %w", op, ErrDelegateNil)
	}
	return nil
}

func (s *DbLedgerStoreDecorator) CreateAccount(ctx context.Context, account *models.LedgerAccount) (*models.LedgerAccount, error) {
	if s.CreateAccountFunc != nil {
		return s.CreateAccountFunc(ctx, account)
	}
	if err := s.delegate("CreateAccount"); err != nil {
		return nil, err
	}
	return s.Delegate.CreateAccount(ctx, account)
}

func (s *DbLedgerStoreDecorator) FindAccount(ctx context.Context, filter *LedgerAccountFilter) (*models.LedgerAccount, error) {
	if s.FindAccountFunc != nil {
		return s.FindAccountFunc(ctx, filter)
	}
	if err := s.delegate("FindAccount"); err != nil {
		return nil, err
	}
	return s.Delegate.FindAccount(ctx, filter)
}

func (s *DbLedgerStoreDecorator) FindAccountByCode(ctx context.Context, code string) (*models.LedgerAccount, error) {
	if s.FindAccountByCodeFunc != nil {
		return s.FindAccountByCodeFunc(ctx, code)
	}
	if err := s.delegate("FindAccountByCode"); err != nil {
		return nil, err
	}
	return s.Delegate.FindAccountByCode(ctx, code)
}

func (s *DbLedgerStoreDecorator) FindAccountByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.LedgerAccount, error) {
	if s.FindAccountByIDForUpdateFunc != nil {
		return s.FindAccountByIDForUpdateFunc(ctx, id)
	}
	if err := s.delegate("FindAccountByIDForUpdate"); err != nil {
		return nil, err
	}
	return s.Delegate.FindAccountByIDForUpdate(ctx, id)
}

func (s *DbLedgerStoreDecorator) AtomicUpdateAccountBalances(ctx context.Context, id uuid.UUID, deltaDebitsPosted, deltaCreditsPosted, deltaDebitsPending, deltaCreditsPending int64) (*models.LedgerAccount, error) {
	if s.AtomicUpdateAccountBalancesFunc != nil {
		return s.AtomicUpdateAccountBalancesFunc(ctx, id, deltaDebitsPosted, deltaCreditsPosted, deltaDebitsPending, deltaCreditsPending)
	}
	if err := s.delegate("AtomicUpdateAccountBalances"); err != nil {
		return nil, err
	}
	return s.Delegate.AtomicUpdateAccountBalances(ctx, id, deltaDebitsPosted, deltaCreditsPosted, deltaDebitsPending, deltaCreditsPending)
}

func (s *DbLedgerStoreDecorator) CreateTransfer(ctx context.Context, transfer *models.LedgerTransfer) (*models.LedgerTransfer, error) {
	if s.CreateTransferFunc != nil {
		return s.CreateTransferFunc(ctx, transfer)
	}
	if err := s.delegate("CreateTransfer"); err != nil {
		return nil, err
	}
	return s.Delegate.CreateTransfer(ctx, transfer)
}

func (s *DbLedgerStoreDecorator) FindTransfer(ctx context.Context, filter *LedgerTransferFilter) (*models.LedgerTransfer, error) {
	if s.FindTransferFunc != nil {
		return s.FindTransferFunc(ctx, filter)
	}
	if err := s.delegate("FindTransfer"); err != nil {
		return nil, err
	}
	return s.Delegate.FindTransfer(ctx, filter)
}

func (s *DbLedgerStoreDecorator) FindTransferByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.LedgerTransfer, error) {
	if s.FindTransferByIDForUpdateFunc != nil {
		return s.FindTransferByIDForUpdateFunc(ctx, id)
	}
	if err := s.delegate("FindTransferByIDForUpdate"); err != nil {
		return nil, err
	}
	return s.Delegate.FindTransferByIDForUpdate(ctx, id)
}

func (s *DbLedgerStoreDecorator) FindTransfers(ctx context.Context, filter *LedgerTransferFilter) ([]*models.LedgerTransfer, error) {
	if s.FindTransfersFunc != nil {
		return s.FindTransfersFunc(ctx, filter)
	}
	if err := s.delegate("FindTransfers"); err != nil {
		return nil, err
	}
	return s.Delegate.FindTransfers(ctx, filter)
}

func (s *DbLedgerStoreDecorator) CountTransfers(ctx context.Context, filter *LedgerTransferFilter) (int64, error) {
	if s.CountTransfersFunc != nil {
		return s.CountTransfersFunc(ctx, filter)
	}
	if err := s.delegate("CountTransfers"); err != nil {
		return 0, err
	}
	return s.Delegate.CountTransfers(ctx, filter)
}

func (s *DbLedgerStoreDecorator) UpdateTransferStatus(ctx context.Context, id uuid.UUID, status models.LedgerTransferStatus) (*models.LedgerTransfer, error) {
	if s.UpdateTransferStatusFunc != nil {
		return s.UpdateTransferStatusFunc(ctx, id, status)
	}
	if err := s.delegate("UpdateTransferStatus"); err != nil {
		return nil, err
	}
	return s.Delegate.UpdateTransferStatus(ctx, id, status)
}
