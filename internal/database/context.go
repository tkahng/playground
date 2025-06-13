package database

import (
	"context"
)

type contextKey string

const (
	txContextKey contextKey = "tx_context_key"
)

func setContextTx(ctx context.Context, tx Dbx) context.Context {
	return context.WithValue(ctx, txContextKey, tx)
}
func getContextTx(ctx context.Context) Dbx {
	if tx, ok := ctx.Value(txContextKey).(Dbx); ok {
		return tx
	} else {
		return nil
	}
}
