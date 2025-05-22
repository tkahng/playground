package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
)

type OAuth2AuthorizationUrlInput struct {
	Provider   shared.Providers `json:"provider"  query:"provider" form:"provider" enum:"google,github" required:"true"`
	RedirectTo string           `json:"redirect_to" query:"redirect_to" form:"redirect_to" format:"uri" required:"false"`
}

type OAuth2AuthorizationUrlOutput struct {
	Body struct {
		Url string `json:"url"`
	} `json:"body"`
}

func (h *Api) OAuth2AuthorizationUrl(ctx context.Context, input *OAuth2AuthorizationUrlInput) (*OAuth2AuthorizationUrlOutput, error) {
	if input == nil {
		return nil, huma.Error400BadRequest("input is required")
	}
	res, err := h.app.Auth().CreateOAuthUrl(
		ctx,
		input.Provider,
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
