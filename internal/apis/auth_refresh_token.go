package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" cookie:"refresh_token" form:"refresh_token" required:"true"`
}

func (a *Api) bindRefreshToken(api huma.API) {
	rl := newAuthRateLimiter(10, time.Minute)
	huma.Register(
		api,
		huma.Operation{
			OperationID: "refresh-token",
			Method:      http.MethodPost,
			Path:        "/auth/refresh-token",
			Summary:     "Refresh token",
			Description: "Exchange a refresh token for a new token pair",
			Tags:        []string{"Auth"},
			Errors:      []int{http.StatusUnauthorized, http.StatusTooManyRequests},
			Middlewares: huma.Middlewares{rl},
		},
		a.RefreshToken,
	)
}
func (api *Api) RefreshToken(ctx context.Context, input *struct{ Body *RefreshTokenInput }) (*AuthenticatedInfoResponse, error) {
	claims, err := api.App().Auth().RefreshToken(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &AuthenticatedInfoResponse{
		Body: *ToApiUserInfoTokens(claims),
	}, nil
}
