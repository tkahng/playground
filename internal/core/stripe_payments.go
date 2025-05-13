package core

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/payment"
	"github.com/tkahng/authgo/internal/tools/types"
)

func (srv *StripeService) FindSubscriptionWithPriceBySessionId(ctx context.Context, db db.Dbx, sessionId string) (*models.SubscriptionWithPrice, error) {
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
	data, err := srv.store.FindSubscriptionWithPriceById(ctx, db, sub.Subscription.ID)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (srv *StripeService) UpsertSubscriptionByIds(ctx context.Context, db db.Dbx, cutomerId, subscriptionId string) error {
	cus, err := srv.store.FindCustomerByStripeId(ctx, db, cutomerId)
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
	err = srv.store.UpsertSubscriptionFromStripe(ctx, db, sub, cus.ID)
	if err != nil {
		return err
	}
	return nil
}

func (srv *StripeService) FindOrCreateCustomerFromUser(ctx context.Context, db db.Dbx, userId uuid.UUID, email string) (*models.StripeCustomer, error) {
	dbCus, err := srv.store.FindCustomerByUserId(ctx, db, userId)
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

	err = srv.store.UpsertCustomerStripeId(ctx, db, userId, stripeCus.ID)
	if err != nil {
		return nil, err
	}
	return srv.store.FindCustomerByUserId(ctx, db, userId)
}

func (srv *StripeService) CreateCheckoutSession(ctx context.Context, db db.Dbx, userId uuid.UUID, priceId string) (string, error) {
	user, err := srv.store.FindUserById(ctx, db, userId)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}
	dbcus, err := srv.FindOrCreateCustomerFromUser(ctx, db, user.ID, user.Email)
	if err != nil {
		return "", err
	}
	val, err := srv.store.FindLatestActiveSubscriptionByUserId(ctx, db, userId)
	if err != nil {
		return "", err
	}
	if val != nil {
		return "", errors.New("user already has a valid subscription")
	}
	firstSub, err := srv.store.IsFirstSubscription(ctx, db, userId)
	if err != nil {
		return "", err
	}
	var trialDays *int64
	if firstSub {
		trialDays = types.Pointer(int64(14))
	}
	valPrice, err := srv.store.FindValidPriceById(ctx, db, priceId)
	if err != nil {
		return "", err
	}
	if valPrice == nil {
		return "", errors.New("price is not valid")
	}
	sesh, err := srv.client.CreateCheckoutSession(dbcus.StripeID, priceId, trialDays)
	if err != nil {
		return "", err
	}
	return sesh.URL, nil
}

func (s *StripeService) CreateBillingPortalSession(ctx context.Context, db db.Dbx, userId uuid.UUID) (string, error) {
	user, err := s.store.FindUserById(ctx, db, userId)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}
	// find or create customer from user
	dbcus, err := s.FindOrCreateCustomerFromUser(ctx, db, user.ID, user.Email)
	if err != nil {
		return "", err
	}
	if dbcus == nil {
		return "", errors.New("customer not found")
	}
	// verify user has a valid subscriptio
	sub, err := s.store.FindLatestActiveSubscriptionByUserId(ctx, db, user.ID)
	if err != nil {
		return "", err
	}
	if sub == nil {
		return "", errors.New("no subscription.  subscribe to access billing portal")
	}
	prods, err := s.store.ListProducts(ctx, db, &shared.StripeProductListParams{
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
	prices, err := s.store.ListPrices(ctx, db, &shared.StripePriceListParams{
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
	var configurations []*payment.ProductBillingConfigurationInput
	for i, id := range prods {
		price := grouped[i]
		con := &payment.ProductBillingConfigurationInput{
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
	url, err := s.client.CreateBillingPortalSession2(dbcus.StripeID, config)
	if err != nil {
		log.Println(err)
		return "", errors.New("failed to create checkout session")
	}
	if url == nil {
		return "", errors.New("failed to create checkout session")
	}
	return url.URL, nil
}
