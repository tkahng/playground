package repository

import "context"

type ModelResource[Model any, Key comparable] interface {
	Index(ctx context.Context)
	Show(ctx context.Context, id Key) (*Model, error)
	Create(ctx context.Context, model Model) (*Model, error)
	Update(ctx context.Context, model Model) (*Model, error)
	Delete(ctx context.Context, id Key) error
}
