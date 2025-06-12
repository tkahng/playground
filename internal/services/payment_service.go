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
	"github.com/tkahng/authgo/internal/stores"

	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type PaymentClient interface {
	Config() *conf.StripeConfig
	CreateBillingPortalSession(customerId string, configurationId string) (*stripe.BillingPortalSession, error)
	CreateCheckoutSession(customerId string, priceId string, quantity int64, trialDays *int64) (*stripe.CheckoutSession, error)
	CreateCustomer(email string, name *string) (*stripe.Customer, error)
	CreatePortalConfiguration(input ...*stripe.BillingPortalConfigurationFeaturesSubscriptionUpdateProductParams) (string, error)
	FindAllPrices() ([]*stripe.Price, error)
	FindAllProducts() ([]*stripe.Product, error)
	FindCheckoutSessionByStripeId(stripeId string) (*stripe.CheckoutSession, error)
	FindOrCreateCustomer(email string, name *string) (*stripe.Customer, error)
	FindSubscriptionByStripeId(stripeId string) (*stripe.Subscription, error)
	UpdateCustomer(customerId string, params *stripe.CustomerParams) (*stripe.Customer, error)
	UpdateItemQuantity(itemId string, priceId string, count int64) (*stripe.SubscriptionItem, error)
}

type PaymentTeamStore interface {
	// team methods
	FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error)
	CountTeamMembers(ctx context.Context, filter *stores.TeamMemberFilter) (int64, error)
}

type PaymentRbacStore interface {
	LoadProductPermissions(ctx context.Context, productIds ...string) ([][]*models.Permission, error)
	FindPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	CreateProductPermissions(ctx context.Context, productId string, permissionIds ...uuid.UUID) error
	CreateProductRoles(ctx context.Context, productId string, roleIds ...uuid.UUID) error
}

type PaymentStripeStore interface {
	// customers crud
	CustomerStore
	SubscriptionStore
	// prices crud
	PriceStore
	// products crud
	ProductStore
	LoadPricesByIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error)
	LoadProductsByIds(ctx context.Context, productIds ...string) ([]*models.StripeProduct, error)
	LoadPricesWithProductByPriceIds(ctx context.Context, priceIds ...string) ([]*models.StripePrice, error)
	LoadSubscriptionsPriceProduct(ctx context.Context, subscriptions ...*models.StripeSubscription) error
	LoadSubscriptionsByIds(ctx context.Context, subscriptionIds ...string) ([]*models.StripeSubscription, error)

	FindActiveSubscriptionsByCustomerIds(ctx context.Context, customerIds ...string) ([]*models.StripeSubscription, error)
	FindActiveSubscriptionByCustomerId(ctx context.Context, customerId string) (*models.StripeSubscription, error)

	FindActiveSubscriptionsByTeamIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.StripeSubscription, error)

	FindActiveSubscriptionsByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.StripeSubscription, error)

	LoadPricesByProductIds(ctx context.Context, productIds ...string) ([][]*models.StripePrice, error)
	LoadProductRoles(ctx context.Context, productIds ...string) ([][]*models.Role, error)

	FindSubscriptionsWithPriceProductByIds(ctx context.Context, subscriptionIds ...string) ([]*models.StripeSubscription, error)
}

type SubscriptionStore interface {
	// subscriptions crud
	IsFirstSubscription(ctx context.Context, customerID string) (bool, error)
	ListSubscriptions(ctx context.Context, input *shared.StripeSubscriptionListParams) ([]*models.StripeSubscription, error)
	CountSubscriptions(ctx context.Context, filter *shared.StripeSubscriptionListFilter) (int64, error)
	UpsertSubscription(ctx context.Context, sub *models.StripeSubscription) error
	UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription) error
}

type ProductStore interface {
	FindProductById(ctx context.Context, productId string) (*models.StripeProduct, error)
	ListProducts(ctx context.Context, input *shared.StripeProductListParams) ([]*models.StripeProduct, error)
	CountProducts(ctx context.Context, filter *shared.StripeProductListFilter) (int64, error)
	UpsertProduct(ctx context.Context, product *models.StripeProduct) error
	UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error
}

