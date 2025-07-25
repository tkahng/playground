package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/shared"
)

func BindStripeApi(api huma.API, appApi *Api) {
	selectCustomerFromUser := middleware.SelectCustomerFromUser(api, appApi.App())
	selectCustomerFromTeam := middleware.SelectCustomerFromTeam(api, appApi.App())
	selectOrCreateOwnerCustomerFromTeam := middleware.SelectOrCreateOwnerCustomerFromTeam(api, appApi.App())
	teamInfoFromParam := middleware.TeamInfoFromParam(api, appApi.App())
	stripeGroup := huma.NewGroup(api)

	// stripe webhook
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "stripe-webhook",
			Method:      http.MethodPost,
			Path:        "/stripe/webhook",
			Summary:     "webhook",
			Description: "Webhook for stripe",
			Tags:        []string{"Stripe", "Webhook"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		appApi.StripeWebhook,
	)
	// stripe products with prices
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "stripe-products-with-prices",
			Method:      http.MethodGet,
			Path:        "/stripe/products",
			Summary:     "stripe-products-with-prices",
			Description: "stripe-products-with-prices",
			Tags:        []string{"Stripe", "Products"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		appApi.StripeProductsWithPrices,
	)

	//  stripe get checkout session by checkoutSessionId
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "get-checkout-session",
			Method:      http.MethodGet,
			Path:        "/subscriptions/checkout-session/{checkoutSessionId}",
			Summary:     "get checkout session",
			Description: "get checkout session",
			Tags:        []string{"Subscriptions", "Checkout Session"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.StripeCheckoutSessionGet,
	)
	// stripe my subscriptions
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "subscriptions-active",
			Method:      http.MethodGet,
			Path:        "/subscriptions/active",
			Summary:     "subscriptions-active",
			Description: "get active user subscriptions",
			Tags:        []string{"Stripe", "Subscriptions"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				selectCustomerFromUser,
			},
		},
		appApi.GetStripeSubscriptions,
	)
	// stripe billing portal
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "stripe-billing-portal",
			Method:      http.MethodPost,
			Path:        "/subscriptions/billing-portals",
			Summary:     "create user billing-portal",
			Description: "billing-portals",
			Tags:        []string{"Subscriptions", "Billing Portal"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Middlewares: huma.Middlewares{
				selectCustomerFromUser,
			},
		},
		appApi.StripeBillingPortal,
	)
	//  stripe checkout session
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "create-checkout-session",
			Method:      http.MethodPost,
			Path:        "/subscriptions/checkout-session",
			Summary:     "create checkout session",
			Description: "user create checkout session",
			Tags:        []string{"Subscriptions", "Checkout Session"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Middlewares: huma.Middlewares{
				selectCustomerFromUser,
			},
		},
		appApi.CreateUserCheckoutSession,
	)
	// stripe my subscriptions
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "team-subscriptions-active",
			Method:      http.MethodGet,
			Path:        "/teams/{team-id}/subscriptions/active",
			Summary:     "team-subscriptions-active",
			Description: "get active team subscriptions",
			Tags:        []string{"Stripe", "Subscriptions", "Team"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				teamInfoFromParam,
				selectCustomerFromTeam,
			},
		},
		appApi.GetTeamStripeSubscriptions,
	)
	//  stripe checkout session team create
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "create-team-checkout-session",
			Method:      http.MethodPost,
			Path:        "/teams/{team-id}/subscriptions/checkout-session",
			Summary:     "create checkout session",
			Description: "user create checkout session",
			Tags:        []string{"Subscriptions", "Checkout Session"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Middlewares: huma.Middlewares{
				teamInfoFromParam,
				selectOrCreateOwnerCustomerFromTeam,
			},
		},
		appApi.CreateTeamCheckoutSession,
	)
	// stripe billing portal
	huma.Register(
		stripeGroup,
		huma.Operation{
			OperationID: "stripe-billing-portal-team",
			Method:      http.MethodPost,
			Path:        "/teams/{team-id}/subscriptions/billing-portals",
			Summary:     "create team billing-portal",
			Description: "billing-portals",
			Tags:        []string{"Subscriptions", "Billing Portal", "Team"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
			Middlewares: huma.Middlewares{
				teamInfoFromParam,
				selectOrCreateOwnerCustomerFromTeam,
			},
		},
		appApi.StripeTeamBillingPortal,
	)

}
