package apis

import (
	"context"
	"fmt"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) AdminUsers(ctx context.Context, input *struct {
	shared.UserListParams
}) (*ApiPaginatedOutput[*shared.User], error) {
	db := api.app.Db()
	fmt.Printf("AdminUsers: %v", input.UserListParams)
	users, err := queries.ListUsers(ctx, db, &input.UserListParams)
	if err != nil {
		return nil, err
	}
	count, err := queries.CountUsers(ctx, db, &input.UserListFilter)
	if err != nil {
		return nil, err
	}
	userIds := []uuid.UUID{}
	userIdsstring := []string{}
	for _, user := range users {
		userIds = append(userIds, user.ID)
		userIdsstring = append(userIdsstring, user.ID.String())
	}
	if slices.Contains(input.Expand, "roles") {
		roles, err := queries.GetUserRoles(ctx, db, userIds...)
		if err != nil {
			return nil, err
		}
		for idx, user := range users {
			user.Roles = roles[idx]
		}
	}

	if slices.Contains(input.Expand, "accounts") {
		accounts, err := queries.GetUserAccounts(ctx, db, userIds...)
		if err != nil {
			return nil, err
		}
		for idx, user := range users {
			user.Accounts = accounts[idx]
		}
	}

	return &ApiPaginatedOutput[*shared.User]{
		Body: ApiPaginatedResponse[*shared.User]{
			Data: mapper.Map(users, shared.FromUserModel),
			Meta: GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminUsersCreate(ctx context.Context, input *struct {
	Body shared.UserCreateInput
}) (*struct {
	Body *shared.User
}, error) {
	db := api.app.Db()
	action := api.app.Auth()
	existingUser, err := queries.FindUserByEmail(ctx, db, input.Body.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, huma.Error409Conflict("User already exists")
	}
	user, err := action.Authenticate(ctx, &shared.AuthenticationInput{
		Email:             input.Body.Email,
		Provider:          shared.ProvidersCredentials,
		Password:          &input.Body.Password,
		Type:              shared.ProviderTypeCredentials,
		ProviderAccountID: input.Body.Email,
		EmailVerifiedAt:   input.Body.EmailVerifiedAt,
	})
	if err != nil {
		return nil, err
	}
	return &struct {
		Body *shared.User
	}{
		Body: shared.FromUserModel(user),
	}, nil

}

func (api *Api) AdminUsersDelete(ctx context.Context, input *struct {
	ID uuid.UUID `path:"user-id" format:"uuid" required:"true"`
}) (*struct{}, error) {
	db := api.app.Db()
	checker := api.app.Checker()
	ok, err := checker.CannotBeSuperUserID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot delete super user")
	}
	// Check if the user has any active subscriptions
	ok, err = checker.CannotHaveValidUserSubscription(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot delete user with active subscription")
	}
	err = queries.DeleteUsers(ctx, db, input.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUsersUpdate(ctx context.Context, input *struct {
	ID   uuid.UUID `path:"user-id" format:"uuid" required:"true"`
	Body shared.UserMutationInput
}) (*struct{}, error) {
	db := api.app.Db()
	err := queries.UpdateUser(ctx, db, input.ID, &input.Body)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type UpdateUserPasswordInput struct {
	Password string `json:"password" required:"true" minLength:"8" maxLength:"100"`
}

func (api *Api) AdminUsersUpdatePassword(ctx context.Context, input *struct {
	ID   uuid.UUID `path:"user-id" format:"uuid" required:"true"`
	Body UpdateUserPasswordInput
}) (*struct{}, error) {
	db := api.app.Db()
	checker := api.app.Checker()
	ok, err := checker.CannotBeSuperUserID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot update super user password")
	}
	err = queries.UpdateUserPassword(ctx, db, input.ID, input.Body.Password)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUsersGet(ctx context.Context, input *struct {
	UserID uuid.UUID `path:"user-id" json:"user_id" format:"uuid" required:"true"`
}) (*struct{ Body *shared.User }, error) {
	db := api.app.Db()
	user, err := queries.FindUserByID(ctx, db, input.UserID)
	if err != nil {
		return nil, err
	}
	return &struct{ Body *shared.User }{Body: shared.FromUserModel(user)}, nil
}
