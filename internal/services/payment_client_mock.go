package services

import (
	"github.com/stretchr/testify/mock"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/conf"
)

type mockPaymentClient struct{ mock.Mock }

// Config implements PaymentClient.
func (m *mockPaymentClient) Config() *conf.StripeConfig {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).(*conf.StripeConfig)
	}
	return nil
}

// CreateBillingPortalSession implements PaymentClient.
func (m *mockPaymentClient) CreateBillingPortalSession(customerId string, configurationId string) (*stripe.BillingPortalSession, error) {
	args := m.Called(customerId, configurationId)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.BillingPortalSession), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreateCheckoutSession implements PaymentClient.
func (m *mockPaymentClient) CreateCheckoutSession(customerId string, priceId string, quantity int64, trialDays *int64) (*stripe.CheckoutSession, error) {
	args := m.Called(customerId, priceId, quantity, trialDays)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.CheckoutSession), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreateCustomer implements PaymentClient.
func (m *mockPaymentClient) CreateCustomer(email string, name *string) (*stripe.Customer, error) {
	args := m.Called(email, name)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreatePortalConfiguration implements PaymentClient.
func (m *mockPaymentClient) CreatePortalConfiguration(input ...*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams) (string, error) {
	args := m.Called(input)
	return args.String(0), args.Error(1)
}

// FindAllPrices implements PaymentClient.
func (m *mockPaymentClient) FindAllPrices() ([]*stripe.Price, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*stripe.Price), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindAllProducts implements PaymentClient.
func (m *mockPaymentClient) FindAllProducts() ([]*stripe.Product, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*stripe.Product), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindCheckoutSessionByStripeId implements PaymentClient.
func (m *mockPaymentClient) FindCheckoutSessionByStripeId(stripeId string) (*stripe.CheckoutSession, error) {
	args := m.Called(stripeId)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.CheckoutSession), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindOrCreateCustomer implements PaymentClient.
func (m *mockPaymentClient) FindOrCreateCustomer(email string, name *string) (*stripe.Customer, error) {
	args := m.Called(email, name)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

// FindSubscriptionByStripeId implements PaymentClient.
func (m *mockPaymentClient) FindSubscriptionByStripeId(stripeId string) (*stripe.Subscription, error) {
	args := m.Called(stripeId)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateCustomer implements PaymentClient.
func (m *mockPaymentClient) UpdateCustomer(customerId string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	args := m.Called(customerId, params)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.Customer), args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateItemQuantity implements PaymentClient.
func (m *mockPaymentClient) UpdateItemQuantity(itemId string, priceId string, count int64) (*stripe.SubscriptionItem, error) {
	args := m.Called(itemId, priceId, count)
	if args.Get(0) != nil {
		return args.Get(0).(*stripe.SubscriptionItem), args.Error(1)
	}
	return nil, args.Error(1)
}

var _ PaymentClient = (*mockPaymentClient)(nil)
