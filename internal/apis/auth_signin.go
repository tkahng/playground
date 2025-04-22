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
	}
}

type SigninDto struct {
	Email    string                `json:"email" form:"email" format:"email" example:"tkahng+01@gmail.com"`
	Password RequiredPasswordField `json:"password" form:"password" minimum:"8" example:"Password123!"`
}

type AuthenticatedInfoResponse struct {
	// SetCookieOutput
	SetCookie []http.Cookie `header:"Set-Cookie"`

	Body shared.UserInfoTokens `json:"body"`
}

func (api *Api) SignIn(ctx context.Context, input *struct{ Body *SigninDto }) (*AuthenticatedInfoResponse, error) {
	db := api.app.Db()
	action := api.app.NewAuthActions(db)
	password := input.Body.Password.String()
	params := &shared.AuthenticationInput{
		Email:             input.Body.Email,
		Provider:          shared.ProvidersCredentials,
		Password:          &password,
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
