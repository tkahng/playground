package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"

	// "github.com/tkahng/authgo/internal/modules/paymentmodule/client"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
)

// var NewPaymentClient = client.NewPaymentClient

type PaymentClient interface {
	Config() *conf.StripeConfig
	CreateBillingPortalSession(customerId string, configurationId string) (*stripe.BillingPortalSession, error)
	CreateCheckoutSession(customerId string, priceId string, quantity int64, trialDays *int64) (*stripe.CheckoutSession, error)
	CreateCustomer(email string, name *string) (*stripe.Customer, error)
	CreatePortalConfiguration(input ...*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams) (string, error)
	FindAllPrices() ([]*stripe.Price, error)
	FindAllProducts() ([]*stripe.Product, error)
	FindCheckoutSessionByStripeId(stripeId string) (*stripe.CheckoutSession, error)
	// FindCustomerByEmailAndUserId(email string, userId string) (*stripe.Customer, error)
	FindOrCreateCustomer(email string, name *string) (*stripe.Customer, error)
	FindSubscriptionByStripeId(stripeId string) (*stripe.Subscription, error)
	UpdateCustomer(customerId string, params *stripe.CustomerParams) (*stripe.Customer, error)
	UpdateItemQuantity(itemId string, priceId string, count int64) (*stripe.SubscriptionItem, error)
}

type PaymentTeamStore interface {
	// team methods
	FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error)
	CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error)
}

type PaymentRbacStore interface {
	// permission methods
	FindPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
}

type PaymentStripeStore interface {
	CountProducts(ctx context.Context, filter *shared.StripeProductListFilter) (int64, error)
	LoadProductPrices(ctx context.Context, where *map[string]any, productIds ...string) ([][]*models.StripePrice, error)
	FindSubscriptionWithPriceById(ctx context.Context, stripeId string) (*models.SubscriptionWithPrice, error)
	FindLatestActiveSubscriptionWithPriceByCustomerId(ctx context.Context, customerId string) (*models.SubscriptionWithPrice, error)
	FindProductByStripeId(ctx context.Context, productId string) (*models.StripeProduct, error)
	// customer methods
	FindCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error)
	CreateCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error)

	UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription) error
	UpsertSubscription(ctx context.Context, sub *models.StripeSubscription) error
	UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error
	UpsertProduct(ctx context.Context, product *models.StripeProduct) error
	UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error
	UpsertPrice(ctx context.Context, price *models.StripePrice) error
	IsFirstSubscription(ctx context.Context, customerId string) (bool, error)
	FindValidPriceById(ctx context.Context, priceId string) (*models.StripePrice, error)
	ListProducts(ctx context.Context, input *shared.StripeProductListParams) ([]*models.StripeProduct, error)
	ListPrices(ctx context.Context, input *shared.StripePriceListParams) ([]*models.StripePrice, error)
}

type PaymentStore interface {
	// team methods
	PaymentTeamStore
	// rbac methods
	PaymentRbacStore
	// payment methods
	PaymentStripeStore
}
type PaymentService interface {
	Client() PaymentClient
	Store() PaymentStore

	// admin methods
	SyncPerms(ctx context.Context) error
	UpsertPriceProductFromStripe(ctx context.Context) error
	UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error
	UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error
	FindAndUpsertAllPrices(ctx context.Context) error
	FindAndUpsertAllProducts(ctx context.Context) error
	// customer methods
	CreateUserCustomer(ctx context.Context, user *models.User) (*models.StripeCustomer, error)
	CreateTeamCustomer(ctx context.Context, team *models.Team, user *models.User) (*models.StripeCustomer, error)

	FindCustomerByUser(ctx context.Context, userId uuid.UUID) (*models.StripeCustomer, error)
	FindCustomerByTeam(ctx context.Context, teamId uuid.UUID) (*models.StripeCustomer, error)

	CreateBillingPortalSession(ctx context.Context, stripeCustomerId string) (string, error)
	CreateCheckoutSession(ctx context.Context, stripeCustomerId string, priceId string) (string, error)

	FindSubscriptionWithPriceBySessionId(ctx context.Context, sessionId string) (*models.SubscriptionWithPrice, error)

	UpsertSubscriptionByIds(ctx context.Context, cutomerId string, subscriptionId string) error

	VerifyAndUpdateTeamSubscriptionQuantity(ctx context.Context, teamId uuid.UUID) error
}

type StripeService struct {
	logger       *slog.Logger
	client       PaymentClient
	paymentStore PaymentStore
}

// CreateTeamCustomer implements PaymentService.
func (srv *StripeService) CreateTeamCustomer(ctx context.Context, team *models.Team, user *models.User) (*models.StripeCustomer, error) {
	customer, err := srv.client.CreateCustomer(user.Email, &team.Name)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, errors.New("no customer found")
	}
	stripeCustomer := &models.StripeCustomer{
		ID:           customer.ID,
		Email:        customer.Email,
		Name:         &team.Name,
		TeamID:       types.Pointer(team.ID),
		CustomerType: models.StripeCustomerTypeTeam,
	}
	return srv.paymentStore.CreateCustomer(ctx, stripeCustomer)
}

// CreateUserCustomer implements PaymentService.
func (srv *StripeService) CreateUserCustomer(ctx context.Context, user *models.User) (*models.StripeCustomer, error) {
	customer, err := srv.client.CreateCustomer(user.Email, user.Name)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, errors.New("no customer found")
	}
	stripeCustomer := &models.StripeCustomer{
		ID:           customer.ID,
		Email:        customer.Email,
		Name:         user.Name,
		UserID:       types.Pointer(user.ID),
		CustomerType: models.StripeCustomerTypeUser,
	}
	return srv.paymentStore.CreateCustomer(ctx, stripeCustomer)
}

