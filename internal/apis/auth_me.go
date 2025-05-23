package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type MeOutput struct {
	Body *shared.UserWithAccounts
}

func (api *Api) Me(ctx context.Context, input *struct{}) (*MeOutput, error) {
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	user, err := api.app.User().Store().FindUserById(ctx, claims.User.ID)
	if err != nil {
		return nil, err
	}
	accounts, err := api.app.UserAccount().Store().GetUserAccounts(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	var acc []*models.UserAccount
	if len(accounts) > 0 {
		acc = accounts[0]
	}

	return &MeOutput{
		Body: &shared.UserWithAccounts{
			User:     shared.FromUserModel(user),
			Accounts: mapper.Map(acc, shared.FromModelUserAccountOutput),
		},
	}, nil

}

func (api *Api) MeUpdate(ctx context.Context, input *struct {
	Body shared.UpdateMeInput
}) (*struct{}, error) {
	db := api.app.Db()
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	err := queries.UpdateMe(ctx, db, claims.User.ID, &input.Body)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) MeDelete(ctx context.Context, input *struct{}) (*struct{}, error) {
	db := api.app.Db()
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	checker := api.app.Checker()
	ok, err := checker.CannotBeSuperUserID(ctx, claims.User.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error403Forbidden("You cannot delete the super user")
	}
	// Check if the user has any active subscriptions
	ok, err = checker.CannotHaveValidUserSubscription(ctx, claims.User.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error403Forbidden("You cannot delete a user with active subscriptions")
	}
	err = queries.DeleteUsers(ctx, db, claims.User.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
