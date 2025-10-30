package apis

import (
	"github.com/danielgtaylor/huma/v2"
)

func bindStripeApi(api huma.API, a *Api) {
	stripeGroup := huma.NewGroup(api)

	// stripe webhook
	a.bindStripeWebhook(stripeGroup)
	// stripe products with prices
	a.bindStripeProductsWithPrices(stripeGroup)

	//  stripe get checkout session by checkoutSessionId
	a.bindStripeCheckoutSessionGet(stripeGroup)
	// stripe my subscriptions
	a.bindGetStripeSubscriptions(stripeGroup)
	// stripe billing portal
	a.bindStripeBillingPortal(stripeGroup)
	//  stripe checkout session
	a.bindCreateUserCheckoutSession(stripeGroup)
	// stripe team subscriptions
	a.bindGetTeamStripeSubscriptions(stripeGroup)
	//  stripe checkout session team create
	a.bindCreateTeamCheckoutSession(stripeGroup)
	// stripe billing portal
	a.bindStripeTeamBillingPortal(stripeGroup)
}
