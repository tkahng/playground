package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type mockPaymentStore struct{ mock.Mock }

// CreateProductRoles implements PaymentStore.
func (m *mockPaymentStore) CreateProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	args := m.Called(ctx, productId, roleIds)
	return args.Error(0)
}

// CountCustomers implements PaymentStore.
func (m *mockPaymentStore) CountCustomers(ctx context.Context, filter *shared.StripeCustomerListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountPrices implements PaymentStore.
func (m *mockPaymentStore) CountPrices(ctx context.Context, filter *shared.StripePriceListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountSubscriptions implements PaymentStore.
func (m *mockPaymentStore) CountSubscriptions(ctx context.Context, filter *shared.StripeSubscriptionListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// ListCustomers implements PaymentStore.
func (m *mockPaymentStore) ListCustomers(ctx context.Context, input *shared.StripeCustomerListParams) ([]*models.StripeCustomer, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeCustomer), args.Error(1)
	}
	return nil, args.Error(1)
}

// ListSubscriptions implements PaymentStore.
func (m *mockPaymentStore) ListSubscriptions(ctx context.Context, input *shared.StripeSubscriptionListParams) ([]*models.StripeSubscription, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// LoadProductRoles implements PaymentStore.
func (m *mockPaymentStore) LoadProductRoles(ctx context.Context, productIds ...string) ([][]*models.Role, error) {
	args := m.Called(ctx, productIds)
	if args.Get(0) != nil {
		return args.Get(0).([][]*models.Role), args.Error(1)
	}
	return nil, args.Error(1)
}

// CountProducts implements PaymentStore.
func (m *mockPaymentStore) CountProducts(ctx context.Context, filter *shared.StripeProductListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// LoadProductPrices implements PaymentStore.
func (m *mockPaymentStore) LoadProductPrices(ctx context.Context, where *map[string]any, productIds ...string) ([][]*models.StripePrice, error) {
	args := m.Called(ctx, where, productIds)
	if args.Get(0) != nil {
		return args.Get(0).([][]*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpsertCustomerStripeId implements PaymentStore.
func (m *mockPaymentStore) UpsertCustomerStripeId(ctx context.Context, customer *models.StripeCustomer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

// CreateCustomer implements PaymentStore.
func (m *mockPaymentStore) CreateCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	args := m.Called(ctx, customer)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeCustomer), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindCustomer implements PaymentStore.
func (m *mockPaymentStore) FindCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	args := m.Called(ctx, customer)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeCustomer), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreateProductPermissions implements PaymentStore.
func (m *mockPaymentStore) CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error {
	args := m.Called(ctx, productId, permissionIds)
	return args.Error(0)
}

// FindLatestActiveSubscriptionWithPriceByCustomerId implements PaymentStore.
func (m *mockPaymentStore) FindLatestActiveSubscriptionWithPriceByCustomerId(ctx context.Context, customerId string) (*models.SubscriptionWithPrice, error) {
	args := m.Called(ctx, customerId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.SubscriptionWithPrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindPermissionByName implements PaymentStore.
func (m *mockPaymentStore) FindPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	args := m.Called(ctx, name)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindProductByStripeId implements PaymentStore.
func (m *mockPaymentStore) FindProductByStripeId(ctx context.Context, productId string) (*models.StripeProduct, error) {
	args := m.Called(ctx, productId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeProduct), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindSubscriptionWithPriceById implements PaymentStore.
func (m *mockPaymentStore) FindSubscriptionWithPriceById(ctx context.Context, stripeId string) (*models.SubscriptionWithPrice, error) {
	args := m.Called(ctx, stripeId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.SubscriptionWithPrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindTeamByStripeCustomerId implements PaymentStore.
func (m *mockPaymentStore) FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error) {
	args := m.Called(ctx, stripeCustomerId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Team), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindValidPriceById implements PaymentStore.
func (m *mockPaymentStore) FindValidPriceById(ctx context.Context, priceId string) (*models.StripePrice, error) {
	args := m.Called(ctx, priceId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// IsFirstSubscription implements PaymentStore.
func (m *mockPaymentStore) IsFirstSubscription(ctx context.Context, customerId string) (bool, error) {
	args := m.Called(ctx, customerId)
	return args.Get(0).(bool), args.Error(1)
}

// ListPrices implements PaymentStore.
func (m *mockPaymentStore) ListPrices(ctx context.Context, input *shared.StripePriceListParams) ([]*models.StripePrice, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// ListProducts implements PaymentStore.
func (m *mockPaymentStore) ListProducts(ctx context.Context, input *shared.StripeProductListParams) ([]*models.StripeProduct, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeProduct), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpsertPrice implements PaymentStore.
func (m *mockPaymentStore) UpsertPrice(ctx context.Context, price *models.StripePrice) error {
	args := m.Called(ctx, price)
	return args.Error(0)
}

// UpsertPriceFromStripe implements PaymentStore.
func (m *mockPaymentStore) UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error {
	args := m.Called(ctx, price)
	return args.Error(0)
}

// UpsertProduct implements PaymentStore.
func (m *mockPaymentStore) UpsertProduct(ctx context.Context, product *models.StripeProduct) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

// UpsertProductFromStripe implements PaymentStore.
func (m *mockPaymentStore) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

// UpsertSubscription implements PaymentStore.
func (m *mockPaymentStore) UpsertSubscription(ctx context.Context, sub *models.StripeSubscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// UpsertSubscriptionFromStripe implements PaymentStore.
func (m *mockPaymentStore) UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}
func (m *mockPaymentStore) FindLatestActiveSubscriptionByTeamId(ctx context.Context, teamId uuid.UUID) (*models.StripeSubscription, error) {
	args := m.Called(ctx, teamId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockPaymentStore) CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	args := m.Called(ctx, teamId)
	return args.Get(0).(int64), args.Error(1)
}

var _ PaymentStore = (*mockPaymentStore)(nil)
