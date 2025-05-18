package payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
)

type PosrgresStripeStore struct {
	db db.Dbx
}

// FindRoleByName implements PaymentStore.
func (s *PosrgresStripeStore) FindRoleByName(ctx context.Context, name string) (*models.Role, error) {
	return queries.FindRoleByName(ctx, s.db, name)
}

// FindPermissionByName implements PaymentStore.
func (s *PosrgresStripeStore) FindPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	return queries.FindPermissionByName(ctx, s.db, name)
}

// CreateProductPermissions implements PaymentStore.
func (s *PosrgresStripeStore) CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error {
	return queries.CreateProductPermissions(ctx, s.db, productId, permissionIds...)
}

var _ PaymentStore2 = (*PosrgresStripeStore)(nil)

func NewPostgresStripeStore() *PosrgresStripeStore {
	return &PosrgresStripeStore{}
}

// CreateProductRoles implements PaymentStore.
func (s *PosrgresStripeStore) CreateProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error {
	return queries.CreateProductRoles(ctx, s.db, productId, roleIds...)
}

// UpsertPriceFromStripe implements PaymentStore.
func (s *PosrgresStripeStore) UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error {
	return queries.UpsertPriceFromStripe(ctx, s.db, price)
}

// UpsertProductFromStripe implements PaymentStore.
func (s *PosrgresStripeStore) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
	return queries.UpsertProductFromStripe(ctx, s.db, product)
}

// FindCustomerByStripeId implements PaymentStore.
func (s *PosrgresStripeStore) FindCustomerByStripeId(ctx context.Context, stripeId string) (*models.StripeCustomer, error) {
	return queries.FindCustomerByStripeId(ctx, s.db, stripeId)
}

// FindCustomerByUserId implements PaymentStore.
func (s *PosrgresStripeStore) FindCustomerByUserId(ctx context.Context, userId uuid.UUID) (*models.StripeCustomer, error) {
	return queries.FindCustomerByUserId(ctx, s.db, userId)
}

// FindLatestActiveSubscriptionByTeamId implements PaymentStore.
func (s *PosrgresStripeStore) FindLatestActiveSubscriptionByTeamId(ctx context.Context, teamId uuid.UUID) (*models.StripeSubscription, error) {
	return queries.FindLatestActiveSubscriptionByTeamId(ctx, s.db, teamId)
}

// FindProductByStripeId implements PaymentStore.
func (s *PosrgresStripeStore) FindProductByStripeId(ctx context.Context, productId string) (*models.StripeProduct, error) {
	return queries.FindProductByStripeId(ctx, s.db, productId)
}

// FindSubscriptionWithPriceById implements PaymentStore.
func (s *PosrgresStripeStore) FindSubscriptionWithPriceById(ctx context.Context, stripeId string) (*models.SubscriptionWithPrice, error) {
	return queries.FindSubscriptionWithPriceById(ctx, s.db, stripeId)
}

// FindTeamById implements PaymentStore.
func (s *PosrgresStripeStore) FindTeamById(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
	return crudrepo.Team.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": teamId.String(),
			},
		},
	)
}

// FindValidPriceById implements PaymentStore.
func (s *PosrgresStripeStore) FindValidPriceById(ctx context.Context, priceId string) (*models.StripePrice, error) {
	return queries.FindValidPriceById(ctx, s.db, priceId)
}

// IsFirstSubscription implements PaymentStore.
func (s *PosrgresStripeStore) IsFirstSubscription(ctx context.Context, userId uuid.UUID) (bool, error) {
	return queries.IsFirstSubscription(ctx, s.db, userId)
}

// ListPrices implements PaymentStore.
func (s *PosrgresStripeStore) ListPrices(ctx context.Context, input *shared.StripePriceListParams) ([]*models.StripePrice, error) {
	return queries.ListPrices(ctx, s.db, input)
}

// ListProducts implements PaymentStore.
func (s *PosrgresStripeStore) ListProducts(ctx context.Context, input *shared.StripeProductListParams) ([]*models.StripeProduct, error) {
	return queries.ListProducts(ctx, s.db, input)
}

// UpsertCustomerStripeId implements PaymentStore.
func (s *PosrgresStripeStore) UpsertCustomerStripeId(ctx context.Context, userId uuid.UUID, stripeCustomerId string) error {
	return queries.UpsertCustomerStripeId(ctx, s.db, userId, stripeCustomerId)
}

// UpsertSubscriptionFromStripe implements PaymentStore.
func (s *PosrgresStripeStore) UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription, userId uuid.UUID) error {
	return queries.UpsertSubscriptionFromStripe(ctx, s.db, sub, userId)
}
