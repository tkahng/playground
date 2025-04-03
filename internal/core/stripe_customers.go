package core

import (
	"context"
	"errors"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
)

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
