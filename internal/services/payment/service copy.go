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

type PaymentService2 interface {
	Client() PaymentClient
	CreateBillingPortalSession(ctx context.Context, userId uuid.UUID) (string, error)
	CreateCheckoutSession(ctx context.Context, userId uuid.UUID, priceId string) (string, error)
	FindAndUpsertAllPrices(ctx context.Context) error
	FindAndUpsertAllProducts(ctx context.Context) error
	FindOrCreateCustomerFromUser(ctx context.Context, userId uuid.UUID, email string) (*models.StripeCustomer, error)
	FindSubscriptionWithPriceBySessionId(ctx context.Context, sessionId string) (*models.SubscriptionWithPrice, error)
	SyncPerms(ctx context.Context) error
	SyncRoles(ctx context.Context) error
	UpsertPriceProductFromStripe(ctx context.Context) error
	UpsertSubscriptionByIds(ctx context.Context, cutomerId string, subscriptionId string) error
}

type StripeService2 struct {
	logger       *slog.Logger
	client       PaymentClient
	paymentStore PaymentStore2
	db           db.Dbx
}

var _ PaymentService2 = (*StripeService2)(nil)

func (srv *StripeService2) Client() PaymentClient {
	return srv.client
}
func NewStripeServiceFromConf2(conf conf.StripeConfig) *StripeService2 {
	return &StripeService2{client: NewStripeClient(conf), logger: slog.Default(), paymentStore: NewPostgresStripeStore()}
}
func NewStripeService2(client PaymentClient, store PaymentStore2) *StripeService2 {
	return &StripeService2{client: client, logger: slog.Default(), paymentStore: store}
}

func (srv *StripeService2) SyncRoles(ctx context.Context) error {
	var err error
	for productId, role := range shared.StripeRoleMap {
		err = func() error {
			var roleName string = role
			product, err := srv.paymentStore.FindProductByStripeId(ctx, productId)
			if err != nil {
				return err
			}
			if product == nil {
				return errors.New("product not found")
			}
			role, err := srv.paymentStore.FindRoleByName(ctx, roleName)
			if err != nil {
				return err
			}
			if role == nil {
				return errors.New("role not found")
			}
			return srv.paymentStore.CreateProductRoles(ctx, product.ID, role.ID)
		}()
	}
	return err
}

func (srv *StripeService2) SyncPerms(ctx context.Context) error {
	var err error
	for productId, role := range shared.StripeRoleMap {
		err = func() error {
			product, err := srv.paymentStore.FindProductByStripeId(ctx, productId)
			if err != nil {
				return err
			}
			if product == nil {
				return errors.New("product not found")
			}
			perm, err := srv.paymentStore.FindPermissionByName(ctx, role)
			if err != nil {
				return err
			}
			if perm == nil {
				return errors.New("permission not found")
			}
			return srv.paymentStore.CreateProductPermissions(ctx, product.ID, perm.ID)
		}()
	}
	return err
}

func (srv *StripeService2) UpsertPriceProductFromStripe(ctx context.Context) error {
	if err := srv.FindAndUpsertAllProducts(ctx); err != nil {
		fmt.Println(err)
		return err
	}
	if err := srv.FindAndUpsertAllPrices(ctx); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (srv *StripeService2) FindAndUpsertAllProducts(ctx context.Context) error {
	products, err := srv.client.FindAllProducts()
	if err != nil {
		srv.logger.Error("error finding all products", "error", err)
		return err
	}
	for _, product := range products {
		err = srv.paymentStore.UpsertProductFromStripe(ctx, product)
		if err != nil {
			srv.logger.Error("error upserting product", "product", product.ID, "error", err)
			continue
		}
	}
	return nil
}

func (srv *StripeService2) FindAndUpsertAllPrices(ctx context.Context) error {
	prices, err := srv.client.FindAllPrices()
	if err != nil {
		srv.logger.Error("error finding all prices", "error", err)
		return err
	}
	for _, price := range prices {
		err = srv.paymentStore.UpsertPriceFromStripe(ctx, price)
		if err != nil {
			srv.logger.Error("error upserting price", "price", price.ID, "error", err)
			continue
		}
	}
	return nil
}

func (srv *StripeService2) FindSubscriptionWithPriceBySessionId(ctx context.Context, sessionId string) (*models.SubscriptionWithPrice, error) {
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
	data, err := srv.paymentStore.FindSubscriptionWithPriceById(ctx, sub.Subscription.ID)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (srv *StripeService2) UpsertSubscriptionByIds(ctx context.Context, cutomerId, subscriptionId string) error {
	cus, err := srv.paymentStore.FindCustomerByStripeId(ctx, cutomerId)
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
	err = srv.paymentStore.UpsertSubscriptionFromStripe(ctx, sub, cus.ID)
	if err != nil {
		return err
	}
	return nil
}

func (srv *StripeService2) FindOrCreateCustomerFromUser(ctx context.Context, userId uuid.UUID, email string) (*models.StripeCustomer, error) {
	dbCus, err := srv.paymentStore.FindCustomerByUserId(ctx, userId)
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

	err = srv.paymentStore.UpsertCustomerStripeId(ctx, userId, stripeCus.ID)
	if err != nil {
		return nil, err
	}
	return srv.paymentStore.FindCustomerByUserId(ctx, userId)
}

func (srv *StripeService2) CreateCheckoutSession(ctx context.Context, userId uuid.UUID, priceId string) (string, error) {
	team, err := srv.paymentStore.FindTeamById(ctx, userId)
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

	val, err := srv.paymentStore.FindLatestActiveSubscriptionByTeamId(ctx, userId)
	if err != nil {
		return "", err
	}
	if val != nil {
		return "", errors.New("user already has a valid subscription")
	}
	firstSub, err := srv.paymentStore.IsFirstSubscription(ctx, team.ID)
	if err != nil {
		return "", err
	}
	var trialDays *int64
	if firstSub {
		trialDays = types.Pointer(int64(14))
	}
	valPrice, err := srv.paymentStore.FindValidPriceById(ctx, priceId)
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

func (srv *StripeService2) CreateBillingPortalSession(ctx context.Context, userId uuid.UUID) (string, error) {
	team, err := srv.paymentStore.FindTeamById(ctx, userId)
	if err != nil {
		return "", err
	}
	if team == nil {
		return "", errors.New("user not found")
	}
	stripe_customer_id := *team.StripeCustomerID
	// find or create customer from user
	// dbcus, err := s.FindOrCreateCustomerFromUser(ctx,  team.ID, team.Email)
	// if err != nil {
	// 	return "", err
	// }
	// if dbcus == nil {
	// 	return "", errors.New("customer not found")
	// }
	// verify user has a valid subscriptio
	sub, err := srv.paymentStore.FindLatestActiveSubscriptionByTeamId(ctx, team.ID)
	if err != nil {
		return "", err
	}
	if sub == nil {
		return "", errors.New("no subscription.  subscribe to access billing portal")
	}
	prods, err := srv.paymentStore.ListProducts(ctx, &shared.StripeProductListParams{
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
	prices, err := srv.paymentStore.ListPrices(ctx, &shared.StripePriceListParams{
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

	config, err := srv.client.CreatePortalConfiguration(configurations...)
	if err != nil {
		return "", err
	}
	url, err := srv.client.CreateBillingPortalSession(stripe_customer_id, config)
	if err != nil {
		log.Println(err)
		return "", errors.New("failed to create checkout session")
	}
	if url == nil {
		return "", errors.New("failed to create checkout session")
	}
	return url.URL, nil
}
