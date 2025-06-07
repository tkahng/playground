package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
)

var (
	ResourceNotImplementedError = errors.New("resource not implemented")
)

type Resource[Model any, Key comparable, Filter any] interface {
	Delete(ctx context.Context, id Key) error
	Update(ctx context.Context, model *Model) (*Model, error)
	Create(ctx context.Context, model *Model) (*Model, error)
	Count(ctx context.Context, filter Filter) (int64, error)
	Find(ctx context.Context, filter Filter) ([]*Model, error)
	FindOne(ctx context.Context, id Key) (*Model, error)
	WithTx(tx database.Dbx) Resource[Model, Key, Filter]
}

type ModelResource[Model any, Key comparable, Filter any] struct {
	Resource[Model, Key, Filter]
}

var _ Resource[models.User, uuid.UUID, any] = (*ModelResource[models.User, uuid.UUID, any])(nil)

func (r *ModelResource[Model, Key, Filter]) WithTx(tx database.Dbx) Resource[Model, Key, Filter] {
	if r.Resource == nil {
		panic(ResourceNotImplementedError)
	}
	return r.Resource.WithTx(tx)
}

func (r *ModelResource[Model, Key, Filter]) Find(ctx context.Context, filter Filter) ([]*Model, error) {
	if r.Resource == nil {
		return nil, ResourceNotImplementedError
	}
	return r.Resource.Find(ctx, filter)
}
func (r *ModelResource[Model, Key, Filter]) Count(ctx context.Context, filter Filter) (int64, error) {
	if r.Resource == nil {
		return 0, ResourceNotImplementedError
	}
	return r.Resource.Count(ctx, filter)
}
func (r *ModelResource[Model, Key, Filter]) FindOne(ctx context.Context, id Key) (*Model, error) {
	if r.Resource == nil {
		return nil, ResourceNotImplementedError
	}
	return r.Resource.FindOne(ctx, id)
}
func (r *ModelResource[Model, Key, Filter]) Create(ctx context.Context, model *Model) (*Model, error) {
	if r.Resource == nil {
		return nil, ResourceNotImplementedError
	}
	return r.Resource.Create(ctx, model)
}
func (r *ModelResource[Model, Key, Filter]) Update(ctx context.Context, model *Model) (*Model, error) {
	if r.Resource == nil {
		return nil, ResourceNotImplementedError
	}
	return r.Resource.Update(ctx, model)
}

func (r *ModelResource[Model, Key, Filter]) Delete(ctx context.Context, id Key) error {
	if r.Resource == nil {
		return ResourceNotImplementedError
	}
	return r.Resource.Delete(ctx, id)
}
