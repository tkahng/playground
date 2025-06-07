package stores

import "github.com/tkahng/authgo/internal/database"

type TransactionProvider struct {
	db database.Dbx
}

type TxProvider[T any] interface {
	Transact(txFunc func(adapters T) error) error
}

func NewTransactionProvider(db database.Dbx) *TransactionProvider {
	return &TransactionProvider{
		db: db,
	}
}
