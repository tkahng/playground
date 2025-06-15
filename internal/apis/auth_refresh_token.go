package apis

import (
	"context"
)

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" cookie:"refresh_token" form:"refresh_token" required:"true"`
}

func (api *Api) RefreshToken(ctx context.Context, input *struct{ Body *RefreshTokenInput }) (*AuthenticatedInfoResponse, error) {
	action := api.app.Auth()
	claims, err := action.HandleRefreshToken(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &AuthenticatedInfoResponse{
		Body: *ToApiUserInfoTokens(claims),
	}, nil
}
