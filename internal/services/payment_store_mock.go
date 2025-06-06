package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type MockPaymentStore struct{ mock.Mock }

// FindActiveSubscriptionsByCustomerIds implements PaymentStore.
func (m *MockPaymentStore) FindActiveSubscriptionsByCustomerIds(ctx context.Context, customerIds ...string) ([]*models.StripeSubscription, error) {
	args := m.Called(ctx, customerIds)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindActiveSubscriptionsByTeamIds implements PaymentStore.
func (m *MockPaymentStore) FindActiveSubscriptionsByTeamIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.StripeSubscription, error) {
	args := m.Called(ctx, teamIds)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindActiveSubscriptionsByUserIds implements PaymentStore.
func (m *MockPaymentStore) FindActiveSubscriptionsByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.StripeSubscription, error) {
	args := m.Called(ctx, userIds)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// LoadPricesByIds implements PaymentStore.
func (m *MockPaymentStore) LoadPricesByIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error) {
	args := m.Called(ctx, priceIds)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// LoadPricesWithProductByPriceIds implements PaymentStore.
func (m *MockPaymentStore) LoadPricesWithProductByPriceIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error) {
	args := m.Called(ctx, priceIds)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// LoadProductsByIds implements PaymentStore.
func (m *MockPaymentStore) LoadProductsByIds(ctx context.Context, productIds ...string) ([]*models.StripeProduct, error) {
	args := m.Called(ctx, productIds)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeProduct), args.Error(1)
	}
	return nil, args.Error(1)
}

