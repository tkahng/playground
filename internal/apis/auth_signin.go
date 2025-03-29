package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) SigninOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "signin",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Sign in",
		Description: "Count the number of colors for all themes",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
		// Security: []map[string][]string{
		// 	middleware.BearerAuthSecurity("colors:read"),
		// },
	}
}

type SigninDto struct {
	Email    string                `json:"email" form:"email" format:"email" example:"tkahng@gmail.com"`
	Password RequiredPasswordField `json:"password" form:"password" minimum:"8" example:"Password123!"`
}

func (api *Api) SignIn(ctx context.Context, input *struct{ Body *SigninDto }) (*AuthenticatedResponse, error) {
	db := api.app.Db()

	password := input.Body.Password.String()
	params := &shared.AuthenticateUserParams{
		Email:             input.Body.Email,
		Provider:          "credentials",
		Password:          &password,
		Type:              "credentials",
		ProviderAccountID: input.Body.Email,
	}
	user, err := api.app.AuthenticateUser(ctx, db, params, false)
	// 	user, err := AuthenticateUser(ctx, db, params)
	if err != nil {
		return nil, fmt.Errorf("error authenticating user: %w", err)
	}
	dto, err := api.app.CreateAuthDto(ctx, user.User.Email)
	if err != nil {
		return nil, fmt.Errorf("error creating auth dto: %w", err)
	}
	return &AuthenticatedResponse{
		Body: *dto,
	}, nil

}
