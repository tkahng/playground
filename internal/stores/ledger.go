package stores

import (
	"context"

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
	Ids              []uuid.UUID                   `query:"ids,omitempty"`
	Statuses         []models.LedgerTransferStatus `query:"statuses,omitempty"`
	TransferCodes    []string                      `query:"transfer_codes,omitempty"`
	DebitAccountIds  []uuid.UUID                   `query:"debit_account_ids,omitempty"`
	CreditAccountIds []uuid.UUID                   `query:"credit_account_ids,omitempty"`
	AccountIds       []uuid.UUID                   `query:"account_ids,omitempty"` // debit OR credit
	ReferenceTypes   []string                      `query:"reference_types,omitempty"`
	ReferenceIds     []uuid.UUID                   `query:"reference_ids,omitempty"`
	LedgerCodes      []string                      `query:"ledger_codes,omitempty"`
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
