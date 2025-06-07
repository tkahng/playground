package resource

import (
	"context"
	"errors"

	"github.com/tkahng/authgo/internal/database"
)

var (
	ResourceNotImplementedError = errors.New("resource not implemented")
)

type Resource[Model any, Key comparable, Filter any] interface {
	Delete(ctx context.Context, id Key) error
	Update(ctx context.Context, model *Model) (*Model, error)
	Create(ctx context.Context, model *Model) (*Model, error)
	Count(ctx context.Context, filter *Filter) (int64, error)
	Find(ctx context.Context, filter *Filter) ([]*Model, error)
	FindByID(ctx context.Context, id Key) (*Model, error)
	WithTx(tx database.Dbx) Resource[Model, Key, Filter]
}