type PriceStore interface {
	FindActivePriceById(ctx context.Context, priceId string) (*models.StripePrice, error)
	ListPrices(ctx context.Context, input *shared.StripePriceListParams) ([]*models.StripePrice, error)
	CountPrices(ctx context.Context, filter *shared.StripePriceListFilter) (int64, error)
	UpsertPrice(ctx context.Context, price *models.StripePrice) error
	UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error
}

type CustomerStore interface {
	FindCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error)
	CreateCustomer(ctx context.Context, customer *models.StripeCustomer) (*models.StripeCustomer, error)
	ListCustomers(ctx context.Context, input *shared.StripeCustomerListParams) ([]*models.StripeCustomer, error)
	CountCustomers(ctx context.Context, filter *shared.StripeCustomerListFilter) (int64, error)
}

type PaymentStore interface {
	// team methods
	// Team()
	PaymentTeamStore
	// rbac methods
	PaymentRbacStore
	// payment methods
	PaymentStripeStore
}
type PaymentService interface {
	Client() PaymentClient

	Adapter() stores.StorageAdapterInterface

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

	// FindSubscriptionWithPriceBySessionId(ctx context.Context, sessionId string) (*models.StripeSubscription, error)

	FindSubscriptionWithPriceProductBySessionId(ctx context.Context, sessionId string) (*models.StripeSubscription, error)

	UpsertSubscriptionByIds(ctx context.Context, cutomerId string, subscriptionId string) error

	VerifyAndUpdateTeamSubscriptionQuantity(ctx context.Context, teamId uuid.UUID) error
}

type StripeService struct {
	logger *slog.Logger
	client PaymentClient
	// paymentStore PaymentStore
	adapter stores.StorageAdapterInterface
}

// Adapter implements PaymentService.
func (srv *StripeService) Adapter() stores.StorageAdapterInterface {
	return srv.adapter
}

var _ PaymentService = (*StripeService)(nil)

func NewPaymentService(
	client PaymentClient,
	adapter stores.StorageAdapterInterface,
) PaymentService {
	return &StripeService{
		client:  client,
		logger:  slog.Default(),
		adapter: adapter,
	}
}

func (s *StripeService) UpsertSubscriptionFromStripe(ctx context.Context, sub *stripe.Subscription) error {
	if sub == nil {
		return nil
	}
	var item *stripe.SubscriptionItem
	if len(sub.Items.Data) > 0 {
		item = sub.Items.Data[0]
	}
	if item == nil || item.Price == nil {
		return errors.New("price not found")
	}

	status := models.StripeSubscriptionStatus(sub.Status)
	err := s.adapter.Subscription().UpsertSubscription(
		ctx,
		&models.StripeSubscription{
			ID:                 sub.ID,
			StripeCustomerID:   sub.Customer.ID,
			Status:             models.StripeSubscriptionStatus(status),
			Metadata:           sub.Metadata,
			ItemID:             item.ID,
			PriceID:            item.Price.ID,
			Quantity:           item.Quantity,
			CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
			Created:            utils.Int64ToISODate(sub.Created),
			CurrentPeriodStart: utils.Int64ToISODate(item.CurrentPeriodStart),
			CurrentPeriodEnd:   utils.Int64ToISODate(item.CurrentPeriodEnd),
			EndedAt:            types.Pointer(utils.Int64ToISODate(sub.EndedAt)),
			CancelAt:           types.Pointer(utils.Int64ToISODate(sub.CancelAt)),
			CanceledAt:         types.Pointer(utils.Int64ToISODate(sub.CanceledAt)),
			TrialStart:         types.Pointer(utils.Int64ToISODate(sub.TrialStart)),
			TrialEnd:           types.Pointer(utils.Int64ToISODate(sub.TrialEnd)),
		},
	)
	return err
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
	return srv.adapter.Customer().CreateCustomer(ctx, stripeCustomer)
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
	return srv.adapter.Customer().CreateCustomer(ctx, stripeCustomer)
}

// FindCustomerByTeam implements PaymentService.
func (srv *StripeService) FindCustomerByTeam(ctx context.Context, teamId uuid.UUID) (*models.StripeCustomer, error) {
	return srv.adapter.Customer().FindCustomer(
		ctx,
		&stores.StripeCustomerFilter{
			TeamIds: []uuid.UUID{teamId},
		},
	)
}

