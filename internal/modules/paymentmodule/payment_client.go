package paymentmodule

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/billingportal/configuration"
	bs "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/price"
	"github.com/stripe/stripe-go/v82/product"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/tkahng/authgo/internal/conf"
)

type StripeClient struct {
	config *conf.StripeConfig
}

func NewPaymentClient(bld conf.StripeConfig) *StripeClient {
	stripe.Key = bld.ApiKey
	payment := &StripeClient{config: &bld}
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

func (client *StripeClient) findCustomerByEmailAndUserId(email string, userId string) (*stripe.Customer, error) {
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

func (s *StripeClient) FindOrCreateCustomer(email string, userId uuid.UUID) (*stripe.Customer, error) {
	cs, _ := s.findCustomerByEmailAndUserId(email, userId.String())
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

type ProductBillingConfigurationInput struct {
	// The list of price IDs for the product that a subscription can be updated to.
	Prices []*string `form:"prices"`
	// The product id.
	Product *string `form:"product"`
}

func (a *StripeClient) CreatePortalConfiguration(input ...*ProductBillingConfigurationInput) (string, error) {
	var prods []*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams
	for _, i := range input {
		prods = append(prods, &stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams{
			Product: i.Product,
			Prices:  i.Prices,
		})
	}
	config := &stripe.BillingPortalConfigurationParams{
		BusinessProfile: &stripe.BillingPortalConfigurationBusinessProfileParams{
			Headline: stripe.String("Manage your subscription"),
		},
		Features: &stripe.BillingPortalConfigurationFeaturesParams{
			SubscriptionUpdate: &stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateParams{
				Enabled:               stripe.Bool(true),
				ProrationBehavior:     stripe.String("create_prorations"),
				DefaultAllowedUpdates: stripe.StringSlice([]string{"price", "promotion_code"}),
				Products:              prods,
			},
			SubscriptionCancel: &stripe.BillingPortalConfigurationFeaturesSubscriptionCancelParams{
				Enabled: stripe.Bool(true),
				Mode:    stripe.String("at_period_end"),
				CancellationReason: &stripe.BillingPortalConfigurationFeaturesSubscriptionCancelCancellationReasonParams{
					Enabled: stripe.Bool(true),
					Options: stripe.StringSlice([]string{
						"too_expensive",
						"missing_features",
						"switched_service",
						"unused",
						"other",
					}),
				},
			},
			PaymentMethodUpdate: &stripe.BillingPortalConfigurationFeaturesPaymentMethodUpdateParams{
				Enabled: stripe.Bool(true),
			},
		},
	}
	result, err := configuration.New(config)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", errors.New("failed to create billing portal configuration")
	}

	return result.ID, nil
}

func (s *StripeClient) CreateBillingPortalSession(customerId string, configurationId string) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Configuration: stripe.String(configurationId),
		Customer:      stripe.String(customerId),
		ReturnURL:     stripe.String(s.config.StripeAppUrl + "/settings/billing"),
	}
	return bs.New(params)
}
