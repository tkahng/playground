package apis

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tkahng/authgo/internal/auth/oauth"
	"github.com/tkahng/authgo/internal/shared"
)

type OAuth2CallbackPostResponse struct {
	Body *shared.UserInfoTokens
}

func (api *Api) OAuth2CallbackPost(ctx context.Context, input *OAuth2CallbackInput) (*AuthenticatedInfoResponse, error) {

	dto, err := OAuth2Callback(ctx, api, input)
	if err != nil {
		return nil, err
	}
	redirectUrl := dto.RedirectTo
	uri, err := url.Parse(redirectUrl)
	if err != nil {
		return nil, err
	}
	q := uri.Query()
	q.Add(string(shared.TokenTypesRefreshToken), dto.Tokens.RefreshToken)
	uri.RawQuery = q.Encode()
	fmt.Println(uri.String())

	return &AuthenticatedInfoResponse{
		Body: dto.UserInfoTokens,
	}, nil
	// return TokenDtoFromUserWithApp(ctx, h.app, user, uuid.NewString())
}

type OAuth2CallbackInput struct {
	Code  string `json:"code" query:"code" required:"true" minLength:"1"`
	State string `json:"state" query:"state" required:"true" minLength:"1"`
	// Provider db.AuthProviders `json:"provider" path:"provider"`
}

type OAuth2CallbackGetResponse struct {
	Status int
	Url    string `header:"Location"`
	// Body   *shared.AuthenticatedDTO
}

func (api *Api) OAuth2CallbackGet(ctx context.Context, input *OAuth2CallbackInput) (*OAuth2CallbackGetResponse, error) {

	dto, err := OAuth2Callback(ctx, api, input)
	if err != nil {
		return nil, err
	}
	redirectUrl := dto.RedirectTo
	uri, err := url.Parse(redirectUrl)
	if err != nil {
		return nil, err
	}
	q := uri.Query()
	q.Add(string(shared.TokenTypesRefreshToken), dto.Tokens.RefreshToken)
	uri.RawQuery = q.Encode()
	fmt.Println(uri.String())

	return &OAuth2CallbackGetResponse{
		Status: http.StatusTemporaryRedirect,
		Url:    uri.String(),
		// RefreshToken: dto.Tokens.RefreshToken,
	}, nil

}

type CallbackOutput struct {
	shared.UserInfoTokens
	RedirectTo string `json:"redirect_to"`
}

func OAuth2Callback(ctx context.Context, api *Api, input *OAuth2CallbackInput) (*CallbackOutput, error) {
	action := api.app.Auth()
	parsedState, err := action.VerifyStateToken(ctx, input.State)
	if err != nil {
		return nil, err
	}
	if parsedState == nil {
		return nil, fmt.Errorf("token not found")
	}
	if parsedState.Type != shared.TokenTypesStateToken {
		return nil, fmt.Errorf("invalid token type. want verification_token, got  %v", parsedState.Type)
	}
	redirectUrl, authUser, shouldReturn, result, err := api.newFunction(ctx, input.Code, parsedState)
	if shouldReturn {
		return result, err
	}
	var prv shared.Providers
	switch parsedState.Provider {
	case shared.OAuthProvidersGithub:
		prv = shared.ProvidersGithub
	case shared.OAuthProvidersGoogle:
		prv = shared.ProvidersGoogle
	}
	params := &shared.AuthenticationInput{
		AvatarUrl:         &authUser.AvatarURL,
		Email:             authUser.Email,
		Name:              &authUser.Username,
		EmailVerifiedAt:   &authUser.Expiry,
		Provider:          prv,
		Type:              shared.ProviderTypeOAuth,
		ProviderAccountID: authUser.Id,
		AccessToken:       &authUser.AccessToken,
		RefreshToken:      &authUser.RefreshToken,
	}
	user, err := action.Authenticate(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error at Oatuh2Callback: %w", err)

	}
	dto, err := action.CreateAuthTokensFromEmail(ctx, user.Email)
	if err != nil || dto == nil {
		return nil, fmt.Errorf("error creating auth dto: %w", err)
	}
	return &CallbackOutput{
		UserInfoTokens: *dto,
		RedirectTo:     redirectUrl,
	}, nil
}

func (api *Api) newFunction(ctx context.Context, code string, parsedState *shared.ProviderStateClaims) (string, *oauth.AuthUser, bool, *CallbackOutput, error) {
	var provider oauth.ProviderConfig
	switch parsedState.Provider {
	case shared.OAuthProvidersGithub:
		provider = oauth.NewProviderByName(oauth.NameGithub)

	case shared.OAuthProvidersGoogle:
		provider = oauth.NewProviderByName(oauth.NameGoogle)
	default:
		return "", nil, true, nil, fmt.Errorf("invalid provider %v", parsedState.Provider)
	}
	if provider == nil {
		return "", nil, true, nil, fmt.Errorf("invalid provider %v", parsedState.Provider)
	}
	if !provider.Active() {
		return "", nil, true, nil, fmt.Errorf("provider %v is not enabled", parsedState.Provider)
	}
	var redirectUrl string
	if parsedState.RedirectTo != "" {
		redirectUrl = parsedState.RedirectTo
	} else {
		redirectUrl = api.app.Cfg().AppConfig.AppUrl
	}
	opts := provider.FetchTokenOptions(parsedState.CodeVerifier)

	// fetch token
	token, err := provider.FetchToken(ctx, code, opts...)
	if err != nil {
		return "", nil, true, nil, fmt.Errorf("failed to fetch OAuth2 token. %w", err)
	}

	// fetch external auth user
	authUser, err := provider.FetchAuthUser(ctx, token)
	if err != nil {
		return "", nil, true, nil, fmt.Errorf("failed to fetch OAuth2 user. %w", err)
	}
	return redirectUrl, authUser, false, nil, nil
}
