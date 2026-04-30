package apis

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/auth"
)

type SigninDto struct {
	Email    string                `json:"email" form:"email" format:"email" example:"admin@k2dv.io"`
	Password RequiredPasswordField `json:"password" form:"password" minimum:"8" example:"Password123!" required:"true"`
}

type AuthenticatedInfoResponse struct {
	// SetCookieOutput
	SetCookie []http.Cookie `header:"Set-Cookie"`

	Body ApiUserInfoTokens `json:"body" required:"true"`
}

func (a *Api) bindSignin(api huma.API) {
	rl := newAuthRateLimiter(10, time.Minute)
	huma.Register(
		api,
		huma.Operation{
			OperationID: "signin",
			Method:      http.MethodPost,
			Path:        "/auth/signin",
			Summary:     "Sign in",
			Description: "Count the number of colors for all themes",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusNotFound, http.StatusTooManyRequests},
			Middlewares: huma.Middlewares{rl},
		},
		a.SignIn,
	)
}
func (api *Api) SignIn(ctx context.Context, input *struct {
	Body *SigninDto `json:"body" required:"true"`
}) (*AuthenticatedInfoResponse, error) {
	dto, err := api.App().Auth().Signin(ctx, &auth.SigninInput{
		Email:    input.Body.Email,
		Password: input.Body.Password.String(),
	})
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
