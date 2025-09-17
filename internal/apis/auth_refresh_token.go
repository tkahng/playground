package apis

import (
	"context"
)

type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token" cookie:"refresh_token" form:"refresh_token" required:"true"`
}

func (api *Api) RefreshToken(ctx context.Context, input *struct{ Body *RefreshTokenInput }) (*AuthenticatedInfoResponse, error) {
	claims, err := api.App().Auth2().RefreshToken(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &AuthenticatedInfoResponse{
		Body: *ToApiUserInfoTokens(claims),
	}, nil
}
