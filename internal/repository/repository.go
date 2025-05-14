package repository

import (
	"context"

	"github.com/tkahng/authgo/internal/db"
)

type CrudRepo[Model any] interface {
	Get(ctx context.Context, dbx db.Dbx, where *map[string]any, order *map[string]string, limit *int, skip *int) ([]*Model, error)
	GetOne(ctx context.Context, dbx db.Dbx, where *map[string]any) (*Model, error)
	Put(ctx context.Context, dbx db.Dbx, models []Model) ([]*Model, error)
	Post(ctx context.Context, dbx db.Dbx, models []Model) ([]*Model, error)
	DeleteReturn(ctx context.Context, dbx db.Dbx, where *map[string]any) ([]*Model, error)
	Delete(ctx context.Context, dbx db.Dbx, where *map[string]any) (int64, error)
	Count(ctx context.Context, dbx db.Dbx, where *map[string]any) (int64, error)
	Builder() SQLBuilderInterface
}
