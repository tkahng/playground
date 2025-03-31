package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
)

type RequiredPasswordField string

func (r RequiredPasswordField) String() string {
	return string(r)
}

type SignupInput struct {
	Email    string                `json:"email" form:"email" format:"email" example:"tkahng@gmail.com"`
	Password RequiredPasswordField `json:"password" form:"password" minimum:"8" example:"Password123!"`
	Name     *string               `json:"name"`
}

func (api *Api) SignupOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "signup",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Sign up",
		Description: "Count the number of colors for all themes",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

func (api *Api) SignUp(ctx context.Context, input *struct{ Body *SignupInput }) (*AuthenticatedResponse, error) {
	db := api.app.Db()
	password := input.Body.Password.String()
	params := &shared.AuthenticateUserParams{
		Email:             input.Body.Email,
		Name:              input.Body.Name,
		EmailVerifiedAt:   nil,
		Provider:          "credentials",
		Password:          &password,
		Type:              "credentials",
		ProviderAccountID: input.Body.Email,
	}
	user, err := api.app.AuthenticateUser(ctx, db, params, true)
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
