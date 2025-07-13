package stores

import (
	"context"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/models"
)

type StripePriceStoreDecorator struct {
	Delegate                   *DbPriceStore
	CountPricesFunc            func(ctx context.Context, filter *StripePriceFilter) (int64, error)
	FindActivePriceByIdFunc    func(ctx context.Context, priceId string) (*models.StripePrice, error)
	FindPriceFunc              func(ctx context.Context, filter *StripePriceFilter) (*models.StripePrice, error)
	ListPricesFunc             func(ctx context.Context, input *StripePriceFilter) ([]*models.StripePrice, error)
	UpsertPriceFunc            func(ctx context.Context, price *models.StripePrice) error
	UpsertPriceFromStripeFunc  func(ctx context.Context, price *stripe.Price) error
	LoadPricesByProductIdsFunc func(ctx context.Context, productIds ...string) ([][]*models.StripePrice, error)
	LoadPricesByIdsFunc        func(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error)
}

func (s *StripePriceStoreDecorator) LoadPricesByIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error) {
	if s.LoadPricesByIdsFunc != nil {
		return s.LoadPricesByIdsFunc(ctx, priceIds...)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.LoadPricesByIds(ctx, priceIds...)
}

// LoadPricesByProductIds implements DbPriceStoreInterface.
func (s *StripePriceStoreDecorator) LoadPricesByProductIds(ctx context.Context, productIds ...string) ([][]*models.StripePrice, error) {
	if s.LoadPricesByProductIdsFunc != nil {
		return s.LoadPricesByProductIdsFunc(ctx, productIds...)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.LoadPricesByProductIds(ctx, productIds...)
}

func (s *StripePriceStoreDecorator) Cleanup() {
	s.CountPricesFunc = nil
	s.FindActivePriceByIdFunc = nil
	s.FindPriceFunc = nil
	s.ListPricesFunc = nil
	s.UpsertPriceFunc = nil
	s.UpsertPriceFromStripeFunc = nil

}

// CountPrices implements DbPriceStoreInterface.
func (s *StripePriceStoreDecorator) CountPrices(ctx context.Context, filter *StripePriceFilter) (int64, error) {
	if s.CountPricesFunc != nil {
		return s.CountPricesFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return s.Delegate.CountPrices(ctx, filter)
}

// FindActivePriceById implements DbPriceStoreInterface.

// FindPrice implements DbPriceStoreInterface.
func (s *StripePriceStoreDecorator) FindPrice(ctx context.Context, filter *StripePriceFilter) (*models.StripePrice, error) {
	if s.FindPriceFunc != nil {
		return s.FindPriceFunc(ctx, filter)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.FindPrice(ctx, filter)
}

// ListPrices implements DbPriceStoreInterface.
func (s *StripePriceStoreDecorator) ListPrices(ctx context.Context, input *StripePriceFilter) ([]*models.StripePrice, error) {
	if s.ListPricesFunc != nil {
		return s.ListPricesFunc(ctx, input)
	}
	if s.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return s.Delegate.ListPrices(ctx, input)
}

// UpsertPrice implements DbPriceStoreInterface.
func (s *StripePriceStoreDecorator) UpsertPrice(ctx context.Context, price *models.StripePrice) error {
	if s.UpsertPriceFunc != nil {
		return s.UpsertPriceFunc(ctx, price)
	}
	if s.Delegate == nil {
		return ErrDelegateNil
	}
	return s.Delegate.UpsertPrice(ctx, price)
}

// UpsertPriceFromStripe implements DbPriceStoreInterface.
func (s *StripePriceStoreDecorator) UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error {
	if s.UpsertPriceFromStripeFunc != nil {
		return s.UpsertPriceFromStripeFunc(ctx, price)
	}
	if s.Delegate == nil {
		return ErrDelegateNil
	}
	return s.Delegate.UpsertPriceFromStripe(ctx, price)
}

var _ DbPriceStoreInterface = (*StripePriceStoreDecorator)(nil)