// FindCustomerByUser implements PaymentService.
func (srv *StripeService) FindCustomerByUser(ctx context.Context, userId uuid.UUID) (*models.StripeCustomer, error) {
	return srv.adapter.Customer().FindCustomer(
		ctx,
		&stores.StripeCustomerFilter{
			UserIds: []uuid.UUID{userId},
		},
	)
}

// VerifyAndUpdateTeamSubscriptionQuantity implements PaymentService.
func (srv *StripeService) VerifyAndUpdateTeamSubscriptionQuantity(ctx context.Context, teamId uuid.UUID) error {
	customer, err := srv.adapter.Customer().FindCustomer(ctx, &stores.StripeCustomerFilter{
		TeamIds: []uuid.UUID{teamId},
	})
	if err != nil {
		return err
	}
	if customer == nil {
		return errors.New("no stripe customer id")
	}
	sub, err := srv.adapter.Subscription().FindActiveSubscriptionByCustomerId(ctx, customer.ID)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.New("no subscription")
	}

	count, err := srv.adapter.TeamMember().CountTeamMembers(ctx, &stores.TeamMemberFilter{
		TeamIds: []uuid.UUID{teamId},
	})
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	if sub.Quantity != count {
		_, err := srv.client.UpdateItemQuantity(
			sub.ItemID,
			sub.PriceID,
			count,
		)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (srv *StripeService) Client() PaymentClient {
	return srv.client
}

func (srv *StripeService) SyncPerms(ctx context.Context) error {
	var err error
	for productId, permission := range shared.StripeProductPermissionMap {
		err = func() error {
			product, err := srv.adapter.Product().FindProduct(ctx, &stores.StripeProductFilter{
				Ids: []string{productId},
			})
			if err != nil {
				return err
			}
			if product == nil {
				return errors.New("product not found")
			}
			perm, err := srv.adapter.Rbac().FindPermissionByName(ctx, permission)
			if err != nil {
				return err
			}
			if perm == nil {
				return errors.New("permission not found")
			}
			return srv.adapter.Rbac().CreateProductPermissions(ctx, product.ID, perm.ID)
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
		err = srv.UpsertProductFromStripe(ctx, product)
		if err != nil {
			srv.logger.Error("error upserting product", "product", product.ID, "error", err)
			continue
		}
	}
	return nil
}

func (s *StripeService) UpsertProductFromStripe(ctx context.Context, product *stripe.Product) error {
	if product == nil {
		return nil
	}
	var image *string
	if len(product.Images) > 0 {
		image = &product.Images[0]
	}
	param := &models.StripeProduct{
		ID:          product.ID,
		Active:      product.Active,
		Name:        product.Name,
		Description: &product.Description,
		Image:       image,
		Metadata:    product.Metadata,
	}
	return s.adapter.Product().UpsertProduct(ctx, param)
}

func (s *StripeService) UpsertPriceFromStripe(ctx context.Context, price *stripe.Price) error {
	if price == nil {
		return nil
	}
	val := &models.StripePrice{
		ID:         price.ID,
		ProductID:  price.Product.ID,
		Active:     price.Active,
		LookupKey:  &price.LookupKey,
		UnitAmount: &price.UnitAmount,
		Currency:   string(price.Currency),
		Type:       models.StripePricingType(price.Type),
		Metadata:   price.Metadata,
	}
	if price.Recurring != nil {
		recur := price.Recurring
		val.Interval = types.Pointer(models.StripePricingPlanInterval(recur.Interval))
		val.IntervalCount = types.Pointer(recur.IntervalCount)
		val.TrialPeriodDays = types.Pointer(recur.TrialPeriodDays)
	}
	return s.adapter.Price().UpsertPrice(ctx, val)
}

func (srv *StripeService) FindAndUpsertAllPrices(ctx context.Context) error {
	prices, err := srv.client.FindAllPrices()
	if err != nil {
		srv.logger.Error("error finding all prices", "error", err)
		return err
	}
	for _, price := range prices {
		err = srv.UpsertPriceFromStripe(ctx, price)
		if err != nil {
			srv.logger.Error("error upserting price", "price", price.ID, "error", err)
			continue
		}
	}
	return nil
}

func (srv *StripeService) FindSubscriptionWithPriceProductBySessionId(ctx context.Context, sessionId string) (*models.StripeSubscription, error) {
	checkoutSession, err := srv.client.FindCheckoutSessionByStripeId(sessionId)
	if err != nil {
		return nil, err
	}
	if checkoutSession == nil {
		return nil, errors.New("subscription not found")
	}
	if checkoutSession.Subscription == nil {
		return nil, errors.New("subscription not found")
	}

	subscription, err := srv.adapter.Subscription().FindSubscriptionsWithPriceProductByIds(ctx, checkoutSession.Subscription.ID)
	if err != nil {
		return nil, err
	}
	if len(subscription) == 0 {
		return nil, errors.New("subscription not found")
	}
	return subscription[0], nil
}

func (srv *StripeService) UpsertSubscriptionByIds(ctx context.Context, cutomerId, subscriptionId string) error {
	customer, err := srv.adapter.Customer().FindCustomer(ctx, &stores.StripeCustomerFilter{
		Ids: []string{cutomerId},
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
	err = srv.UpsertSubscriptionFromStripe(ctx, sub)
	if err != nil {
		return err
	}
	return nil
}

func (srv *StripeService) CreateCheckoutSession(ctx context.Context, stripeCustomerId string, priceId string) (string, error) {
	customer, err := srv.adapter.Customer().FindCustomer(ctx, &stores.StripeCustomerFilter{
		Ids: []string{stripeCustomerId},
	})
	if err != nil {
		return "", err
	}
	if customer == nil {
		return "", errors.New("customer not found")
	}
	var count int64
	if customer.TeamID != nil {
		team, err := srv.adapter.TeamGroup().FindTeam(ctx, &stores.TeamFilter{
			CustomerIds: []string{stripeCustomerId},
		})
		// team, err := srv.paymentStore.FindTeamByStripeCustomerId(ctx, stripeCustomerId)
		if err != nil {
			return "", err
		}
		if team == nil {
			return "", errors.New("team not found")
		}
		count, err = srv.adapter.TeamMember().CountTeamMembers(ctx, &stores.TeamMemberFilter{
			TeamIds: []uuid.UUID{team.ID},
		})
		if err != nil {
			return "", err
		}
	} else {
		count = 1
	}
	val, err := srv.adapter.Subscription().FindActiveSubscriptionByCustomerId(ctx, stripeCustomerId)
	if err != nil {
		return "", err
	}
	if val != nil {
		return "", errors.New("user already has a valid subscription")
	}
	firstSub, err := srv.adapter.Subscription().IsFirstSubscription(ctx, stripeCustomerId)
	if err != nil {
		return "", err
	}
	var trialDays *int64
	if firstSub {
		trialDays = types.Pointer(int64(14))
	}
	valPrice, err := srv.adapter.Price().FindActivePriceById(ctx, priceId)
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
	team, err := srv.adapter.TeamGroup().FindTeamByStripeCustomerId(ctx, stripeCustomerId)
	if err != nil {
		return "", err
	}
	if team == nil {
		return "", errors.New("team not found")
	}

	sub, err := srv.adapter.Subscription().FindActiveSubscriptionByCustomerId(ctx, stripeCustomerId)
	if err != nil {
		return "", err
	}
	if sub == nil {
		return "", errors.New("no subscription.  subscribe to access billing portal")
	}
	prods, err := srv.adapter.Product().ListProducts(ctx, &stores.StripeProductFilter{
		PaginatedInput: stores.PaginatedInput{
			PerPage: 100,
		},
		Active: types.OptionalParam[bool]{IsSet: true, Value: true},
	})
	if err != nil {
		return "", err
	}
	prodIds := make([]string, len(prods))
	for i, p := range prods {
		prodIds[i] = p.ID
	}
	prices, err := srv.adapter.Price().ListPrices(ctx, &stores.StripePriceFilter{
		PaginatedInput: stores.PaginatedInput{
			PerPage: 100,
		},
		Active:     types.OptionalParam[bool]{IsSet: true, Value: true},
		ProductIds: prodIds,
	})
	if err != nil {
		return "", err
	}
	grouped := mapper.MapToMany(prices, prodIds, func(p *models.StripePrice) string { return p.ProductID })
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
	url, err := srv.client.CreateBillingPortalSession(stripeCustomerId, config)
	if err != nil {
		log.Println(err)
		return "", errors.New("failed to create checkout session")
	}
	if url == nil {
		return "", errors.New("failed to create checkout session")
	}
	return url.URL, nil
}
