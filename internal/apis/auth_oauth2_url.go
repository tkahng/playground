package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/models"
)

type OAuth2AuthorizationUrlInput struct {
	Provider   models.Providers `json:"provider"  query:"provider" form:"provider" enum:"google,github" required:"true"`
	RedirectTo string           `json:"redirect_to" query:"redirect_to" form:"redirect_to" format:"uri" required:"false"`
}

type OAuth2AuthorizationUrlOutput struct {
	Body struct {
		Url string `json:"url"`
	} `json:"body"`
}

func (api *Api) OAuth2AuthorizationUrl(ctx context.Context, input *OAuth2AuthorizationUrlInput) (*OAuth2AuthorizationUrlOutput, error) {
	if input == nil {
		return nil, huma.Error400BadRequest("input is required")
	}
	res, err := api.App().Auth().CreateOAuthUrl(
		ctx,
		models.Providers(input.Provider),
		input.RedirectTo,
	)
	if err != nil {
		return nil, err
	}

	return &OAuth2AuthorizationUrlOutput{
		Body: struct {
			Url string `json:"url"`
		}{
			Url: res,
		},
	}, nil

}
