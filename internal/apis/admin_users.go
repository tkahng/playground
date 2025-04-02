package apis

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/dataloader"
	"github.com/tkahng/authgo/internal/tools/security"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func (api *Api) AdminUsersOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-users",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin users",
		Description: "List of users",
		Tags:        []string{"Auth", "Admin"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type PaginatedOutput[T any] struct {
	Body shared.PaginatedResponse[T] `json:"body"`
}

type UserInfo struct {
	models.User
	Roles       []string           `json:"roles"`
	Permissions []string           `json:"permissions"`
	Providers   []models.Providers `json:"providers"`
}

func (api *Api) AdminUsers(ctx context.Context, input *struct {
	shared.UserListParams
}) (*PaginatedOutput[*UserInfo], error) {
	db := api.app.Db()
	utils.PrettyPrintJSON(input)
	users, err := repository.ListUsers(ctx, db, &input.UserListParams)
	if err != nil {
		return nil, err
	}
	count, err := repository.CountUsers(ctx, db, &input.UserListFilter)
	if err != nil {
		return nil, err
	}

	ids := dataloader.Map(users, func(user *models.User) uuid.UUID {
		return user.ID
	})
	m := make(map[uuid.UUID]*repository.RolePermissionClaims)
	claims, err := repository.GetUsersWithRolesAndPermissions(ctx, db, ids...)
	if err != nil {
		return nil, err
	}
	for _, claim := range claims {
		m[claim.UserID] = &claim
	}
	info := dataloader.Map(users, func(user *models.User) *UserInfo {
		claims := m[user.ID]
		return &UserInfo{
			User:        *user,
			Roles:       claims.Roles,
			Permissions: claims.Permissions,
			Providers:   claims.Providers,
		}
	})

	return &PaginatedOutput[*UserInfo]{
		Body: shared.PaginatedResponse[*UserInfo]{
			Data: info,
			Meta: shared.Meta{
				Page:    input.Page,
				PerPage: input.PerPage,
				Total:   int(count),
			},
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
		Tags:        []string{"Auth", "Admin"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type CreateUserInput struct {
	Email           string
	Name            *string
	AvatarUrl       *string
	EmailVerifiedAt *time.Time
	Password        *string
}

func (api *Api) AdminUsersCreate(ctx context.Context, input *struct {
	Body CreateUserInput
}) (*struct {
	Body models.User
}, error) {
	db := api.app.Db()
	existingUser, err := repository.GetUserByEmail(ctx, db, input.Body.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, huma.Error409Conflict("User already exists")
	}
	// if password is not provided, generate a random password
	if input.Body.Password == nil {
		// generate random password
		hash, err := security.CreateHash(uuid.NewString(), argon2id.DefaultParams)
		if err != nil {
			return nil, err
		}
		input.Body.Password = &hash
	}
	params := &shared.AuthenticateUserParams{
		Email:             input.Body.Email,
		Name:              input.Body.Name,
		AvatarUrl:         input.Body.AvatarUrl,
		EmailVerifiedAt:   input.Body.EmailVerifiedAt,
		Provider:          models.ProvidersCredentials,
		Password:          input.Body.Password,
		Type:              models.ProviderTypesCredentials,
		ProviderAccountID: input.Body.Email,
	}
	// create user
	user, err := repository.CreateUser(ctx, db, params)
	if err != nil {
		return nil, err
	}
	// create account
	account, err := repository.CreateAccount(ctx, db, user, params)
	if err != nil {
		return nil, err
	}
	fmt.Println(account)
	return &struct{ Body models.User }{Body: *user}, nil

}

func (api *Api) AdminUsersDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-users-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Delete user",
		Description: "Delete user",
		Tags:        []string{"Auth", "Admin", "Users"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminUsersDelete(ctx context.Context, input *struct {
	ID uuid.UUID `path:"id"`
}) (*struct{}, error) {
	db := api.app.Db()
	err := repository.DeleteUsers(ctx, db, input.ID)
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
		Tags:        []string{"Auth", "Admin", "Users"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminUsersUpdate(ctx context.Context, input *struct {
	ID   uuid.UUID `path:"id"`
	Body repository.UpdateUserInput
}) (*struct{}, error) {
	db := api.app.Db()
	err := repository.UpdateUser(ctx, db, input.ID, &input.Body)
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
		Tags:        []string{"Auth", "Admin", "Users"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type UpdateUserPasswordInput struct {
	Password string
}

func (api *Api) AdminUsersUpdatePassword(ctx context.Context, input *struct {
	ID   uuid.UUID `path:"id"`
	Body UpdateUserPasswordInput
}) (*struct{}, error) {
	db := api.app.Db()
	err := repository.UpdateUserPassword(ctx, db, input.ID, input.Body.Password)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
