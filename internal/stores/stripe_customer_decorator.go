package stores

import (
	"context"

	"github.com/tkahng/playground/internal/models"
)

type CustomerStoreDecorator struct {
	Delegate               *DbCustomerStore
	CountCustomersFunc     func(ctx context.Context, filter *StripeCustomerFilter) (int64, error)
	FindCustomerFunc       func(ctx context.Context, filter *StripeCustomerFilter) (*models.StripeCustomer, error)
	ListCustomersFunc      func(ctx context.Context, input *StripeCustomerFilter) ([]*models.StripeCustomer, error)
	CreateCustomerFunc     func(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error)
	LoadCustomersByIdsFunc func(ctx context.Context, ids ...string) ([]*models.StripeCustomer, error)
}

// LoadCustomersByIds implements DbCustomerStoreInterface.
func (c *CustomerStoreDecorator) LoadCustomersByIds(ctx context.Context, ids ...string) ([]*models.StripeCustomer, error) {
	if c.LoadCustomersByIdsFunc != nil {
		return c.LoadCustomersByIdsFunc(ctx, ids...)
	}
	if c.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return c.Delegate.LoadCustomersByIds(ctx, ids...)
}

func (c *CustomerStoreDecorator) Cleanup() {
	c.CountCustomersFunc = nil
	c.FindCustomerFunc = nil
	c.ListCustomersFunc = nil
	c.CreateCustomerFunc = nil
}

// CountCustomers implements DbCustomerStoreInterface.
func (c *CustomerStoreDecorator) CountCustomers(ctx context.Context, filter *StripeCustomerFilter) (int64, error) {
	if c.CountCustomersFunc != nil {
		return c.CountCustomersFunc(ctx, filter)
	}
	if c.Delegate == nil {
		return 0, ErrDelegateNil
	}
	return c.Delegate.CountCustomers(ctx, filter)
}

// CreateCustomer implements DbCustomerStoreInterface.
func (c *CustomerStoreDecorator) CreateCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	if c.CreateCustomerFunc != nil {
		return c.CreateCustomerFunc(ctx, customer)
	}
	if c.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return c.Delegate.CreateCustomer(ctx, customer)
}

// FindCustomer implements DbCustomerStoreInterface.
func (c *CustomerStoreDecorator) FindCustomer(ctx context.Context, customer *StripeCustomerFilter) (*models.StripeCustomer, error) {
	if c.FindCustomerFunc != nil {
		return c.FindCustomerFunc(ctx, customer)
	}
	if c.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return c.Delegate.FindCustomer(ctx, customer)
}

// ListCustomers implements DbCustomerStoreInterface.
func (c *CustomerStoreDecorator) ListCustomers(ctx context.Context, input *StripeCustomerFilter) ([]*models.StripeCustomer, error) {
	if c.ListCustomersFunc != nil {
		return c.ListCustomersFunc(ctx, input)
	}
	if c.Delegate == nil {
		return nil, ErrDelegateNil
	}
	return c.Delegate.ListCustomers(ctx, input)
}

// CountCustomers implements DbCustomerStoreInterface.

var _ DbCustomerStoreInterface = (*CustomerStoreDecorator)(nil)
