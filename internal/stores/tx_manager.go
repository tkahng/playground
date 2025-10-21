package stores

import (
	"context"

	"github.com/tkahng/playground/internal/database"
)

type TxManager interface {
	RunInTxCtx(ctx context.Context, fn func(ctx context.Context) error) error
}

type TxManagerImpl struct {
	db database.Dbx
}

func NewTxManager(db database.Dbx) *TxManagerImpl {
	return &TxManagerImpl{db: db}
}

// RunInTxCtx implements TxManager.
func (t *TxManagerImpl) RunInTxCtx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.RunInTxCtx(ctx, fn)
}

var _ TxManager = (*TxManagerImpl)(nil)
