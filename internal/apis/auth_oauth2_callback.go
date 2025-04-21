package apis

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
	"golang.org/x/oauth2"
)

func (api *Api) OAuth2CallbackPostOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "oauth-callback-post",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Oauth callback",
		Description: "Oauth callback",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

type OAuth2CallbackPostResponse struct {
	Body *shared.AuthenticatedDTO
}

func (api *Api) OAuth2CallbackPost(ctx context.Context, input *OAuth2CallbackInput) (*AuthenticatedResponse, error) {

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
	q.Add(string(shared.RefreshTokenType), dto.Tokens.RefreshToken)
	uri.RawQuery = q.Encode()
	fmt.Println(uri.String())

	return &AuthenticatedResponse{
		Body: dto.AuthenticatedDTO,
	}, nil
	// return TokenDtoFromUserWithApp(ctx, h.app, user, uuid.NewString())
}

type OAuth2CallbackInput struct {
	Code  string `json:"code" query:"code" required:"true" minLength:"1"`
	State string `json:"state" query:"state" required:"true" minLength:"1"`
	// Provider db.AuthProviders `json:"provider" path:"provider"`
}

func (h *Api) OAuth2CallbackGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "oauth2-callback-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "OAuth2 callback",
		Description: "Count the number of colors for all themes",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
		// Security: []map[string][]string{
		// 	middleware.BearerAuthSecurity("colors:read"),
		// },
	}
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
	q.Add(string(shared.RefreshTokenType), dto.Tokens.RefreshToken)
	uri.RawQuery = q.Encode()
	fmt.Println(uri.String())

	return &OAuth2CallbackGetResponse{
		Status: http.StatusTemporaryRedirect,
		Url:    uri.String(),
		// RefreshToken: dto.Tokens.RefreshToken,
	}, nil
	// return &OAuth2CallbackResponse{
	// 	Body: dto,
	// }, nil
	// return TokenDtoFromUserWithApp(ctx, h.app, user, uuid.NewString())
}

type CallbackOutput struct {
	shared.AuthenticatedDTO
	RedirectTo string `json:"redirect_to"`
}

func OAuth2Callback(ctx context.Context, api *Api, input *OAuth2CallbackInput) (*CallbackOutput, error) {
	authOpts := api.app.Settings().Auth
	db := api.app.Db()
	parsedState, err := core.ParseProviderStateToken(input.State, authOpts.StateToken)
	if err != nil {
		return nil, err
	}
	if parsedState == nil {
		return nil, fmt.Errorf("token not found")
	}
	if parsedState.Type != shared.StateTokenType {
		return nil, fmt.Errorf("invalid token type. want verification_token, got  %v", parsedState.Type)
	}
	var provider core.ProviderConfig
	switch parsedState.Provider {
	case shared.OAuthProvidersGithub:
		provider = &authOpts.OAuth2Config.Github
	case shared.OAuthProvidersGoogle:
		provider = &authOpts.OAuth2Config.Google
	default:
		return nil, fmt.Errorf("invalid provider %v", parsedState.Provider)
	}
	if provider == nil {
		return nil, fmt.Errorf("invalid provider %v", parsedState.Provider)
	}
	if !provider.Active() {
		return nil, fmt.Errorf("provider %v is not enabled", parsedState.Provider)
	}
	var redirectUrl string
	if parsedState.RedirectTo != "" {
		redirectUrl = parsedState.RedirectTo
	} else {
		redirectUrl = api.app.Cfg().AppConfig.AppUrl
	}
	var opts []oauth2.AuthCodeOption

	if provider.Pkce() {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", parsedState.CodeVerifier))
	}

	// fetch token
	token, err := provider.FetchToken(ctx, input.Code, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OAuth2 token. %w", err)
	}

	// fetch external auth user
	authUser, err := provider.FetchAuthUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OAuth2 user. %w", err)
	}
	params := &shared.AuthenticateUserParams{
		AvatarUrl:         &authUser.AvatarURL,
		Email:             authUser.Email,
		Name:              &authUser.Username,
		EmailVerifiedAt:   &authUser.Expiry,
		Provider:          models.Providers(parsedState.Provider),
		Type:              models.ProviderTypesOauth,
		ProviderAccountID: authUser.Id,
		AccessToken:       &authUser.AccessToken,
		RefreshToken:      &authUser.RefreshToken,
	}
	user, err := api.app.AuthenticateUser(ctx, db, params, true)
	if err != nil {
		return nil, fmt.Errorf("error at Oatuh2Callback: %w", err)

	}
	dto, err := api.app.CreateAuthDto(ctx, user.User.Email)
	if err != nil || dto == nil {
		return nil, fmt.Errorf("error creating auth dto: %w", err)
	}
	return &CallbackOutput{
		AuthenticatedDTO: *dto,
		RedirectTo:       redirectUrl,
	}, nil
}
