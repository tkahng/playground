package apis

import (
	"context"
	"fmt"

	"github.com/tkahng/authgo/internal/auth/oauth"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
	"golang.org/x/oauth2"
)

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

	conf := h.app.Cfg()
	action := h.app.Auth()
	redirectTo := input.RedirectTo
	if redirectTo == "" {
		redirectTo = conf.AppConfig.AppUrl
	}
	provider := oauth.NewProviderByName(string(input.Provider))
	if provider == nil {
		return nil, fmt.Errorf("provider %v not found", input.Provider)
	}
	if !provider.Active() {
		return nil, fmt.Errorf("provider %v is not enabled", input.Provider)
	}
	urlOpts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
	}
	info := &shared.ProviderStatePayload{
		Type:       shared.TokenTypesStateToken,
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
	state, err := action.CreateAndPersistStateToken(ctx, info)
	if err != nil {
		return nil, err
	}
	res := provider.BuildAuthURL(state, urlOpts...)

	return &OAuth2AuthorizationUrlOutput{
		Body: struct {
			Url string `json:"url"`
		}{
			Url: res,
		},
	}, nil

}
