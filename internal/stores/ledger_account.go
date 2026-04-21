package stores

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
)

func (s *DBLedgerStore) CreateAccount(ctx context.Context, account *models.LedgerAccount) (*models.LedgerAccount, error) {
	if account == nil {
		return nil, errors.New("ledger account is nil")
	}
	if account.Metadata == nil {
		account.Metadata = []byte("{}")
	}
	return repository.LedgerAccount.PostOne(ctx, s.db, account)
}

func buildLedgerAccountQuery(filter *LedgerAccountFilter) squirrel.SelectBuilder {
	q := squirrel.Select(repository.LedgerAccountBuilder.ColumnNames()...).From("ledger.accounts")
	if filter == nil {
		return q
	}
	if len(filter.Ids) > 0 {
		q = q.Where(squirrel.Eq{"ledger.accounts.id": filter.Ids})
	}
	if len(filter.Codes) > 0 {
		q = q.Where(squirrel.Eq{"ledger.accounts.code": filter.Codes})
	}
	if len(filter.EntityTypes) > 0 {
		q = q.Where(squirrel.Eq{"ledger.accounts.entity_type": filter.EntityTypes})
	}
	if len(filter.EntityIds) > 0 {
		q = q.Where(squirrel.Eq{"ledger.accounts.entity_id": filter.EntityIds})
	}
	if len(filter.LedgerCodes) > 0 {
		q = q.Where(squirrel.Eq{"ledger.accounts.ledger_code": filter.LedgerCodes})
	}
	return q
}

func (s *DBLedgerStore) FindAccount(ctx context.Context, filter *LedgerAccountFilter) (*models.LedgerAccount, error) {
	q := buildLedgerAccountQuery(filter).Limit(1)
	data, err := database.PgxQueryRowsToStruct[models.LedgerAccount](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

func (s *DBLedgerStore) FindAccountByCode(ctx context.Context, code string) (*models.LedgerAccount, error) {
	return s.FindAccount(ctx, &LedgerAccountFilter{Codes: []string{code}})
}

// FindAccountByIDForUpdate fetches an account and locks the row for update (must be called inside a transaction).
func (s *DBLedgerStore) FindAccountByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.LedgerAccount, error) {
	cols := strings.Join(repository.LedgerAccountBuilder.ColumnNames(), ", ")
	query := fmt.Sprintf("SELECT %s FROM ledger.accounts WHERE id = $1 FOR UPDATE", cols)
	data, err := database.QueryAll[*models.LedgerAccount](ctx, s.db, query, id)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

// AtomicUpdateAccountBalances applies delta increments to an account's balance columns atomically.
// All deltas may be negative (for decrement). Must be called inside a transaction.
func (s *DBLedgerStore) AtomicUpdateAccountBalances(
	ctx context.Context,
	id uuid.UUID,
	deltaDebitsPosted, deltaCreditsPosted,
	deltaDebitsPending, deltaCreditsPending int64,
) (*models.LedgerAccount, error) {
	cols := strings.Join(repository.LedgerAccountBuilder.ColumnNames(), ", ")
	query := fmt.Sprintf(`
		UPDATE ledger.accounts
		SET debits_posted   = debits_posted   + $1,
		    credits_posted  = credits_posted  + $2,
		    debits_pending  = debits_pending  + $3,
		    credits_pending = credits_pending + $4
		WHERE id = $5
		RETURNING %s`, cols)
	data, err := database.QueryAll[*models.LedgerAccount](ctx, s.db, query,
		deltaDebitsPosted, deltaCreditsPosted,
		deltaDebitsPending, deltaCreditsPending,
		id,
	)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("ledger account %s not found", id)
	}
	return data[0], nil
}
