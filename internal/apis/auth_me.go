package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type MeOutput struct {
	Body *shared.UserWithAccounts
}

func (api *Api) Me(ctx context.Context, input *struct{}) (*MeOutput, error) {
	db := api.app.Db()
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	user, err := queries.FindUserById(ctx, db, claims.User.ID)
	if err != nil {
		return nil, err
	}
	accounts, err := queries.ListUserAccounts(ctx, db, &shared.UserAccountListParams{
		UserAccountListFilter: shared.UserAccountListFilter{UserIds: []string{user.ID.String()}},
	})
	if err != nil {
		return nil, err
	}
	return &MeOutput{
		Body: &shared.UserWithAccounts{
			User:     shared.FromCrudUser(user),
			Accounts: mapper.Map(accounts, shared.FromCrudUserAccountOutput),
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
	err := checker.CannotBeSuperUserID(ctx, claims.User.ID)
	if err != nil {
		return nil, err
	}
	// Check if the user has any active subscriptions
	err = checker.CannotHaveValidUserSubscription(ctx, claims.User.ID)
	if err != nil {
		return nil, err
	}
	err = queries.DeleteUsers(ctx, db, claims.User.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
