package apis

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
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

type PaginatedOutput[T any] struct {
	Body shared.PaginatedResponse[T] `json:"body"`
}

type UserAccountDetail struct {
	ID        uuid.UUID            `db:"id,pk" json:"id"`
	UserID    uuid.UUID            `db:"user_id" json:"user_id"`
	Type      models.ProviderTypes `db:"type" json:"type" enum:"oauth,credentials"`
	Provider  models.Providers     `db:"provider" json:"providers,omitempty" required:"false" enum:"google,apple,facebook,github,credentials"`
	CreatedAt time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt time.Time            `db:"updated_at" json:"updated_at"`
}

type UserDetail struct {
	*shared.User
	Roles    []*shared.RoleWithPermissions `json:"roles,omitempty" required:"false"`
	Accounts []*shared.UserAccountOutput   `json:"accounts,omitempty" required:"false"`
}

func ToUserAccountDetail(userAccount *models.UserAccount) *UserAccountDetail {
	return &UserAccountDetail{
		ID:        userAccount.ID,
		UserID:    userAccount.UserID,
		Type:      userAccount.Type,
		Provider:  userAccount.Provider,
		CreatedAt: userAccount.CreatedAt,
		UpdatedAt: userAccount.UpdatedAt,
	}
}

func (api *Api) AdminUsers(ctx context.Context, input *struct {
	shared.UserListParams
}) (*PaginatedOutput[*UserDetail], error) {
	db := api.app.Db()
	users, err := repository.ListUsers(ctx, db, &input.UserListParams)
	if err != nil {
		return nil, err
	}
	count, err := repository.CountUsers(ctx, db, &input.UserListFilter)
	if err != nil {
		return nil, err
	}

	if slices.Contains(input.Expand, "roles") {
		err = users.LoadUserRoles(ctx, db)
		if err != nil {
			return nil, err
		}
	}

	if slices.Contains(input.Expand, "permissions") {
		err = users.LoadUserPermissions(ctx, db)
		if err != nil {
			return nil, err
		}
	}

	if slices.Contains(input.Expand, "accounts") {
		err = users.LoadUserUserAccounts(ctx, db)
		if err != nil {
			return nil, err
		}
	}
	info := mapper.Map(users, func(user *models.User) *UserDetail {
		return &UserDetail{
			User:     shared.ToUser(user),
			Roles:    mapper.Map(user.R.Roles, shared.ToRoleWithPermissions),
			Accounts: mapper.Map(user.R.UserAccounts, shared.ToUserAccountOutput),
		}
	})

	return &PaginatedOutput[*UserDetail]{
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

type CreateUserInput struct {
	Email           string     `json:"email" required:"true" format:"email" maxLength:"100"`
	Name            *string    `json:"name" required:"false" maxLength:"100"`
	AvatarUrl       *string    `json:"avatar_url" required:"false" format:"uri" maxLength:"200"`
	EmailVerifiedAt *time.Time `json:"email_verified_at" required:"false" format:"date-time"`
	Password        string     `json:"password" required:"true" minLength:"8" maxLength:"100"`
}

func (api *Api) AdminUsersCreate(ctx context.Context, input *struct {
	Body CreateUserInput
}) (*struct {
	Body *shared.User
}, error) {
	db := api.app.Db()
	action := api.app.NewAuthActions(db)
	existingUser, err := repository.FindUserByEmail(ctx, db, input.Body.Email)
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
	})

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
	ID uuid.UUID `path:"id" format:"uuid" required:"true"`
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
		Tags:        []string{"Admin", "Users"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminUsersUpdate(ctx context.Context, input *struct {
	ID   uuid.UUID `path:"id" format:"uuid" required:"true"`
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
	ID   uuid.UUID `path:"id" format:"uuid" required:"true"`
	Body UpdateUserPasswordInput
}) (*struct{}, error) {
	db := api.app.Db()
	err := repository.UpdateUserPassword(ctx, db, input.ID, input.Body.Password)
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
	ID uuid.UUID `path:"id" format:"uuid" required:"true"`
}) (*struct{ Body *shared.User }, error) {
	db := api.app.Db()
	user, err := repository.FindUserById(ctx, db, input.ID)
	if err != nil {
		return nil, err
	}
	err = user.LoadUserRoles(ctx,
		db,
		models.ThenLoadRolePermissions(),
	)
	if err != nil {
		return nil, err
	}
	return &struct{ Body *shared.User }{Body: shared.ToUser(user)}, nil
}
