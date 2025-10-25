package services

import (
	"errors"
	"sync"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/conf"
)

var mockPaymentErr = errors.New("this is a test payment client")

type MockPaymentClient struct {
	customerByEmail map[string]*stripe.Customer
	mu              sync.RWMutex
}

func NewMockPaymentClient() *MockPaymentClient {
	return &MockPaymentClient{}
}

// Config implements PaymentClient.
func (t *MockPaymentClient) Config() *conf.StripeConfig {

	return nil
}

// CreateBillingPortalSession implements PaymentClient.
func (t *MockPaymentClient) CreateBillingPortalSession(customerId string, configurationId string) (*stripe.BillingPortalSession, error) {

	return nil, mockPaymentErr
}

// CreateCheckoutSession implements PaymentClient.
func (t *MockPaymentClient) CreateCheckoutSession(customerId string, priceId string, quantity int64, trialDays *int64) (*stripe.CheckoutSession, error) {

	return nil, mockPaymentErr
}

// CreateCustomer implements PaymentClient.
func (t *MockPaymentClient) CreateCustomer(email string, name *string) (*stripe.Customer, error) {
	t.mu.Unlock()
	defer t.mu.Lock()
	var nameString string = "name"
	if name != nil {
		nameString = *name
	}
	customer := &stripe.Customer{
		Email: email,
		Name:  nameString,
	}
	t.customerByEmail[email] = customer
	return customer, nil
}

func (s *MockPaymentClient) GetSavedCustomerByEmail(key string) *stripe.Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.customerByEmail[key]
}

// CreatePortalConfiguration implements PaymentClient.
func (t *MockPaymentClient) CreatePortalConfiguration(input ...*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams) (string, error) {

	return "", mockPaymentErr
}

// FindAllPrices implements PaymentClient.
func (t *MockPaymentClient) FindAllPrices() ([]*stripe.Price, error) {

	return nil, mockPaymentErr
}

// FindAllProducts implements PaymentClient.
func (t *MockPaymentClient) FindAllProducts() ([]*stripe.Product, error) {

	return nil, mockPaymentErr
}

// FindCheckoutSessionByStripeId implements PaymentClient.
func (t *MockPaymentClient) FindCheckoutSessionByStripeId(stripeId string) (*stripe.CheckoutSession, error) {

	return nil, mockPaymentErr
}

// FindOrCreateCustomer implements PaymentClient.
func (t *MockPaymentClient) FindOrCreateCustomer(email string, name *string) (*stripe.Customer, error) {

	return nil, mockPaymentErr
}

// FindSubscriptionByStripeId implements PaymentClient.
func (t *MockPaymentClient) FindSubscriptionByStripeId(stripeId string) (*stripe.Subscription, error) {

	return nil, mockPaymentErr
}

// UpdateCustomer implements PaymentClient.
func (t *MockPaymentClient) UpdateCustomer(customerId string, params *stripe.CustomerParams) (*stripe.Customer, error) {

	return nil, mockPaymentErr
}

// UpdateItemQuantity implements PaymentClient.
func (t *MockPaymentClient) UpdateItemQuantity(itemId string, priceId string, count int64) (*stripe.SubscriptionItem, error) {

	return nil, mockPaymentErr
}

var _ PaymentClient = &MockPaymentClient{}
