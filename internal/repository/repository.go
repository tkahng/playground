package repository

import (
	"context"

	"github.com/tkahng/playground/internal/database"
)

type Repository[Model any] interface {
	Get(ctx context.Context, dbx database.Dbx, where *map[string]any, order *map[string]string, limit *int, skip *int) ([]*Model, error)
	GetOne(ctx context.Context, dbx database.Dbx, where *map[string]any) (*Model, error)
	Put(ctx context.Context, dbx database.Dbx, models []Model) ([]*Model, error)
	PutOne(ctx context.Context, dbx database.Dbx, model *Model) (*Model, error)
	PostOne(ctx context.Context, dbx database.Dbx, model *Model) (*Model, error)
	Post(ctx context.Context, dbx database.Dbx, models []Model) ([]*Model, error)
	PostExec(ctx context.Context, dbx database.Dbx, models []Model) (int64, error)

	DeleteReturn(ctx context.Context, dbx database.Dbx, where *map[string]any) ([]*Model, error)
	Delete(ctx context.Context, dbx database.Dbx, where *map[string]any) (int64, error)
	Count(ctx context.Context, dbx database.Dbx, where *map[string]any) (int64, error)
	Builder() SQLBuilderInterface
}
