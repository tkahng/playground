package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
)

type SigninDto struct {
	Email    string                `json:"email" form:"email" format:"email" example:"admin@k2dv.io"`
	Password RequiredPasswordField `json:"password" form:"password" minimum:"8" example:"Password123!"`
}

type AuthenticatedInfoResponse struct {
	// SetCookieOutput
	SetCookie []http.Cookie `header:"Set-Cookie"`

	Body ApiUserInfoTokens `json:"body"`
}

func (api *Api) SignIn(ctx context.Context, input *struct{ Body *SigninDto }) (*AuthenticatedInfoResponse, error) {
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
