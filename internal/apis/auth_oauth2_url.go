package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/db/models"
)

func (h *Api) OAuth2AuthorizationUrlOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "oauth2-authorization-url",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "OAuth2 authorization",
		Description: "Count the number of colors for all themes",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

type OAuth2AuthorizationUrlInput struct {
	Provider models.Providers `json:"provider" path:"provider" query:"provider" form:"provider"`
}

type OAuth2AuthorizationUrlOutput struct {
	Body struct {
		Url string `json:"url"`
	} `json:"body"`
}

func (h *Api) OAuth2AuthorizationUrl(ctx context.Context, input *OAuth2AuthorizationUrlInput) (*OAuth2AuthorizationUrlOutput, error) {
	// provider, err := auth.NewProviderByName(input.Provider)
	// res := provider.AuthURL()
	return &OAuth2AuthorizationUrlOutput{
		Body: struct {
			Url string `json:"url"`
		}{
			// Url: res,
		},
	}, nil

}

func (h *Api) OAuth2AuthorizationUrlsOperation() huma.Operation {
	return huma.Operation{
		OperationID: "oauth2-authorization-urls",
		Method:      http.MethodGet,
		Path:        "/auth/providers/authorization-url",
		Summary:     "OAuth2 authorization",
		Description: "Count the number of colors for all themes",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

type SkipTakeDTO struct {
	Skip int `json:"skip" default:"0" minimum:"0" required:"false"`
	Take int `json:"take" default:"10" minimum:"1" maximum:"50" required:"false"`
}

type PageMetaDTO struct {
	TotalCount int  `json:"total_count" default:"0" minimum:"0"`
	HasMore    bool `json:"has_more" default:"false"`
}

type OrderByDTO struct {
	OrderBy string `json:"order_by" default:"created_at" required:"false"`
	Sort    string `json:"sort" default:"asc" required:"false" enum:"asc,desc"`
}
