package apis

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) AdminUsersOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-users",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin users",
		Description: "List of users",
		Tags:        []string{"Users", "Admin"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type UserDetail struct {
	*shared.User
	Roles    []*shared.RoleWithPermissions `json:"roles,omitempty" required:"false"`
	Accounts []*shared.UserAccountOutput   `json:"accounts,omitempty" required:"false"`
}

func (api *Api) AdminUsers(ctx context.Context, input *struct {
	shared.UserListParams
}) (*shared.PaginatedOutput[*UserDetail], error) {
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
	rolesmap := make(map[uuid.UUID][]*crudModels.Role)
	permsmap := make(map[uuid.UUID][]*crudModels.Permission)
	accountsmap := make(map[uuid.UUID][]*crudModels.UserAccount)
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
		for _, role := range roles {
			rolesmap[role.Key] = role.Data
		}
	}

	if slices.Contains(input.Expand, "permissions") {
		perms, err := queries.GetUserPermissions(ctx, db, userIds...)
		if err != nil {
			return nil, err
		}
		for _, perm := range perms {
			permsmap[perm.Key] = perm.Data
		}
	}

	if slices.Contains(input.Expand, "accounts") {
		accounts, err := queries.GetUserAccounts(ctx, db, userIds...)
		if err != nil {
			return nil, err
		}
		for _, account := range accounts {
			accountsmap[account.Key] = account.Data
		}
	}
	info := mapper.Map(users, func(user *crudModels.User) *UserDetail {
		var roles []*crudModels.Role
		if v, ok := rolesmap[user.ID]; ok {
			roles = v
		}

		var accounts []*crudModels.UserAccount
		if v, ok := accountsmap[user.ID]; ok {
			accounts = v
		}
		return &UserDetail{
			User: shared.FromCrudUser(user),
			Roles: mapper.Map(roles, func(role *crudModels.Role) *shared.RoleWithPermissions {
				var permissions []*crudModels.Permission
				if v, ok := permsmap[role.ID]; ok {
					permissions = v
				}
				return shared.FromCrudRoleWithPermissions(role, permissions)
			}),
			Accounts: mapper.Map(accounts, shared.FromCrudUserAccountOutput),
		}
	})

	return &shared.PaginatedOutput[*UserDetail]{
		Body: shared.PaginatedResponse[*UserDetail]{
			Data: info,
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminUsersCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-users-create",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Create user",
		Description: "Create user",
		Tags:        []string{"Users", "Admin"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminUsersCreate(ctx context.Context, input *struct {
	Body shared.UserCreateInput
}) (*struct {
	Body *shared.User
}, error) {
	db := api.app.Db()
	action := api.app.NewAuthActions()
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
	return &struct{ Body *shared.User }{Body: user}, nil

}

func (api *Api) AdminUsersDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-users-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Delete user",
		Description: "Delete user",
		Tags:        []string{"Admin", "Users"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminUsersDelete(ctx context.Context, input *struct {
	ID uuid.UUID `path:"user-id" format:"uuid" required:"true"`
}) (*struct{}, error) {
	db := api.app.Db()
	checker := api.app.NewChecker(ctx)
	err := checker.CannotBeSuperUserID(input.ID)
	if err != nil {
		return nil, err
	}
	// Check if the user has any active subscriptions
	err = checker.CannotHaveValidSubscription(input.ID)
	if err != nil {
		return nil, err
	}
	err = queries.DeleteUsers(ctx, db, input.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUsersUpdateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-users-update",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Update user",
		Description: "Update user",
		Tags:        []string{"Admin", "Users"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
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

func (api *Api) AdminUsersUpdatePasswordOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-users-update-password",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Update user password",
		Description: "Update user password",
		Tags:        []string{"Admin", "Users"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type UpdateUserPasswordInput struct {
	Password string `json:"password" required:"true" minLength:"8" maxLength:"100"`
}

func (api *Api) AdminUsersUpdatePassword(ctx context.Context, input *struct {
	ID   uuid.UUID `path:"user-id" format:"uuid" required:"true"`
	Body UpdateUserPasswordInput
}) (*struct{}, error) {
	db := api.app.Db()
	checker := api.app.NewChecker(ctx)
	err := checker.CannotBeSuperUserID(input.ID)
	if err != nil {
		return nil, err
	}
	err = queries.UpdateUserPassword(ctx, db, input.ID, input.Body.Password)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUsersGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-user-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Get user",
		Description: "Get user",
		Tags:        []string{"Admin", "Users"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminUsersGet(ctx context.Context, input *struct {
	UserID uuid.UUID `path:"user-id" json:"user_id" format:"uuid" required:"true"`
}) (*struct{ Body *shared.User }, error) {
	db := api.app.Db()
	user, err := queries.FindUserById(ctx, db, input.UserID)
	if err != nil {
		return nil, err
	}
	return &struct{ Body *shared.User }{Body: shared.FromCrudUser(user)}, nil
}
