package stores

import (
	"context"

	"github.com/tkahng/playground/internal/database"
)

type TxManager interface {
	RunInTxContext(ctx context.Context, fn func(ctx context.Context) error) error
}

type TxManagerImpl struct {
	db database.Dbx
}

func NewTxManager(db database.Dbx) *TxManagerImpl {
	return &TxManagerImpl{db: db}
}

// RunInTxContext implements TxManager.
func (t *TxManagerImpl) RunInTxContext(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.RunInTxContext(context.Background(), fn)
}

var _ TxManager = (*TxManagerImpl)(nil)
