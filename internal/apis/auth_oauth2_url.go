package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/models"
)

type OAuth2AuthorizationUrlInput struct {
	Provider   models.Providers `json:"provider"  query:"provider" form:"provider" enum:"google,github" required:"true"` // only support google and github for now
	RedirectTo string           `json:"redirect_to,omitempty" query:"redirect_to" form:"redirect_to" format:"uri" required:"false"`
	Email      string           `json:"email,omitempty" query:"email" form:"email" format:"email" required:"false"`
}

type OAuth2AuthorizationUrlOutput struct {
	Body struct {
		Url string `json:"url"`
	} `json:"body"`
}

func (a *Api) OAuth2AuthorizationUrlBind(api huma.API) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "oauth2-authorization-url",
			Method:      http.MethodPost,
			Path:        "/auth/authorization-url",
			Summary:     "OAuth2 Authorization URL",
			Description: "Create OAuth2 authorization URL",
			Tags:        []string{"Auth", "OAuth2"},
			Errors:      []int{http.StatusNotFound},
		},
		func(ctx context.Context, input *struct {
			Body OAuth2AuthorizationUrlInput
		}) (*OAuth2AuthorizationUrlOutput, error) {
			if input == nil {
				return nil, huma.Error400BadRequest("input is required")
			}
			res, err := a.App().Auth().OAuth2Url(
				ctx,
				models.Providers(input.Body.Provider),
				input.Body.RedirectTo,
				input.Body.Email,
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
		},
	)
}
