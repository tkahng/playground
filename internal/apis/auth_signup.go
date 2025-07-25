package apis

import (
	"context"
	"fmt"

	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
)

type RequiredPasswordField string

func (r RequiredPasswordField) String() string {
	return string(r)
}

type SignupInput struct {
	Email    string                `json:"email" form:"email" format:"email" example:"tkahng+01@gmail.com"`
	Password RequiredPasswordField `json:"password" form:"password" minimum:"8" example:"Password123!"`
	Name     *string               `json:"name"`
}

func (api *Api) SignUp(ctx context.Context, input *struct{ Body SignupInput }) (*AuthenticatedInfoResponse, error) {
	action := api.App().Auth()
	password := input.Body.Password.String()
	hash, err := action.Password().HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}
	params := &services.AuthenticationInput{
		Email:             input.Body.Email,
		Provider:          models.ProvidersCredentials,
		Password:          &password,
		HashPassword:      &hash,
		Type:              models.ProviderTypeCredentials,
		Name:              input.Body.Name,
		ProviderAccountID: input.Body.Email,
	}
	user, err := action.Authenticate(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error authenticating user: %w", err)
	}
	dto, err := action.CreateAuthTokensFromEmail(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("error creating auth dto: %w", err)
	}
	if dto == nil {
		return nil, fmt.Errorf("error creating auth dto: %w", err)
	}
	return &AuthenticatedInfoResponse{
		Body: *ToApiUserInfoTokens(dto),
	}, nil
}
