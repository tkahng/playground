package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tkahng/authgo/internal/shared"
)

type SigninDto struct {
	Email    string                `json:"email" form:"email" format:"email" example:"admin@k2dv.io"`
	Password RequiredPasswordField `json:"password" form:"password" minimum:"8" example:"Password123!"`
}

type AuthenticatedInfoResponse struct {
	// SetCookieOutput
	SetCookie []http.Cookie `header:"Set-Cookie"`

	Body shared.UserInfoTokens `json:"body"`
}

func (api *Api) SignIn(ctx context.Context, input *struct{ Body *SigninDto }) (*AuthenticatedInfoResponse, error) {
	action := api.app.Auth()
	password := input.Body.Password.String()
	hash, err := action.Password().HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}
	params := &shared.AuthenticationInput{
		Email:             input.Body.Email,
		Provider:          shared.ProvidersCredentials,
		Password:          &hash,
		Type:              shared.ProviderTypeCredentials,
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
		Body: *dto,
	}, nil

}
