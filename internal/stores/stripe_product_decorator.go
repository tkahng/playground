package stores

import (
	"context"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/models"
)

type StripeProductStoreDecorator struct {
	Delegate                    *DbProductStore
	CountProductsFunc           func(ctx context.Context, filter *StripeProductFilter) (int64, error)
	FindProductFunc             func(ctx context.Context, filter *StripeProductFilter) (*models.StripeProduct, error)
	FindProductByIdFunc         func(ctx context.Context, productId string) (*models.StripeProduct, error)
	ListProductsFunc            func(ctx context.Context, input *StripeProductFilter) ([]*models.StripeProduct, error)
	UpsertProductFunc           func(ctx context.Context, product *models.StripeProduct) error
	UpsertProductFromStripeFunc func(ctx context.Context, product *stripe.Product) error
	LoadProductsByIdsFunc       func(ctx context.Context, productIds ...string) ([]*models.StripeProduct, error)
}

func (s *StripeProductStoreDecorator) Cleanup() {
	s.CountProductsFunc = nil
	s.FindProductFunc = nil
	s.FindProductByIdFunc = nil
	s.ListProductsFunc = nil
	s.UpsertProductFunc = nil
	s.UpsertProductFromStripeFunc = nil
}

func (s *StripeProductStoreDecorator) LoadProductsByIds(ctx context.Context, productIds ...string) ([]*models.StripeProduct, error) {
	if s.LoadProductsByIdsFunc != nil {
		return s.LoadProductsByIdsFunc(ctx, productIds...)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.LoadProductsByIds(ctx, productIds...)
}

// CountProducts implements DbProductStoreInterface.
func (s *StripeProductStoreDecorator) CountProducts(ctx context.Context, filter *StripeProductFilter) (int64, error) {
	if s.CountProductsFunc != nil {
		return s.CountProductsFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return s.Delegate.CountProducts(ctx, filter)
}

// FindProduct implements DbProductStoreInterface.
func (s *StripeProductStoreDecorator) FindProduct(ctx context.Context, filter *StripeProductFilter) (*models.StripeProduct, error) {
	if s.FindProductFunc != nil {
		return s.FindProductFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.FindProduct(ctx, filter)
}

// FindProductById implements DbProductStoreInterface.
func (s *StripeProductStoreDecorator) FindProductById(ctx context.Context, productId string) (*models.StripeProduct, error) {
	if s.FindProductByIdFunc != nil {
		return s.FindProductByIdFunc(ctx, productId)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.FindProductById(ctx, productId)
}

// ListProducts implements DbProductStoreInterface.
func (s *StripeProductStoreDecorator) ListProducts(ctx context.Context, input *StripeProductFilter) ([]*models.StripeProduct, error) {
	if s.ListProductsFunc != nil {
		return s.ListProductsFunc(ctx, input)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.ListProducts(ctx, input)
}

// UpsertProduct implements DbProductStoreInterface.
func (s *StripeProductStoreDecorator) UpsertProduct(ctx context.Context, product *models.StripeProduct) error {
	if s.UpsertProductFunc != nil {
		return s.UpsertProductFunc(ctx, product)
	}
	if s.Delegate == nil {
		return ErrDelegateNil
	}
	return s.Delegate.UpsertProduct(ctx, product)
}

// UpsertProductFromStripe implements DbProductStoreInterface.
func (s *StripeProductStoreDecorator) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
	if s.UpsertProductFromStripeFunc != nil {
		return s.UpsertProductFromStripeFunc(ctx, product)
	}
	if s.Delegate == nil {
		return ErrDelegateNil
	}
	return s.Delegate.UpsertProductFromStripe(ctx, product)
}

var _ DbProductStoreInterface = (*StripeProductStoreDecorator)(nil)
