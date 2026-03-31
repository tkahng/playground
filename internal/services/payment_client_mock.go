package services

import (
	"errors"
	"fmt"

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
)
var (
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
)

var (
	PointsProduct = &stripe.Product{
		ID:     "prod_points_100",
		Name:   "100 Points",
		Active: true,
	}
	PointsPrice100 = &stripe.Price{
		ID:         "price_points_100",
		Product:    PointsProduct,
		Active:     true,
		Currency:   stripe.CurrencyUSD,
		Type:       stripe.PriceTypeOneTime,
		UnitAmount: 99,
		Metadata: map[string]string{
			"points_amount": "100",
		},
	}
)

var (
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
	SubscriptionItems                    []*stripe.SubscriptionItem
	CustomerByEmail                      map[string]*stripe.Customer
	Customers                            []*stripe.Customer
	ConfigFunc                           func() *conf.StripeConfig
	CreateBillingPortalSessionFunc       func(customerId string, configurationId string, retunrUrl string) (*stripe.BillingPortalSession, error)
	CreateCheckoutSessionFunc            func(customerId string, priceId string, quantity int64, trialDays *int64) (*stripe.CheckoutSession, error)
	CreatePointsCheckoutSessionFunc      func(customerID, userID string, pointsAmount int64, priceID string) (*stripe.CheckoutSession, error)
	CreateCustomerFunc                   func(email string, name *string, metadata *map[string]string) (*stripe.Customer, error)
	CreatePortalConfigurationFunc        func(input ...*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams) (string, error)
	FindAllPricesFunc                    func() ([]*stripe.Price, error)
	FindAllProductsFunc                  func() ([]*stripe.Product, error)
	FindCheckoutSessionByStripeIdFunc    func(stripeId string) (*stripe.CheckoutSession, error)
	FindOrCreateCustomerFunc             func(email string, name *string) (*stripe.Customer, error)
	FindSubscriptionByStripeIdFunc       func(stripeId string) (*stripe.Subscription, error)
	UpdateCustomerFunc                   func(customerId string, params *stripe.CustomerParams) (*stripe.Customer, error)
	UpdateItemQuantityFunc               func(itemId string, priceId string, count int64) (*stripe.SubscriptionItem, error)
}

func (t *MockPaymentClient) GetCustomerByFunc(fn func(*stripe.Customer) bool) *stripe.Customer {
	for _, customer := range t.Customers {
		if fn(customer) {
			return customer
		}
	}
	return nil
}

func (t *MockPaymentClient) GetUpdateSubscriptionInput(fn func(*stripe.SubscriptionItem) bool) *stripe.SubscriptionItem {
	for _, input := range t.SubscriptionItems {
		if fn(input) {
			return input
		}
	}
	return nil
}

func NewMockPaymentClient() *MockPaymentClient {
	return &MockPaymentClient{
		CustomerByEmail: make(map[string]*stripe.Customer),
	}
}

// Config implements PaymentClient.
func (t *MockPaymentClient) Config() *conf.StripeConfig {
	if t.ConfigFunc != nil {
		return t.ConfigFunc()
	}
	return nil
}

// CreateBillingPortalSession implements PaymentClient.
func (t *MockPaymentClient) CreateBillingPortalSession(customerId string, configurationId string, retunrUrl string) (*stripe.BillingPortalSession, error) {
	if t.CreateBillingPortalSessionFunc != nil {
		return t.CreateBillingPortalSessionFunc(customerId, configurationId, retunrUrl)
	}
	return nil, mockPaymentErr
}

// CreateCheckoutSession implements PaymentClient.
func (t *MockPaymentClient) CreateCheckoutSession(customerId string, priceId string, quantity int64, trialDays *int64) (*stripe.CheckoutSession, error) {
	if t.CreateCheckoutSessionFunc != nil {
		return t.CreateCheckoutSessionFunc(customerId, priceId, quantity, trialDays)
	}
	return nil, mockPaymentErr
}

// CreatePointsCheckoutSession implements PaymentClient.
func (t *MockPaymentClient) CreatePointsCheckoutSession(customerID, userID string, pointsAmount int64, priceID string) (*stripe.CheckoutSession, error) {
	if t.CreatePointsCheckoutSessionFunc != nil {
		return t.CreatePointsCheckoutSessionFunc(customerID, userID, pointsAmount, priceID)
	}
	return nil, mockPaymentErr
}

