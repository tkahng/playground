package core

import (
	"context"
	"errors"
	"log"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/types"
)

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

func (srv *StripeService) FindOrCreateCustomerFromUser(ctx context.Context, exec bob.Executor, user *models.User) (*models.StripeCustomer, error) {
	if user == nil {
		return nil, nil
	}
	userId := user.ID
	dbCus, err := repository.FindCustomerByUserId(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	if dbCus != nil {
		return dbCus, nil
	}
	stripeCus, err := srv.client.FindOrCreateCustomer(user.Email, userId)
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

func (srv *StripeService) CreateCheckoutSession(ctx context.Context, db bob.Executor, user *models.User, priceId string) (string, error) {
	dbcus, err := srv.FindOrCreateCustomerFromUser(ctx, db, user)
	if err != nil {
		return "", err
	}
	val, err := repository.FindLatestCheckoutSubscriptionByUserId(ctx, db, user.ID)
	if err != nil {
		return "", err
	}
	if val != nil {
		return "", errors.New("user already has a valid subscription")
	}
	firstSub, err := repository.IsFirstSubscription(ctx, db, user.ID)
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

func (s *StripeService) CreateBillingPortalSession(ctx context.Context, db bob.Executor, user *models.User) (string, error) {
	dbcus, err := s.FindOrCreateCustomerFromUser(ctx, db, user)
	if err != nil {
		return "", err
	}
	// verify user has a valid subscriptio
	sub, err := repository.FindLatestCheckoutSubscriptionByUserId(ctx, db, user.ID)
	if err != nil {
		return "", err
	}
	if sub == nil {
		return "", huma.Error400BadRequest("no subscription.  subscribe to access billing portal")
	}
	url, err := s.client.CreateBillingPortalSession(dbcus.StripeID)
	if err != nil {
		log.Println(err)
		return "", huma.Error500InternalServerError("failed to create checkout session")
	}
	if url == nil {
		return "", huma.Error500InternalServerError("failed to create checkout session")
	}
	return url.URL, nil
}