// FindCustomerByTeam implements PaymentService.
func (srv *StripeService) FindCustomerByTeam(ctx context.Context, teamId uuid.UUID) (*models.StripeCustomer, error) {
	return srv.paymentStore.FindCustomer(
		ctx,
		&models.StripeCustomer{
			TeamID: types.Pointer(teamId),
		},
	)
}

// FindCustomerByUser implements PaymentService.
func (srv *StripeService) FindCustomerByUser(ctx context.Context, userId uuid.UUID) (*models.StripeCustomer, error) {
	return srv.paymentStore.FindCustomer(
		ctx,
		&models.StripeCustomer{
			UserID: types.Pointer(userId),
		},
	)
}

// VerifyAndUpdateTeamSubscriptionQuantity implements PaymentService.
func (srv *StripeService) VerifyAndUpdateTeamSubscriptionQuantity(ctx context.Context, teamId uuid.UUID) error {
	customer, err := srv.paymentStore.FindCustomer(ctx, &models.StripeCustomer{
		TeamID: types.Pointer(teamId),
	})
	if err != nil {
		return err
	}
	if customer == nil {
		return errors.New("no stripe customer id")
	}
	subs, err := srv.paymentStore.FindLatestActiveSubscriptionWithPriceByCustomerId(ctx, customer.ID)
	if err != nil {
		return err
	}
	if subs == nil {
		return errors.New("no subscription")
	}
	sub := subs.Subscription
	count, err := srv.paymentStore.CountTeamMembers(ctx, teamId)
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	if sub.Quantity != count {
		_, err := srv.client.UpdateItemQuantity(sub.ItemID, sub.PriceID, count)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

var _ PaymentService = (*StripeService)(nil)

func (srv *StripeService) Client() PaymentClient {
	return srv.client
}

func NewPaymentService(
	client PaymentClient,
	paymentStore PaymentStore,
) PaymentService {
	return &StripeService{
		client:       client,
		logger:       slog.Default(),
		paymentStore: paymentStore,
	}
}

func (srv *StripeService) Store() PaymentStore {
	return srv.paymentStore
}

func (srv *StripeService) SyncPerms(ctx context.Context) error {
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

func (srv *StripeService) UpsertPriceProductFromStripe(ctx context.Context) error {
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

func (srv *StripeService) FindAndUpsertAllProducts(ctx context.Context) error {
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

func (srv *StripeService) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
	err := srv.paymentStore.UpsertProductFromStripe(ctx, product)
	if err != nil {
		srv.logger.Error("error upserting product", "product", product.ID, "error", err)
		return err
	}
	return nil
}

func (srv *StripeService) UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error {
	err := srv.paymentStore.UpsertPriceFromStripe(ctx, price)
	if err != nil {
		srv.logger.Error("error upserting price", "price", price.ID, "error", err)
		return err
	}
	return nil
}

func (srv *StripeService) FindAndUpsertAllPrices(ctx context.Context) error {
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

func (srv *StripeService) FindSubscriptionWithPriceBySessionId(ctx context.Context, sessionId string) (*models.SubscriptionWithPrice, error) {
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

func (srv *StripeService) UpsertSubscriptionByIds(ctx context.Context, cutomerId, subscriptionId string) error {
	customer, err := srv.paymentStore.FindCustomer(ctx, &models.StripeCustomer{
		ID: cutomerId,
	})
	if err != nil {
		return err
	}
	if customer == nil {
		return errors.New("customer not found")
	}
	sub, err := srv.client.FindSubscriptionByStripeId(subscriptionId)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.New("subscription not found")
	}
	err = srv.paymentStore.UpsertSubscriptionFromStripe(ctx, sub)
	if err != nil {
		return err
	}
	return nil
}

func (srv *StripeService) CreateCheckoutSession(ctx context.Context, stripeCustomerId string, priceId string) (string, error) {
	customer, err := srv.paymentStore.FindCustomer(ctx, &models.StripeCustomer{
		ID: stripeCustomerId,
	})
	if err != nil {
		return "", err
	}
	if customer == nil {
		return "", errors.New("customer not found")
	}
	var count int64
	if customer.TeamID != nil {
		team, err := srv.paymentStore.FindTeamByStripeCustomerId(ctx, stripeCustomerId)
		if err != nil {
			return "", err
		}
		if team == nil {
			return "", errors.New("team not found")
		}
		count, err = srv.paymentStore.CountTeamMembers(ctx, team.ID)
		if err != nil {
			return "", err
		}
	} else {
		count = 1
	}
	val, err := srv.paymentStore.FindLatestActiveSubscriptionWithPriceByCustomerId(ctx, stripeCustomerId)
	if err != nil {
		return "", err
	}
	if val != nil {
		return "", errors.New("user already has a valid subscription")
	}
	firstSub, err := srv.paymentStore.IsFirstSubscription(ctx, stripeCustomerId)
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
	sesh, err := srv.client.CreateCheckoutSession(stripeCustomerId, priceId, count, trialDays)
	if err != nil {
		return "", err
	}
	return sesh.URL, nil
}

func (srv *StripeService) CreateBillingPortalSession(ctx context.Context, stripeCustomerId string) (string, error) {
	team, err := srv.paymentStore.FindTeamByStripeCustomerId(ctx, stripeCustomerId)
	if err != nil {
		return "", err
	}
	if team == nil {
		return "", errors.New("team not found")
	}
	stripe_customer_id := stripeCustomerId

	sub, err := srv.paymentStore.FindLatestActiveSubscriptionWithPriceByCustomerId(ctx, stripe_customer_id)
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
	var configurations []*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams
	for i, id := range prods {
		price := grouped[i]
		con := &stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams{
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
