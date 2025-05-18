package paymentmodule

import (
	"context"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type PaymentStore interface {
	// FindPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	// CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
	FindSubscriptionWithPriceById(ctx context.Context, stripeId string) (*models.SubscriptionWithPrice, error)
	FindProductByStripeId(ctx context.Context, productId string) (*models.StripeProduct, error)
	FindCustomerByStripeId(ctx context.Context, stripeId string) (*models.StripeCustomer, error)
	FindCustomerByUserId(ctx context.Context, userId uuid.UUID) (*models.StripeCustomer, error)
	UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription, userId uuid.UUID) error
	UpsertSubscription(ctx context.Context, sub *models.StripeSubscription) error
	UpsertCustomerStripeId(ctx context.Context, userId uuid.UUID, stripeCustomerId string) error
	UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error
	UpsertProduct(ctx context.Context, product *models.StripeProduct) error
	UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error
	UpsertPrice(ctx context.Context, price *models.StripePrice) error
	FindTeamById(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	FindLatestActiveSubscriptionByTeamId(ctx context.Context, teamId uuid.UUID) (*models.StripeSubscription, error)
	IsFirstSubscription(ctx context.Context, teamId uuid.UUID) (bool, error)
	FindValidPriceById(ctx context.Context, priceId string) (*models.StripePrice, error)
	ListProducts(ctx context.Context, input *shared.StripeProductListParams) ([]*models.StripeProduct, error)
	ListPrices(ctx context.Context, input *shared.StripePriceListParams) ([]*models.StripePrice, error)
}

type RBACStore interface {
	FindPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
}
