package services

import (
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/conf"
)

type TestPaymentClient struct {
	ConfigFunc                        func() *conf.StripeConfig
	CreateBillingPortalSessionFunc    func(customerId string, configurationId string) (*stripe.BillingPortalSession, error)
	CreateCheckoutSessionFunc         func(customerId string, priceId string, quantity int64, trialDays *int64) (*stripe.CheckoutSession, error)
	CreateCustomerFunc                func(email string, name *string) (*stripe.Customer, error)
	CreatePortalConfigurationFunc     func(input ...*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams) (string, error)
	FindAllPricesFunc                 func() ([]*stripe.Price, error)
	FindAllProductsFunc               func() ([]*stripe.Product, error)
	FindCheckoutSessionByStripeIdFunc func(stripeId string) (*stripe.CheckoutSession, error)
	FindOrCreateCustomerFunc          func(email string, name *string) (*stripe.Customer, error)
	FindSubscriptionByStripeIdFunc    func(stripeId string) (*stripe.Subscription, error)
	UpdateCustomerFunc                func(customerId string, params *stripe.CustomerParams) (*stripe.Customer, error)
	UpdateItemQuantityFunc            func(itemId string, priceId string, count int64) (*stripe.SubscriptionItem, error)
}

func NewTestPaymentClient() *TestPaymentClient {
	return &TestPaymentClient{}
}

// Config implements PaymentClient.
func (t *TestPaymentClient) Config() *conf.StripeConfig {
	if t.ConfigFunc != nil {
		return t.ConfigFunc()
	}
	return nil
}

// CreateBillingPortalSession implements PaymentClient.
func (t *TestPaymentClient) CreateBillingPortalSession(customerId string, configurationId string) (*stripe.BillingPortalSession, error) {
	if t.CreateBillingPortalSessionFunc != nil {
		return t.CreateBillingPortalSessionFunc(customerId, configurationId)
	}
	return nil, nil
}

// CreateCheckoutSession implements PaymentClient.
func (t *TestPaymentClient) CreateCheckoutSession(customerId string, priceId string, quantity int64, trialDays *int64) (*stripe.CheckoutSession, error) {
	if t.CreateCheckoutSessionFunc != nil {
		return t.CreateCheckoutSessionFunc(customerId, priceId, quantity, trialDays)
	}
	return nil, nil
}

// CreateCustomer implements PaymentClient.
func (t *TestPaymentClient) CreateCustomer(email string, name *string) (*stripe.Customer, error) {
	if t.CreateCustomerFunc != nil {
		return t.CreateCustomerFunc(email, name)
	}
	return nil, nil
}

// CreatePortalConfiguration implements PaymentClient.
func (t *TestPaymentClient) CreatePortalConfiguration(input ...*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams) (string, error) {
	if t.CreatePortalConfigurationFunc != nil {
		return t.CreatePortalConfigurationFunc(input...)
	}
	return "", nil
}

// FindAllPrices implements PaymentClient.
func (t *TestPaymentClient) FindAllPrices() ([]*stripe.Price, error) {
	if t.FindAllPricesFunc != nil {
		return t.FindAllPricesFunc()
	}
	return nil, nil
}

// FindAllProducts implements PaymentClient.
func (t *TestPaymentClient) FindAllProducts() ([]*stripe.Product, error) {
	if t.FindAllProductsFunc != nil {
		return t.FindAllProductsFunc()
	}
	return nil, nil
}

// FindCheckoutSessionByStripeId implements PaymentClient.
func (t *TestPaymentClient) FindCheckoutSessionByStripeId(stripeId string) (*stripe.CheckoutSession, error) {
	if t.FindCheckoutSessionByStripeIdFunc != nil {
		return t.FindCheckoutSessionByStripeIdFunc(stripeId)
	}
	return nil, nil
}

// FindOrCreateCustomer implements PaymentClient.
func (t *TestPaymentClient) FindOrCreateCustomer(email string, name *string) (*stripe.Customer, error) {
	if t.FindOrCreateCustomerFunc != nil {
		return t.FindOrCreateCustomerFunc(email, name)
	}
	return nil, nil
}

// FindSubscriptionByStripeId implements PaymentClient.
func (t *TestPaymentClient) FindSubscriptionByStripeId(stripeId string) (*stripe.Subscription, error) {
	if t.FindSubscriptionByStripeIdFunc != nil {
		return t.FindSubscriptionByStripeIdFunc(stripeId)
	}
	return nil, nil
}

// UpdateCustomer implements PaymentClient.
func (t *TestPaymentClient) UpdateCustomer(customerId string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	if t.UpdateCustomerFunc != nil {
		return t.UpdateCustomerFunc(customerId, params)
	}
	return nil, nil
}

// UpdateItemQuantity implements PaymentClient.
func (t *TestPaymentClient) UpdateItemQuantity(itemId string, priceId string, count int64) (*stripe.SubscriptionItem, error) {
	if t.UpdateItemQuantityFunc != nil {
		return t.UpdateItemQuantityFunc(itemId, priceId, count)
	}
	return nil, nil
}

var _ PaymentClient = &TestPaymentClient{}
