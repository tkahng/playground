package core

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/payment"
	"github.com/tkahng/authgo/internal/types"
)

func (srv *StripeService) FindSubscriptionWithPriceBySessionId(ctx context.Context, db queries.Queryer, sessionId string) (*models.StripeSubscription, error) {
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
	return queries.FindSubscriptionWithPriceById(ctx, db, sub.Subscription.ID)
}

func (srv *StripeService) UpsertSubscriptionByIds(ctx context.Context, db queries.Queryer, cutomerId, subscriptionId string) error {
	cus, err := queries.FindCustomerByStripeId(ctx, db, cutomerId)
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
	err = queries.UpsertSubscriptionFromStripe(ctx, db, sub, cus.ID)
	if err != nil {
		return err
	}
	return nil
}

func (srv *StripeService) FindOrCreateCustomerFromUser(ctx context.Context, exec queries.Queryer, userId uuid.UUID, email string) (*models.StripeCustomer, error) {
	dbCus, err := queries.FindCustomerByUserId(ctx, exec, userId)
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

	err = queries.UpsertCustomerStripeId(ctx, exec, userId, stripeCus.ID)
	if err != nil {
		return nil, err
	}
	return queries.FindCustomerByUserId(ctx, exec, userId)
}

func (srv *StripeService) CreateCheckoutSession(ctx context.Context, db queries.Queryer, userId uuid.UUID, priceId string) (string, error) {
	user, err := queries.FindUserById(ctx, db, userId)
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
	val, err := queries.FindLatestActiveSubscriptionByUserId(ctx, db, userId)
	if err != nil {
		return "", err
	}
	if val != nil {
		return "", errors.New("user already has a valid subscription")
	}
	firstSub, err := queries.IsFirstSubscription(ctx, db, userId)
	if err != nil {
		return "", err
	}
	var trialDays *int64
	if firstSub {
		trialDays = types.Pointer(int64(14))
	}
	valPrice, err := queries.FindValidPriceById(ctx, db, priceId)
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

func (s *StripeService) CreateBillingPortalSession(ctx context.Context, db queries.Queryer, userId uuid.UUID) (string, error) {
	user, err := queries.FindUserById(ctx, db, userId)
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
	sub, err := queries.FindLatestActiveSubscriptionByUserId(ctx, db, user.ID)
	if err != nil {
		return "", err
	}
	if sub == nil {
		return "", errors.New("no subscription.  subscribe to access billing portal")
	}
	prods, err := queries.ListProducts(ctx, db, &shared.StripeProductListParams{
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
	err = prods.LoadStripeProductProductStripePrices(
		ctx,
		db,
		models.SelectWhere.StripePrices.Active.EQ(true),
	)
	if err != nil {
		return "", err
	}
	prodssa := mapper.Map(prods, func(user *models.StripeProduct) *payment.ProductBillingConfigurationInput {
		return &payment.ProductBillingConfigurationInput{
			Product: &user.ID,
			Prices:  mapper.Map(user.R.ProductStripePrices, func(p *models.StripePrice) *string { return &p.ID }),
		}
	})
	config, err := s.client.CreatePortalConfiguration(prodssa...)
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
