package core

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/types"
)

func (srv *StripeService) FindSubscriptionWithPriceBySessionId(ctx context.Context, db bob.Executor, sessionId string) (*models.StripeSubscription, error) {
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
	return repository.FindSubscriptionWithPriceById(ctx, db, sub.Subscription.ID)
}

func (srv *StripeService) UpsertSubscriptionByIds(ctx context.Context, db bob.Executor, cutomerId, subscriptionId string) error {
	cus, err := repository.FindCustomerByStripeId(ctx, db, cutomerId)
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
	err = repository.UpsertSubscriptionFromStripe(ctx, db, sub, cus.ID)
	if err != nil {
		return err
	}
	return nil
}

func (srv *StripeService) FindOrCreateCustomerFromUser(ctx context.Context, exec bob.Executor, userId uuid.UUID, email string) (*models.StripeCustomer, error) {
	dbCus, err := repository.FindCustomerByUserId(ctx, exec, userId)
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

	err = repository.UpsertCustomerStripeId(ctx, exec, userId, stripeCus.ID)
	if err != nil {
		return nil, err
	}
	return repository.FindCustomerByUserId(ctx, exec, userId)
}

func (srv *StripeService) CreateCheckoutSession(ctx context.Context, db bob.Executor, userId uuid.UUID, priceId string) (string, error) {
	user, err := repository.FindUserById(ctx, db, userId)
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
	val, err := repository.FindLatestActiveSubscriptionByUserId(ctx, db, userId)
	if err != nil {
		return "", err
	}
	if val != nil {
		return "", errors.New("user already has a valid subscription")
	}
	firstSub, err := repository.IsFirstSubscription(ctx, db, userId)
	if err != nil {
		return "", err
	}
	var trialDays *int64
	if firstSub {
		trialDays = types.Pointer(int64(14))
	}
	valPrice, err := repository.FindValidPriceById(ctx, db, priceId)
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

func (s *StripeService) CreateBillingPortalSession(ctx context.Context, db bob.Executor, userId uuid.UUID) (string, error) {
	user, err := repository.FindUserById(ctx, db, userId)
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
	sub, err := repository.FindLatestActiveSubscriptionByUserId(ctx, db, user.ID)
	if err != nil {
		return "", err
	}
	if sub == nil {
		return "", errors.New("no subscription.  subscribe to access billing portal")
	}
	url, err := s.client.CreateBillingPortalSession(dbcus.StripeID)
	if err != nil {
		log.Println(err)
		return "", errors.New("failed to create checkout session")
	}
	if url == nil {
		return "", errors.New("failed to create checkout session")
	}
	return url.URL, nil
}
