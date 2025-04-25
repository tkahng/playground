package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) MeOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "me",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Me",
		Description: "Me",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type MeOutput struct {
	Body *shared.UserWithAccounts
}

func (api *Api) Me(ctx context.Context, input *struct{}) (*MeOutput, error) {
	db := api.app.Db()
	claims := core.GetContextUserClaims(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	user, err := repository.FindUserById(ctx, db, claims.User.ID)
	if err != nil {
		return nil, err
	}
	err = user.LoadUserUserAccounts(
		ctx,
		db,
	)
	if err != nil {
		return nil, err
	}
	return &MeOutput{
		Body: &shared.UserWithAccounts{
			User:     shared.ToUser(user),
			Accounts: mapper.Map(user.R.UserAccounts, shared.ToUserAccountOutput),
		},
	}, nil

}
