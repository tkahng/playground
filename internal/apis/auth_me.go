package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/queries"
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
	claims := core.GetContextUserInfo(ctx)
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

func (api *Api) MeUpdateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "meUpdate",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Me Update",
		Description: "Me Update",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) MeUpdate(ctx context.Context, input *struct {
	Body shared.UpdateMeInput
}) (*struct{}, error) {
	db := api.app.Db()
	claims := core.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	err := queries.UpdateMe(ctx, db, claims.User.ID, &input.Body)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) MeDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "me-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Me delete",
		Description: "Me delete",
		Tags:        []string{"Auth", "Me"},
		Errors:      []int{http.StatusUnauthorized, http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) MeDelete(ctx context.Context, input *struct{}) (*struct{}, error) {
	db := api.app.Db()
	claims := core.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	checker := api.app.NewChecker(ctx)
	err := checker.CannotBeSuperUserID(claims.User.ID)
	if err != nil {
		return nil, err
	}
	// Check if the user has any active subscriptions
	err = checker.CannotHaveValidSubscription(claims.User.ID)
	if err != nil {
		return nil, err
	}
	err = queries.DeleteUsers(ctx, db, claims.User.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