// LoadSubscriptionsByIds implements PaymentStore.
func (m *MockPaymentStore) LoadSubscriptionsByIds(ctx context.Context, subscriptionIds ...string) ([]*models.StripeSubscription, error) {
	args := m.Called(ctx, subscriptionIds)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// LoadSubscriptionsPriceProduct implements PaymentStore.
func (m *MockPaymentStore) LoadSubscriptionsPriceProduct(ctx context.Context, subscriptions ...*models.StripeSubscription) error {
	args := m.Called(ctx, subscriptions)
	return args.Error(0)
}

// FindSubscriptionsWithPriceProductByIds implements PaymentStore.
func (m *MockPaymentStore) FindSubscriptionsWithPriceProductByIds(ctx context.Context, subscriptionIds ...string) ([]*models.StripeSubscription, error) {
	args := m.Called(ctx, subscriptionIds)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindSubscriptionById implements PaymentStore.
func (m *MockPaymentStore) FindSubscriptionById(ctx context.Context, subscriptionId string) (*models.StripeSubscription, error) {
	args := m.Called(ctx, subscriptionId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

var _ PaymentStore = (*MockPaymentStore)(nil)

// LoadProductPermissions implements PaymentStore.
func (m *MockPaymentStore) LoadProductPermissions(ctx context.Context, productIds ...string) ([][]*models.Permission, error) {
	args := m.Called(ctx, productIds)
	if args.Get(0) != nil {
		return args.Get(0).([][]*models.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreateProductRoles implements PaymentStore.
func (m *MockPaymentStore) CreateProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	args := m.Called(ctx, productId, roleIds)
	return args.Error(0)
}

// CountCustomers implements PaymentStore.
func (m *MockPaymentStore) CountCustomers(ctx context.Context, filter *shared.StripeCustomerListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountPrices implements PaymentStore.
func (m *MockPaymentStore) CountPrices(ctx context.Context, filter *shared.StripePriceListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// CountSubscriptions implements PaymentStore.
func (m *MockPaymentStore) CountSubscriptions(ctx context.Context, filter *shared.StripeSubscriptionListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// ListCustomers implements PaymentStore.
func (m *MockPaymentStore) ListCustomers(ctx context.Context, input *shared.StripeCustomerListParams) ([]*models.StripeCustomer, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeCustomer), args.Error(1)
	}
	return nil, args.Error(1)
}

// ListSubscriptions implements PaymentStore.
func (m *MockPaymentStore) ListSubscriptions(ctx context.Context, input *shared.StripeSubscriptionListParams) ([]*models.StripeSubscription, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// LoadProductRoles implements PaymentStore.
func (m *MockPaymentStore) LoadProductRoles(ctx context.Context, productIds ...string) ([][]*models.Role, error) {
	args := m.Called(ctx, productIds)
	if args.Get(0) != nil {
		return args.Get(0).([][]*models.Role), args.Error(1)
	}
	return nil, args.Error(1)
}

// CountProducts implements PaymentStore.
func (m *MockPaymentStore) CountProducts(ctx context.Context, filter *shared.StripeProductListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// LoadPricesByProductIds implements PaymentStore.
func (m *MockPaymentStore) LoadPricesByProductIds(ctx context.Context, productIds ...string) ([][]*models.StripePrice, error) {
	args := m.Called(ctx, productIds)
	if args.Get(0) != nil {
		return args.Get(0).([][]*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpsertCustomerStripeId implements PaymentStore.
func (m *MockPaymentStore) UpsertCustomerStripeId(ctx context.Context, customer *models.StripeCustomer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

// CreateCustomer implements PaymentStore.
func (m *MockPaymentStore) CreateCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	args := m.Called(ctx, customer)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeCustomer), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindCustomer implements PaymentStore.
func (m *MockPaymentStore) FindCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error) {
	args := m.Called(ctx, customer)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeCustomer), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreateProductPermissions implements PaymentStore.
func (m *MockPaymentStore) CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error {
	args := m.Called(ctx, productId, permissionIds)
	return args.Error(0)
}

// FindActiveSubscriptionByCustomerId implements PaymentStore.
func (m *MockPaymentStore) FindActiveSubscriptionByCustomerId(ctx context.Context, customerId string) (*models.StripeSubscription, error) {
	args := m.Called(ctx, customerId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindPermissionByName implements PaymentStore.
func (m *MockPaymentStore) FindPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	args := m.Called(ctx, name)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Permission), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindProductById implements PaymentStore.
func (m *MockPaymentStore) FindProductById(ctx context.Context, productId string) (*models.StripeProduct, error) {
	args := m.Called(ctx, productId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripeProduct), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindTeamByStripeCustomerId implements PaymentStore.
func (m *MockPaymentStore) FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error) {
	args := m.Called(ctx, stripeCustomerId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Team), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindActivePriceById implements PaymentStore.
func (m *MockPaymentStore) FindActivePriceById(ctx context.Context, priceId string) (*models.StripePrice, error) {
	args := m.Called(ctx, priceId)
	if args.Get(0) != nil {
		return args.Get(0).(*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// IsFirstSubscription implements PaymentStore.
func (m *MockPaymentStore) IsFirstSubscription(ctx context.Context, customerId string) (bool, error) {
	args := m.Called(ctx, customerId)
	return args.Get(0).(bool), args.Error(1)
}

// ListPrices implements PaymentStore.
func (m *MockPaymentStore) ListPrices(ctx context.Context, input *shared.StripePriceListParams) ([]*models.StripePrice, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripePrice), args.Error(1)
	}
	return nil, args.Error(1)
}

// ListProducts implements PaymentStore.
func (m *MockPaymentStore) ListProducts(ctx context.Context, input *shared.StripeProductListParams) ([]*models.StripeProduct, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.StripeProduct), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpsertPrice implements PaymentStore.
func (m *MockPaymentStore) UpsertPrice(ctx context.Context, price *models.StripePrice) error {
	args := m.Called(ctx, price)
	return args.Error(0)
}

// UpsertPriceFromStripe implements PaymentStore.
func (m *MockPaymentStore) UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error {
	args := m.Called(ctx, price)
	return args.Error(0)
}

// UpsertProduct implements PaymentStore.
func (m *MockPaymentStore) UpsertProduct(ctx context.Context, product *models.StripeProduct) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

// UpsertProductFromStripe implements PaymentStore.
func (m *MockPaymentStore) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

// UpsertSubscription implements PaymentStore.
func (m *MockPaymentStore) UpsertSubscription(ctx context.Context, sub *models.StripeSubscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// UpsertSubscriptionFromStripe implements PaymentStore.
func (m *MockPaymentStore) UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockPaymentStore) CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	args := m.Called(ctx, teamId)
	return args.Get(0).(int64), args.Error(1)
}