func (t *MockPaymentClient) CreateCustomer(email string, name *string, metadata *map[string]string) (*stripe.Customer, error) {
	if t.CreateCustomerFunc != nil {
		return t.CreateCustomerFunc(email, name, metadata)
	}
	var nameString string
	var meta map[string]string
	if name != nil {
		nameString = *name
	} else {
		nameString = ""
	}
	if metadata != nil {
		meta = *metadata
	}
	customer := &stripe.Customer{
		ID:       fmt.Sprintf("cus_%s-%s", email, nameString),
		Email:    email,
		Name:     nameString,
		Metadata: meta,
	}
	t.CustomerByEmail[email] = customer
	t.Customers = append(t.Customers, customer)
	return customer, nil
}

// CreatePortalConfiguration implements PaymentClient.
func (t *MockPaymentClient) CreatePortalConfiguration(input ...*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams) (string, error) {
	if t.CreatePortalConfigurationFunc != nil {
		return t.CreatePortalConfigurationFunc(input...)
	}
	return "", mockPaymentErr
}

// FindAllPrices implements PaymentClient.
func (t *MockPaymentClient) FindAllPrices() ([]*stripe.Price, error) {
	if t.FindAllPricesFunc != nil {
		return t.FindAllPricesFunc()
	}
	return Prices, nil
}

// FindAllProducts implements PaymentClient.
func (t *MockPaymentClient) FindAllProducts() ([]*stripe.Product, error) {
	if t.FindAllProductsFunc != nil {
		return t.FindAllProductsFunc()
	}
	return Products, nil
}

// FindCheckoutSessionByStripeId implements PaymentClient.
func (t *MockPaymentClient) FindCheckoutSessionByStripeId(stripeId string) (*stripe.CheckoutSession, error) {
	if t.FindCheckoutSessionByStripeIdFunc != nil {
		return t.FindCheckoutSessionByStripeIdFunc(stripeId)
	}
	return nil, mockPaymentErr
}

// FindOrCreateCustomer implements PaymentClient.
func (t *MockPaymentClient) FindOrCreateCustomer(email string, name *string) (*stripe.Customer, error) {
	if t.FindOrCreateCustomerFunc != nil {
		return t.FindOrCreateCustomerFunc(email, name)
	}
	for _, customer := range t.CustomerByEmail {
		if customer.Email == email {
			return customer, nil
		}
	}
	return nil, mockPaymentErr
}

// FindSubscriptionByStripeId implements PaymentClient.
func (t *MockPaymentClient) FindSubscriptionByStripeId(stripeId string) (*stripe.Subscription, error) {
	if t.FindSubscriptionByStripeIdFunc != nil {
		return t.FindSubscriptionByStripeIdFunc(stripeId)
	}
	return nil, mockPaymentErr
}

// UpdateCustomer implements PaymentClient.
func (t *MockPaymentClient) UpdateCustomer(customerId string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	if t.UpdateCustomerFunc != nil {
		return t.UpdateCustomerFunc(customerId, params)
	}
	for _, customer := range t.Customers {
		if customer.ID == customerId {
			if params.Name != nil {
				customer.Name = *params.Name
			}
			if params.Email != nil {
				customer.Email = *params.Email
			}
			if params.Description != nil {
				customer.Description = *params.Description
			}
			if params.Phone != nil {
				customer.Phone = *params.Phone
			}
			return customer, nil
		}
	}
	return nil, errors.New("could not find customer")
}

// UpdateItemQuantity implements PaymentClient.
func (t *MockPaymentClient) UpdateItemQuantity(itemId string, priceId string, count int64) (*stripe.SubscriptionItem, error) {
	if t.UpdateItemQuantityFunc != nil {
		return t.UpdateItemQuantityFunc(itemId, priceId, count)
	}
	item := &stripe.SubscriptionItem{
		ID: itemId,
		Price: &stripe.Price{
			ID: priceId,
		},
		Quantity: count,
	}

	t.SubscriptionItems = append(t.SubscriptionItems, item)
	return item, nil
}

var _ PaymentClient = &MockPaymentClient{}
