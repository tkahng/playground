package payment

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
)

type PaymentService interface {
	Client() PaymentClient
	CreateBillingPortalSession(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (string, error)
	CreateCheckoutSession(ctx context.Context, dbx db.Dbx, userId uuid.UUID, priceId string) (string, error)
	FindAndUpsertAllPrices(ctx context.Context, dbx db.Dbx) error
	FindAndUpsertAllProducts(ctx context.Context, dbx db.Dbx) error
	FindOrCreateCustomerFromUser(ctx context.Context, dbx db.Dbx, userId uuid.UUID, email string) (*models.StripeCustomer, error)
	FindSubscriptionWithPriceBySessionId(ctx context.Context, dbx db.Dbx, sessionId string) (*models.SubscriptionWithPrice, error)
	// Logger() *slog.Logger
	SyncPerms(ctx context.Context, dbx db.Dbx) error
	// SyncProductPerms(ctx context.Context, dbx db.Dbx, productId string, permName string) error
	// SyncProductRole(ctx context.Context, dbx db.Dbx, productId string, roleName string) error
	SyncRoles(ctx context.Context, dbx db.Dbx) error
	UpsertPriceProductFromStripe(ctx context.Context, dbx db.Dbx) error
	UpsertSubscriptionByIds(ctx context.Context, dbx db.Dbx, cutomerId string, subscriptionId string) error
}

type StripeService struct {
	logger *slog.Logger
	client PaymentClient
	store  PaymentStore
}

var _ PaymentService = (*StripeService)(nil)

func (srv *StripeService) Client() PaymentClient {
	return srv.client
}
func NewStripeServiceFromConf(conf conf.StripeConfig) *StripeService {
	return &StripeService{client: NewStripeClient(conf), logger: slog.Default(), store: NewStripeStore()}
}
func NewStripeService(client PaymentClient, store PaymentStore) *StripeService {
	return &StripeService{client: client, logger: slog.Default(), store: store}
}

func (srv *StripeService) SyncRoles(ctx context.Context, dbx db.Dbx) error {
	var err error
	for productId, role := range shared.StripeRoleMap {
		err = func() error {
			var roleName string = role
			product, err := srv.store.FindProductByStripeId(ctx, dbx, productId)
			if err != nil {
				return err
			}
			if product == nil {
				return errors.New("product not found")
			}
			role, err := srv.store.FindRoleByName(ctx, dbx, roleName)
			if err != nil {
				return err
			}
			if role == nil {
				return errors.New("role not found")
			}
			return srv.store.CreateProductRoles(ctx, dbx, product.ID, role.ID)
		}()
	}
	return err
}

func (srv *StripeService) SyncPerms(ctx context.Context, dbx db.Dbx) error {
	var err error
	for productId, role := range shared.StripeRoleMap {
		err = func() error {
			product, err := srv.store.FindProductByStripeId(ctx, dbx, productId)
			if err != nil {
				return err
			}
			if product == nil {
				return errors.New("product not found")
			}
			perm, err := srv.store.FindPermissionByName(ctx, dbx, role)
			if err != nil {
				return err
			}
			if perm == nil {
				return errors.New("permission not found")
			}
			return srv.store.CreateProductPermissions(ctx, dbx, product.ID, perm.ID)
		}()
	}
	return err
}

// func (srv *StripeService) SyncProductRole(ctx context.Context, dbx db.Dbx, productId string, roleName string) error {
// 	product, err := srv.store.FindProductByStripeId(ctx, dbx, productId)
// 	if err != nil {
// 		return err
// 	}
// 	if product == nil {
// 		return errors.New("product not found")
// 	}
// 	role, err := srv.store.FindRoleByName(ctx, dbx, roleName)
// 	if err != nil {
// 		return err
// 	}
// 	if role == nil {
// 		return errors.New("role not found")
// 	}
// 	return srv.store.CreateProductRoles(ctx, dbx, product.ID, role.ID)
// }

// func (srv *StripeService) SyncProductPerms(ctx context.Context, dbx db.Dbx, productId string, permName string) error {
// 	product, err := srv.store.FindProductByStripeId(ctx, dbx, productId)
// 	if err != nil {
// 		return err
// 	}
// 	if product == nil {
// 		return errors.New("product not found")
// 	}
// 	perm, err := srv.store.FindPermissionByName(ctx, dbx, permName)
// 	if err != nil {
// 		return err
// 	}
// 	if perm == nil {
// 		return errors.New("permission not found")
// 	}
// 	return srv.store.CreateProductPermissions(ctx, dbx, product.ID, perm.ID)
// }

func (srv *StripeService) UpsertPriceProductFromStripe(ctx context.Context, dbx db.Dbx) error {
	if err := srv.FindAndUpsertAllProducts(ctx, dbx); err != nil {
		fmt.Println(err)
		return err
	}
	if err := srv.FindAndUpsertAllPrices(ctx, dbx); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (srv *StripeService) FindAndUpsertAllProducts(ctx context.Context, dbx db.Dbx) error {
	products, err := srv.client.FindAllProducts()
	if err != nil {
		srv.logger.Error("error finding all products", "error", err)
		return err
	}
	for _, product := range products {
		err = srv.store.UpsertProductFromStripe(ctx, dbx, product)
		if err != nil {
			srv.logger.Error("error upserting product", "product", product.ID, "error", err)
			continue
		}
	}
	return nil
}

func (srv *StripeService) FindAndUpsertAllPrices(ctx context.Context, dbx db.Dbx) error {
	prices, err := srv.client.FindAllPrices()
	if err != nil {
		srv.logger.Error("error finding all prices", "error", err)
		return err
	}
	for _, price := range prices {
		err = srv.store.UpsertPriceFromStripe(ctx, dbx, price)
		if err != nil {
			srv.logger.Error("error upserting price", "price", price.ID, "error", err)
			continue
		}
	}
	return nil
}

func (srv *StripeService) FindSubscriptionWithPriceBySessionId(ctx context.Context, dbx db.Dbx, sessionId string) (*models.SubscriptionWithPrice, error) {
	sub, err := srv.client.FindCheckoutSessionByStripeId(sessionId)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, errors.New("subscription not found")
	}
	if sub.Subscription == nil {
		return nil, errors.New("subscription not found")
	}
	data, err := srv.store.FindSubscriptionWithPriceById(ctx, dbx, sub.Subscription.ID)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (srv *StripeService) UpsertSubscriptionByIds(ctx context.Context, dbx db.Dbx, cutomerId, subscriptionId string) error {
	cus, err := srv.store.FindCustomerByStripeId(ctx, dbx, cutomerId)
	if err != nil {
		return err
	}
	if cus == nil {
		return errors.New("customer not found")
	}
	sub, err := srv.client.FindSubscriptionByStripeId(subscriptionId)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.New("subscription not found")
	}
	err = srv.store.UpsertSubscriptionFromStripe(ctx, dbx, sub, cus.ID)
	if err != nil {
		return err
	}
	return nil
}

func (srv *StripeService) FindOrCreateCustomerFromUser(ctx context.Context, dbx db.Dbx, userId uuid.UUID, email string) (*models.StripeCustomer, error) {
	dbCus, err := srv.store.FindCustomerByUserId(ctx, dbx, userId)
	if err != nil {
		return nil, err
	}
	if dbCus != nil {
		return dbCus, nil
	}
	stripeCus, err := srv.client.FindOrCreateCustomer(email, userId)
	if err != nil {
		return nil, err
	}
	if stripeCus == nil {
		return nil, errors.New("failed to find or create customer in stripe")
	}

	err = srv.store.UpsertCustomerStripeId(ctx, dbx, userId, stripeCus.ID)
	if err != nil {
		return nil, err
	}
	return srv.store.FindCustomerByUserId(ctx, dbx, userId)
}

func (srv *StripeService) CreateCheckoutSession(ctx context.Context, dbx db.Dbx, userId uuid.UUID, priceId string) (string, error) {
	team, err := srv.store.FindTeamById(ctx, dbx, userId)
	if err != nil {
		return "", err
	}
	if team == nil {
		return "", errors.New("user not found")
	}
	if team.StripeCustomerID == nil {
		return "", errors.New("team does not have a stripe customer id")
	}
	customer_stripe_id := *team.StripeCustomerID

	val, err := srv.store.FindLatestActiveSubscriptionByTeamId(ctx, dbx, userId)
	if err != nil {
		return "", err
	}
	if val != nil {
		return "", errors.New("user already has a valid subscription")
	}
	firstSub, err := srv.store.IsFirstSubscription(ctx, dbx, team.ID)
	if err != nil {
		return "", err
	}
	var trialDays *int64
	if firstSub {
		trialDays = types.Pointer(int64(14))
	}
	valPrice, err := srv.store.FindValidPriceById(ctx, dbx, priceId)
	if err != nil {
		return "", err
	}
	if valPrice == nil {
		return "", errors.New("price is not valid")
	}
	sesh, err := srv.client.CreateCheckoutSession(customer_stripe_id, priceId, trialDays)
	if err != nil {
		return "", err
	}
	return sesh.URL, nil
}

func (s *StripeService) CreateBillingPortalSession(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (string, error) {
	team, err := s.store.FindTeamById(ctx, dbx, userId)
	if err != nil {
		return "", err
	}
	if team == nil {
		return "", errors.New("user not found")
	}
	stripe_customer_id := *team.StripeCustomerID
	// find or create customer from user
	// dbcus, err := s.FindOrCreateCustomerFromUser(ctx, dbx, team.ID, team.Email)
	// if err != nil {
	// 	return "", err
	// }
	// if dbcus == nil {
	// 	return "", errors.New("customer not found")
	// }
	// verify user has a valid subscriptio
	sub, err := s.store.FindLatestActiveSubscriptionByTeamId(ctx, dbx, team.ID)
	if err != nil {
		return "", err
	}
	if sub == nil {
		return "", errors.New("no subscription.  subscribe to access billing portal")
	}
	prods, err := s.store.ListProducts(ctx, dbx, &shared.StripeProductListParams{
		PaginatedInput: shared.PaginatedInput{
			PerPage: 100,
		},
		StripeProductListFilter: shared.StripeProductListFilter{
			Active: shared.Active,
		},
	})
	if err != nil {
		return "", err
	}
	prodIds := make([]string, len(prods))
	for i, p := range prods {
		prodIds[i] = p.ID
	}
	prices, err := s.store.ListPrices(ctx, dbx, &shared.StripePriceListParams{
		PaginatedInput: shared.PaginatedInput{
			PerPage: 100,
		},
		StripePriceListFilter: shared.StripePriceListFilter{
			Active:     shared.Active,
			ProductIds: prodIds,
		},
	})
	grouped := mapper.MapToMany(prices, prodIds, func(p *models.StripePrice) string { return p.ProductID })
	if err != nil {
		return "", err
	}
	var configurations []*ProductBillingConfigurationInput
	for i, id := range prods {
		price := grouped[i]
		con := &ProductBillingConfigurationInput{
			Product: &id.ID,
			Prices: mapper.Map(price, func(p *models.StripePrice) *string {
				return &p.ID
			}),
		}
		configurations = append(configurations, con)
	}

	config, err := s.client.CreatePortalConfiguration(configurations...)
	if err != nil {
		return "", err
	}
	url, err := s.client.CreateBillingPortalSession(stripe_customer_id, config)
	if err != nil {
		log.Println(err)
		return "", errors.New("failed to create checkout session")
	}
	if url == nil {
		return "", errors.New("failed to create checkout session")
	}
	return url.URL, nil
}
