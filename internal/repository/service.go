package repository

import (
	"context"

	"github.com/tkahng/authgo/internal/database"
)

type DbService[Model any] interface {
	Get(ctx context.Context, where *map[string]any, order *map[string]string, limit *int, skip *int) ([]*Model, error)
	GetOne(ctx context.Context, where *map[string]any) (*Model, error)
	Put(ctx context.Context, models []Model) ([]*Model, error)
	Post(ctx context.Context, models []Model) ([]*Model, error)
	DeleteReturn(ctx context.Context, where *map[string]any) ([]*Model, error)
	Delete(ctx context.Context, where *map[string]any) (int64, error)
	Count(ctx context.Context, where *map[string]any) (int64, error)
	Builder() SQLBuilderInterface
}

var _ DbService[any] = (*PostgresDbService[any])(nil)

func NewPostgresDbService[Model any](db func() database.Dbx, repo CrudRepo[Model]) *PostgresDbService[Model] {
	return &PostgresDbService[Model]{
		db:   db,
		repo: repo,
	}
}

type PostgresDbService[Model any] struct {
	db   func() database.Dbx
	repo CrudRepo[Model]
}

// Builder implements DbService.
func (p *PostgresDbService[Model]) Builder() SQLBuilderInterface {
	return p.repo.Builder()
}

// Count implements DbService.
func (p *PostgresDbService[Model]) Count(ctx context.Context, where *map[string]any) (int64, error) {
	return p.repo.Count(ctx, p.db(), where)
}

// Delete implements DbService.
func (p *PostgresDbService[Model]) Delete(ctx context.Context, where *map[string]any) (int64, error) {
	return p.repo.Delete(ctx, p.db(), where)
}

// DeleteReturn implements DbService.
func (p *PostgresDbService[Model]) DeleteReturn(ctx context.Context, where *map[string]any) ([]*Model, error) {
	return p.repo.DeleteReturn(ctx, p.db(), where)
}

// Get implements DbService.
func (p *PostgresDbService[Model]) Get(ctx context.Context, where *map[string]any, order *map[string]string, limit *int, skip *int) ([]*Model, error) {
	return p.repo.Get(ctx, p.db(), where, order, limit, skip)
}

// GetOne implements DbService.
func (p *PostgresDbService[Model]) GetOne(ctx context.Context, where *map[string]any) (*Model, error) {
	return p.repo.GetOne(ctx, p.db(), where)
}

// Post implements DbService.
func (p *PostgresDbService[Model]) Post(ctx context.Context, models []Model) ([]*Model, error) {
	return p.repo.Post(ctx, p.db(), models)
}

// Put implements DbService.
func (p *PostgresDbService[Model]) Put(ctx context.Context, models []Model) ([]*Model, error) {
	return p.repo.Put(ctx, p.db(), models)
}
