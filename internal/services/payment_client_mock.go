package services

import (
	"errors"
	"sync"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/conf"
)

var (
	ProProduct = &stripe.Product{
		ID:          "prod_pro",
		Name:        "Pro",
		Description: "Pro product description",
		Active:      true,
		Metadata: map[string]string{
			"index": "1",
		},
	}
	AdvancedProduct = &stripe.Product{
		ID:          "prod_advanced",
		Name:        "Advanced",
		Description: "Advanced product description",
		Active:      true,
		Metadata: map[string]string{
			"index": "2",
		},
	}
	ProMonthlyPrice = &stripe.Price{
		ID:            "price_pro_month_usd_5000",
		Product:       ProProduct,
		Active:        true,
		Currency:      stripe.CurrencyUSD,
		Type:          stripe.PriceTypeRecurring,
		BillingScheme: stripe.PriceBillingSchemePerUnit,
		UnitAmount:    5000,
		Recurring: &stripe.PriceRecurring{
			Interval:      stripe.PriceRecurringIntervalMonth,
			IntervalCount: 1,
		},
		Metadata: map[string]string{
			"index": "1",
		},
	}
	ProYearlyPrice = &stripe.Price{
		ID:            "price_pro_year_usd_50000",
		Product:       ProProduct,
		Active:        true,
		Currency:      stripe.CurrencyUSD,
		BillingScheme: stripe.PriceBillingSchemePerUnit,
		Type:          stripe.PriceTypeRecurring,
		UnitAmount:    50000,
		Recurring: &stripe.PriceRecurring{
			Interval:      stripe.PriceRecurringIntervalYear,
			IntervalCount: 1,
		},
		Metadata: map[string]string{
			"index": "2",
		},
	}
	AdvancedMonthlyPrice = &stripe.Price{
		ID:            "price_advanced_month_usd_8500",
		Product:       AdvancedProduct,
		Active:        true,
		Currency:      stripe.CurrencyUSD,
		BillingScheme: stripe.PriceBillingSchemePerUnit,
		Type:          stripe.PriceTypeRecurring,
		UnitAmount:    8500,
		Recurring: &stripe.PriceRecurring{
			Interval:      stripe.PriceRecurringIntervalYear,
			IntervalCount: 1,
		},
		Metadata: map[string]string{
			"index": "3",
		},
	}
	AdvancedYearlyPrice = &stripe.Price{
		ID:            "price_advanced_year_usd_85000",
		Product:       AdvancedProduct,
		Active:        true,
		Currency:      stripe.CurrencyUSD,
		BillingScheme: stripe.PriceBillingSchemePerUnit,
		Type:          stripe.PriceTypeRecurring,
		UnitAmount:    85000,
		Recurring: &stripe.PriceRecurring{
			Interval:      stripe.PriceRecurringIntervalYear,
			IntervalCount: 1,
		},
		Metadata: map[string]string{
			"index": "4",
		},
	}
	Products = []*stripe.Product{
		ProProduct,
		AdvancedProduct,
	}
	Prices = []*stripe.Price{
		ProMonthlyPrice,
		ProYearlyPrice,
		AdvancedMonthlyPrice,
		AdvancedYearlyPrice,
	}
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
