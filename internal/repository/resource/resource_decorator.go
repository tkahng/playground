package resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
)

type ResourceDecorator[Model any, Key comparable, Filter any] struct {
	Delegate     Resource[Model, Key, Filter]
	DeleteFunc   func(ctx context.Context, key Key) error
	CountFunc    func(ctx context.Context, filter *Filter) (int64, error)
	CreateFunc   func(ctx context.Context, model *Model) (*Model, error)
	FindFunc     func(ctx context.Context, filter *Filter) ([]*Model, error)
	FindByIDFunc func(ctx context.Context, id Key) (*Model, error)
	FindOneFunc  func(ctx context.Context, filter *Filter) (*Model, error)
	UpdateFunc   func(ctx context.Context, model *Model) (*Model, error)
	WithTxFunc   func(tx database.Dbx) Resource[Model, Key, Filter]
}

// Count implements Resource.
func (r *ResourceDecorator[Model, Key, Filter]) Count(ctx context.Context, filter *Filter) (int64, error) {
	if r.CountFunc != nil {
		return r.CountFunc(ctx, filter)
	}
	if r.Delegate == nil {
		return 0, ErrResourceNotImplemented
	}
	return r.Delegate.Count(ctx, filter)
}

// Create implements Resource.
func (r *ResourceDecorator[Model, Key, Filter]) Create(ctx context.Context, model *Model) (*Model, error) {
	if r.CreateFunc != nil {
		return r.CreateFunc(ctx, model)
	}
	if r.Delegate == nil {
		return nil, ErrResourceNotImplemented
	}
	return r.Delegate.Create(ctx, model)
}

// Delete implements Resource.
func (r *ResourceDecorator[Model, Key, Filter]) Delete(ctx context.Context, id Key) error {
	if r.DeleteFunc != nil {
		return r.DeleteFunc(ctx, id)
	}
	if r.Delegate == nil {
		return ErrResourceNotImplemented
	}
	return r.Delegate.Delete(ctx, id)
}

// Find implements Resource.
func (r *ResourceDecorator[Model, Key, Filter]) Find(ctx context.Context, filter *Filter) ([]*Model, error) {
	if r.FindFunc != nil {
		return r.FindFunc(ctx, filter)
	}
	if r.Delegate == nil {
		return nil, ErrResourceNotImplemented
	}
	return r.Delegate.Find(ctx, filter)
}

// FindByID implements Resource.
func (r *ResourceDecorator[Model, Key, Filter]) FindByID(ctx context.Context, id Key) (*Model, error) {
	if r.FindByIDFunc != nil {
		return r.FindByIDFunc(ctx, id)
	}
	if r.Delegate == nil {
		return nil, ErrResourceNotImplemented
	}
	return r.Delegate.FindByID(ctx, id)
}

// FindOne implements Resource.
func (r *ResourceDecorator[Model, Key, Filter]) FindOne(ctx context.Context, filter *Filter) (*Model, error) {
	if r.FindOneFunc != nil {
		return r.FindOneFunc(ctx, filter)
	}
	if r.Delegate == nil {
		return nil, ErrResourceNotImplemented
	}
	return r.Delegate.FindOne(ctx, filter)
}

// Update implements Resource.
func (r *ResourceDecorator[Model, Key, Filter]) Update(ctx context.Context, model *Model) (*Model, error) {
	if r.UpdateFunc != nil {
		return r.UpdateFunc(ctx, model)
	}
	if r.Delegate == nil {
		return nil, ErrResourceNotImplemented
	}
	return r.Delegate.Update(ctx, model)
}

// WithTx implements Resource.
func (r *ResourceDecorator[Model, Key, Filter]) WithTx(tx database.Dbx) Resource[Model, Key, Filter] {
	if r.WithTxFunc != nil {
		return r.WithTxFunc(tx)
	}
	if r.Delegate == nil {
		return nil
	}
	return r.Delegate.WithTx(tx)
}

var _ Resource[any, any, any] = (*ResourceDecorator[any, any, any])(nil)

func NewResourceDecorator[Model any, Key comparable, Filter any](
	delegate Resource[Model, Key, Filter],
) *ResourceDecorator[Model, Key, Filter] {
	return &ResourceDecorator[Model, Key, Filter]{
		Delegate: delegate,
	}
}

type ResourceDecoratorAdapter struct {
	user        *ResourceDecorator[models.User, uuid.UUID, UserFilter]
	permission  *ResourceDecorator[models.Permission, uuid.UUID, PermissionsFilter]
	userAccount *ResourceDecorator[models.UserAccount, uuid.UUID, UserAccountFilter]
	token       *ResourceDecorator[models.Token, uuid.UUID, TokenFilter]
}

func (r *ResourceDecoratorAdapter) User() Resource[models.User, uuid.UUID, UserFilter] {
	return r.user
}
func (r *ResourceDecoratorAdapter) Permission() Resource[models.Permission, uuid.UUID, PermissionsFilter] {
	return r.permission
}
func (r *ResourceDecoratorAdapter) UserAccount() Resource[models.UserAccount, uuid.UUID, UserAccountFilter] {
	return r.userAccount
}
func (r *ResourceDecoratorAdapter) Token() Resource[models.Token, uuid.UUID, TokenFilter] {
	return r.token
}
func NewResourceDecoratorAdapter() *ResourceDecoratorAdapter {
	return &ResourceDecoratorAdapter{
		user:        NewResourceDecorator[models.User, uuid.UUID, UserFilter](nil),
		permission:  NewResourceDecorator[models.Permission, uuid.UUID, PermissionsFilter](nil),
		userAccount: NewResourceDecorator[models.UserAccount, uuid.UUID, UserAccountFilter](nil),
		token:       NewResourceDecorator[models.Token, uuid.UUID, TokenFilter](nil),
	}
}
