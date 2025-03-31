package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
	"golang.org/x/oauth2"
)

func (h *Api) OAuth2AuthorizationUrlOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "oauth2-authorization-url",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "OAuth2 authorization",
		Description: "url for oauth2 authorization",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

type OAuth2AuthorizationUrlInput struct {
	Provider   shared.OAuthProviders `json:"provider"  query:"provider" form:"provider" enum:"google,github" required:"true"`
	RedirectTo string                `json:"redirect_to" query:"redirect_to" form:"redirect_to" format:"uri" required:"false"`
}

type OAuth2AuthorizationUrlOutput struct {
	Body struct {
		Url string `json:"url"`
	} `json:"body"`
}

func (h *Api) OAuth2AuthorizationUrl(ctx context.Context, input *OAuth2AuthorizationUrlInput) (*OAuth2AuthorizationUrlOutput, error) {

	settings := h.app.Settings()
	db := h.app.Db()
	redirectTo := input.RedirectTo
	if redirectTo == "" {
		redirectTo = settings.Meta.AppURL
	}
	provider, err := settings.Auth.OAuth2Config.GetProvider(string(input.Provider))
	if err != nil {
		return nil, err
	}
	if !provider.Active() {
		return nil, fmt.Errorf("provider %v is not enabled", input.Provider)
	}
	urlOpts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
	}
	info := &core.ProviderStatePayload{
		Type:       shared.StateTokenType,
		Provider:   input.Provider,
		RedirectTo: redirectTo,
		Token:      security.GenerateTokenKey(),
	}
	if provider.Pkce() {

		info.CodeVerifier = security.RandomString(43)
		info.CodeChallenge = security.S256Challenge(info.CodeVerifier)
		info.CodeChallengeMethod = "S256"
		urlOpts = append(urlOpts,
			oauth2.SetAuthURLParam("code_challenge", info.CodeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", info.CodeChallengeMethod),
		)
	}
	state, err := core.CreateAndPersistStateToken(ctx, db, info, settings.Auth.StateToken)
	if err != nil {
		return nil, err
	}
	res := provider.BuildAuthURL(state, urlOpts...)

	// provider, err := auth.NewProviderByName(input.Provider)
	// res := provider.AuthURL()
	return &OAuth2AuthorizationUrlOutput{
		Body: struct {
			Url string `json:"url"`
		}{
			Url: res,
		},
	}, nil

}

func (h *Api) OAuth2AuthorizationUrlsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "oauth2-authorization-urls",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "OAuth2 authorization",
		Description: "List of auth urls",
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
