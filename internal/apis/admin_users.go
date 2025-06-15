package apis

import (
	"context"
	"fmt"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type UserListFilter struct {
	PaginatedInput
	SortParams
	Providers     []ApiProviders            `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	Q             string                    `query:"q,omitempty" required:"false"`
	Ids           []string                  `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Emails        []string                  `query:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	RoleIds       []string                  `query:"role_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	EmailVerified types.OptionalParam[bool] `query:"email_verified,omitempty" required:"false"`
	Expand        []string                  `query:"expand,omitempty" required:"false" minimum:"1" uniqueItems:"true" enum:"roles,permissions,accounts,subscriptions"`
}

func (api *Api) AdminUsers(ctx context.Context, input *struct {
	UserListFilter
}) (*ApiPaginatedOutput[*shared.User], error) {
	adapter := api.app.Adapter()
	fmt.Printf("AdminUsers: %v", input.UserListFilter)
	filter := &stores.UserFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	filter.Q = input.Q
	filter.Providers = mapper.Map(input.Providers, func(p ApiProviders) models.Providers {
		return models.Providers(p.String())
	})
	filter.Ids = utils.ParseValidUUIDs(input.Ids...)
	filter.Emails = input.Emails
	filter.RoleIds = utils.ParseValidUUIDs(input.RoleIds...)
	filter.EmailVerified = input.EmailVerified

	users, err := adapter.User().FindUsers(ctx, filter)
	if err != nil {
		return nil, err
	}
	count, err := adapter.User().CountUsers(ctx, filter)
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
		roles, err := adapter.Rbac().GetUserRoles(ctx, userIds...)
		if err != nil {
			return nil, err
		}
		for idx, user := range users {
			user.Roles = roles[idx]
		}
	}

	if slices.Contains(input.Expand, "accounts") {
		accounts, err := adapter.UserAccount().GetUserAccounts(ctx, userIds...)
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
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminUsersCreate(ctx context.Context, input *struct {
	Body shared.UserCreateInput
}) (*struct {
	Body *shared.User
}, error) {
	action := api.app.Auth()
	adapter := api.app.Adapter()
	existingUser, err := adapter.User().FindUser(ctx, &stores.UserFilter{
		Emails: []string{input.Body.Email},
	})
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
	checker := api.app.Checker()
	adapter := api.app.Adapter()
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
	err = adapter.User().DeleteUser(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUsersUpdate(ctx context.Context, input *struct {
	ID   uuid.UUID `path:"user-id" format:"uuid" required:"true"`
	Body shared.UserMutationInput
}) (*struct{}, error) {
	adapter := api.app.Adapter()
	user, err := adapter.User().FindUserByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	user.Email = input.Body.Email
	user.Name = input.Body.Name
	user.Image = input.Body.Image
	user.EmailVerifiedAt = input.Body.EmailVerifiedAt
	err = adapter.User().UpdateUser(ctx, user)
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
	checker := api.app.Checker()
	ok, err := checker.CannotBeSuperUserID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot update super user password")
	}
	err = api.app.Adapter().UserAccount().UpdateUserPassword(ctx, input.ID, input.Body.Password)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUsersGet(ctx context.Context, input *struct {
	UserID uuid.UUID `path:"user-id" json:"user_id" format:"uuid" required:"true"`
}) (*struct{ Body *shared.User }, error) {
	user, err := api.app.Adapter().User().FindUserByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	return &struct{ Body *shared.User }{Body: shared.FromUserModel(user)}, nil
}
