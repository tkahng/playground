package payment

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	bs "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/price"
	"github.com/stripe/stripe-go/v82/product"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/tkahng/authgo/internal/conf"
)

const (
	StripeProductProID      string = "prod_pro"
	StripeProductAdvancedID string = "prod_advanced"
)

type StripeClient struct {
	config *conf.StripeConfig
}

func NewStripeClient(bld conf.StripeConfig) *StripeClient {
	cfg := &bld
	stripe.Key = cfg.ApiKey
	payment := &StripeClient{config: cfg}
	return payment
}

func (s *StripeClient) Config() *conf.StripeConfig {
	return s.config
}

func (s *StripeClient) CreateCustomer(email string, userId string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"user_id": userId,
		},
	}
	return customer.New(params)
}

// method to find customer
func (s *StripeClient) FindCustomerByEmail(email string) (*stripe.Customer, error) {
	var cs *stripe.Customer
	params := &stripe.CustomerListParams{
		Email: stripe.String(email),
		ListParams: stripe.ListParams{
			Limit: stripe.Int64(1),
		},
	}
	list := customer.List(params)
	for list.Next() {
		cs = list.Customer()
		break
	}

	return cs, nil
}

func (client *StripeClient) FindCustomerByEmailAndUserId(email string, userId string) (*stripe.Customer, error) {
	var cs *stripe.Customer
	params := &stripe.CustomerSearchParams{
		SearchParams: stripe.SearchParams{
			Query: fmt.Sprintf("email:'%s' AND metadata['user_id']:'%s'", email, userId),
			Limit: stripe.Int64(1),
		},
	}
	result := customer.Search(params)

	for result.Next() {
		cs = result.Customer()
		break
	}
	return cs, nil
}

func (s *StripeClient) FindAllCustomers() ([]*stripe.Customer, error) {
	var data []*stripe.Customer
	params := &stripe.CustomerListParams{}
	list := customer.List(params)
	for list.Next() {
		prod := list.Customer()
		if prod != nil {
			data = append(data, prod)
		}

	}

	return data, nil
}

func (s *StripeClient) FindAllProducts() ([]*stripe.Product, error) {
	var data []*stripe.Product
	params := &stripe.ProductListParams{}
	list := product.List(params)
	for list.Next() {
		prod := list.Product()
		if prod != nil {
			data = append(data, prod)
		}

	}

	return data, nil
}

func (s *StripeClient) FindAllPrices() ([]*stripe.Price, error) {
	var data []*stripe.Price
	params := &stripe.PriceListParams{}
	list := price.List(params)
	for list.Next() {
		prod := list.Price()
		if prod != nil {
			data = append(data, prod)
		}

	}

	return data, nil
}

func (s *StripeClient) FindAllSubscriptions() ([]*stripe.Subscription, error) {
	var data []*stripe.Subscription
	params := &stripe.SubscriptionListParams{}
	list := subscription.List(params)
	for list.Next() {
		prod := list.Subscription()
		if prod != nil {
			data = append(data, prod)
		}

	}

	return data, nil
}

func (s *StripeClient) FindCustomerByStripeId(stripeId string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{}
	return customer.Get(stripeId, params)
}

func (s *StripeClient) FindOrCreateCustomer(email string, userId uuid.UUID) (*stripe.Customer, error) {
	cs, _ := s.FindCustomerByEmailAndUserId(email, userId.String())
	if cs == nil {
		return s.CreateCustomer(email, userId.String())
	}
	return cs, nil
}

// find stripe subscription by stripe id
func (s *StripeClient) FindSubscriptionByStripeId(stripeId string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{}
	params.AddExpand("default_payment_method")
	return subscription.Get(stripeId, params)
}

// find stripe subscription by stripe id
func (s *StripeClient) FindCheckoutSessionByStripeId(stripeId string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{}
	return session.Get(stripeId, params)
}

func (s *StripeClient) CreateCheckoutSession(customerId, priceId string, trialDays *int64) (*stripe.CheckoutSession, error) {
	lineParams := []*stripe.CheckoutSessionLineItemParams{
		{
			Price:    stripe.String(priceId),
			Quantity: stripe.Int64(1),
		},
	}
	customerUpdateParams := &stripe.CheckoutSessionCustomerUpdateParams{
		Address: stripe.String("auto"),
	}
	subscriptionParams := &stripe.CheckoutSessionSubscriptionDataParams{
		Metadata:        map[string]string{},
		TrialPeriodDays: trialDays,
	}

	sessionParams := &stripe.CheckoutSessionParams{
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
		Customer:           stripe.String(customerId),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		CustomerUpdate:     customerUpdateParams,
		Mode:               stripe.String("subscription"),
		SuccessURL:         stripe.String(s.config.StripeAppUrl + "/payment/success?sessionId={CHECKOUT_SESSION_ID}"),
		// CancelURL:          stripe.String(s.config.AppUrl + "/payment/cancel"),
		LineItems:        lineParams,
		SubscriptionData: subscriptionParams,
	}
	return session.New(sessionParams)
}

func (s *StripeClient) CreateBillingPortalSession(customerId string) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerId),
		ReturnURL: stripe.String(s.config.StripeAppUrl + "/settings/billing"),
	}
	return bs.New(params)
}
