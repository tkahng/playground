package apis

import (
	"context"
	"fmt"
)

type SignoutDto struct {
	RefreshToken string `json:"refresh_token" cookie:"refresh_token" form:"refresh_token" required:"true"`
}

func (api *Api) Signout(ctx context.Context, input *struct{ Body SignoutDto }) (*struct{}, error) {
	action := api.app.Auth()
	err := action.Signout(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("error signing out: %w", err)
	}
	return nil, nil
}
