package payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
)

type PaymentStore interface {
	FindSubscriptionWithPriceById(ctx context.Context, dbx db.Dbx, stripeId string) (*models.SubscriptionWithPrice, error)
	FindProductByStripeId(ctx context.Context, dbx db.Dbx, productId string) (*models.StripeProduct, error)
	FindCustomerByStripeId(ctx context.Context, dbx db.Dbx, stripeId string) (*models.StripeCustomer, error)
	FindCustomerByUserId(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (*models.StripeCustomer, error)
	UpsertSubscriptionFromStripe(ctx context.Context, dbx db.Dbx, sub *stripe.Subscription, userId uuid.UUID) error
	UpsertCustomerStripeId(ctx context.Context, dbx db.Dbx, userId uuid.UUID, stripeCustomerId string) error
	UpsertProductFromStripe(ctx context.Context, dbx db.Dbx, product *stripe.Product) error
	UpsertPriceFromStripe(ctx context.Context, dbx db.Dbx, price *stripe.Price) error
	CreateProductRoles(ctx context.Context, db db.Dbx, productId string, roleIds ...uuid.UUID) error
	CreateProductPermissions(ctx context.Context, db db.Dbx, productId string, permissionIds ...uuid.UUID) error
	FindUserById(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (*models.User, error)
	FindLatestActiveSubscriptionByUserId(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (*models.StripeSubscription, error)
	IsFirstSubscription(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (bool, error)
	FindValidPriceById(ctx context.Context, dbx db.Dbx, priceId string) (*models.StripePrice, error)
	ListProducts(ctx context.Context, dbx db.Dbx, input *shared.StripeProductListParams) ([]*models.StripeProduct, error)
	ListPrices(ctx context.Context, dbx db.Dbx, input *shared.StripePriceListParams) ([]*models.StripePrice, error)
}

type StripeStore struct {
}

// CreateProductPermissions implements PaymentStore.
func (s *StripeStore) CreateProductPermissions(ctx context.Context, db db.Dbx, productId string, permissionIds ...uuid.UUID) error {
	return queries.CreateProductPermissions(ctx, db, productId, permissionIds...)
}

var _ PaymentStore = (*StripeStore)(nil)

func NewStripeStore() *StripeStore {
	return &StripeStore{}
}

// CreateProductRoles implements PaymentStore.
func (s *StripeStore) CreateProductRoles(ctx context.Context, db db.Dbx, productId string, roleIds ...uuid.UUID) error {
	return queries.CreateProductRoles(ctx, db, productId, roleIds...)
}

// UpsertPriceFromStripe implements PaymentStore.
func (s *StripeStore) UpsertPriceFromStripe(ctx context.Context, dbx db.Dbx, price *stripe.Price) error {
	return queries.UpsertPriceFromStripe(ctx, dbx, price)
}

// UpsertProductFromStripe implements PaymentStore.
func (s *StripeStore) UpsertProductFromStripe(ctx context.Context, dbx db.Dbx, product *stripe.Product) error {
	return queries.UpsertProductFromStripe(ctx, dbx, product)
}

// FindCustomerByStripeId implements PaymentStore.
func (s *StripeStore) FindCustomerByStripeId(ctx context.Context, dbx db.Dbx, stripeId string) (*models.StripeCustomer, error) {
	return queries.FindCustomerByStripeId(ctx, dbx, stripeId)
}

// FindCustomerByUserId implements PaymentStore.
func (s *StripeStore) FindCustomerByUserId(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (*models.StripeCustomer, error) {
	return queries.FindCustomerByUserId(ctx, dbx, userId)
}

// FindLatestActiveSubscriptionByUserId implements PaymentStore.
func (s *StripeStore) FindLatestActiveSubscriptionByUserId(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (*models.StripeSubscription, error) {
	return queries.FindLatestActiveSubscriptionByUserId(ctx, dbx, userId)
}

// FindProductByStripeId implements PaymentStore.
func (s *StripeStore) FindProductByStripeId(ctx context.Context, dbx db.Dbx, productId string) (*models.StripeProduct, error) {
	return queries.FindProductByStripeId(ctx, dbx, productId)
}

// FindSubscriptionWithPriceById implements PaymentStore.
func (s *StripeStore) FindSubscriptionWithPriceById(ctx context.Context, dbx db.Dbx, stripeId string) (*models.SubscriptionWithPrice, error) {
	return queries.FindSubscriptionWithPriceById(ctx, dbx, stripeId)
}

// FindUserById implements PaymentStore.
func (s *StripeStore) FindUserById(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (*models.User, error) {
	return queries.FindUserById(ctx, dbx, userId)
}

// FindValidPriceById implements PaymentStore.
func (s *StripeStore) FindValidPriceById(ctx context.Context, dbx db.Dbx, priceId string) (*models.StripePrice, error) {
	return queries.FindValidPriceById(ctx, dbx, priceId)
}

// IsFirstSubscription implements PaymentStore.
func (s *StripeStore) IsFirstSubscription(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (bool, error) {
	return queries.IsFirstSubscription(ctx, dbx, userId)
}

// ListPrices implements PaymentStore.
func (s *StripeStore) ListPrices(ctx context.Context, dbx db.Dbx, input *shared.StripePriceListParams) ([]*models.StripePrice, error) {
	return queries.ListPrices(ctx, dbx, input)
}

// ListProducts implements PaymentStore.
func (s *StripeStore) ListProducts(ctx context.Context, dbx db.Dbx, input *shared.StripeProductListParams) ([]*models.StripeProduct, error) {
	return queries.ListProducts(ctx, dbx, input)
}

// UpsertCustomerStripeId implements PaymentStore.
func (s *StripeStore) UpsertCustomerStripeId(ctx context.Context, dbx db.Dbx, userId uuid.UUID, stripeCustomerId string) error {
	return queries.UpsertCustomerStripeId(ctx, dbx, userId, stripeCustomerId)
}

// UpsertSubscriptionFromStripe implements PaymentStore.
func (s *StripeStore) UpsertSubscriptionFromStripe(ctx context.Context, dbx db.Dbx, sub *stripe.Subscription, userId uuid.UUID) error {
	return queries.UpsertSubscriptionFromStripe(ctx, dbx, sub, userId)
}
