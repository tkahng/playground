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

func (s *DBLedgerStore) CreateTransfer(ctx context.Context, transfer *models.LedgerTransfer) (*models.LedgerTransfer, error) {
	if transfer == nil {
		return nil, errors.New("ledger transfer is nil")
	}
	if transfer.Metadata == nil {
		transfer.Metadata = []byte("{}")
	}
	return repository.LedgerTransfer.PostOne(ctx, s.db, transfer)
}

func buildLedgerTransferQuery(filter *LedgerTransferFilter) squirrel.SelectBuilder {
	q := squirrel.Select(repository.LedgerTransferBuilder.ColumnNames()...).From("ledger.transfers")
	if filter == nil {
		return q
	}
	if len(filter.Ids) > 0 {
		q = q.Where(squirrel.Eq{"ledger.transfers.id": filter.Ids})
	}
	if len(filter.Statuses) > 0 {
		q = q.Where(squirrel.Eq{"ledger.transfers.status": filter.Statuses})
	}
	if len(filter.TransferCodes) > 0 {
		q = q.Where(squirrel.Eq{"ledger.transfers.transfer_code": filter.TransferCodes})
	}
	if len(filter.DebitAccountIds) > 0 {
		q = q.Where(squirrel.Eq{"ledger.transfers.debit_account_id": filter.DebitAccountIds})
	}
	if len(filter.CreditAccountIds) > 0 {
		q = q.Where(squirrel.Eq{"ledger.transfers.credit_account_id": filter.CreditAccountIds})
	}
	if len(filter.AccountIds) > 0 {
		q = q.Where(squirrel.Or{
			squirrel.Eq{"ledger.transfers.debit_account_id": filter.AccountIds},
			squirrel.Eq{"ledger.transfers.credit_account_id": filter.AccountIds},
		})
	}
	if len(filter.ReferenceTypes) > 0 {
		q = q.Where(squirrel.Eq{"ledger.transfers.reference_type": filter.ReferenceTypes})
	}
	if len(filter.ReferenceIds) > 0 {
		q = q.Where(squirrel.Eq{"ledger.transfers.reference_id": filter.ReferenceIds})
	}
	if len(filter.LedgerCodes) > 0 {
		q = q.Where(squirrel.Eq{"ledger.transfers.ledger_code": filter.LedgerCodes})
	}
	return q
}

func (s *DBLedgerStore) FindTransfer(ctx context.Context, filter *LedgerTransferFilter) (*models.LedgerTransfer, error) {
	q := buildLedgerTransferQuery(filter).
		OrderBy("ledger.transfers.created_at DESC").
		Limit(1)
	data, err := database.PgxQueryRowsToStruct[models.LedgerTransfer](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

// FindTransferByIDForUpdate fetches a transfer and locks the row for update (must be in a transaction).
func (s *DBLedgerStore) FindTransferByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.LedgerTransfer, error) {
	cols := strings.Join(repository.LedgerTransferBuilder.ColumnNames(), ", ")
	query := fmt.Sprintf("SELECT %s FROM ledger.transfers WHERE id = $1 FOR UPDATE", cols)
	data, err := database.QueryAll[*models.LedgerTransfer](ctx, s.db, query, id)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data[0], nil
}

func (s *DBLedgerStore) FindTransfers(ctx context.Context, filter *LedgerTransferFilter) ([]*models.LedgerTransfer, error) {
	q := buildLedgerTransferQuery(filter).OrderBy("ledger.transfers.created_at DESC")
	q = queryPagination(q, filter)
	data, err := database.PgxQueryRowsToStruct[models.LedgerTransfer](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *DBLedgerStore) CountTransfers(ctx context.Context, filter *LedgerTransferFilter) (int64, error) {
	q := squirrel.Select("COUNT(*)").From("ledger.transfers")
	if filter != nil {
		if len(filter.Ids) > 0 {
			q = q.Where(squirrel.Eq{"ledger.transfers.id": filter.Ids})
		}
		if len(filter.Statuses) > 0 {
			q = q.Where(squirrel.Eq{"ledger.transfers.status": filter.Statuses})
		}
		if len(filter.TransferCodes) > 0 {
			q = q.Where(squirrel.Eq{"ledger.transfers.transfer_code": filter.TransferCodes})
		}
		if len(filter.AccountIds) > 0 {
			q = q.Where(squirrel.Or{
				squirrel.Eq{"ledger.transfers.debit_account_id": filter.AccountIds},
				squirrel.Eq{"ledger.transfers.credit_account_id": filter.AccountIds},
			})
		}
		if len(filter.ReferenceTypes) > 0 {
			q = q.Where(squirrel.Eq{"ledger.transfers.reference_type": filter.ReferenceTypes})
		}
		if len(filter.ReferenceIds) > 0 {
			q = q.Where(squirrel.Eq{"ledger.transfers.reference_id": filter.ReferenceIds})
		}
	}
	c, err := database.QueryWithBuilder[database.CountOutput](ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(c) == 0 {
		return 0, nil
	}
	return c[0].Count, nil
}

// UpdateTransferStatus updates only the status column of a transfer.
func (s *DBLedgerStore) UpdateTransferStatus(ctx context.Context, id uuid.UUID, status models.LedgerTransferStatus) (*models.LedgerTransfer, error) {
	cols := strings.Join(repository.LedgerTransferBuilder.ColumnNames(), ", ")
	query := fmt.Sprintf("UPDATE ledger.transfers SET status = $1 WHERE id = $2 RETURNING %s", cols)
	data, err := database.QueryAll[*models.LedgerTransfer](ctx, s.db, query, string(status), id)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("ledger transfer %s not found", id)
	}
	return data[0], nil
}
