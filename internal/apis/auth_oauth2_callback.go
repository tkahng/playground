package apis

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
)

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
	q.Add(string(models.TokenTypesRefreshToken), dto.Tokens.RefreshToken)
	uri.RawQuery = q.Encode()
	fmt.Println(uri.String())

	return &AuthenticatedInfoResponse{
		Body: dto.ApiUserInfoTokens,
	}, nil
}

type OAuth2CallbackInput struct {
	Code  string `json:"code" query:"code" required:"true" minLength:"1"`
	State string `json:"state" query:"state" required:"true" minLength:"1"`
	// Provider db.AuthProviders `json:"provider" path:"provider"`
}

type OAuth2CallbackGetResponse struct {
	Status int
	Url    string `header:"Location"`
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
	q.Add(string(models.TokenTypesRefreshToken), dto.Tokens.RefreshToken)
	uri.RawQuery = q.Encode()
	fmt.Println(uri.String())

	return &OAuth2CallbackGetResponse{
		Status: http.StatusTemporaryRedirect,
		Url:    uri.String(),
		// RefreshToken: dto.Tokens.RefreshToken,
	}, nil

}

func ToApiUserInfoTokens(userInfo *models.UserInfoTokens) *ApiUserInfoTokens {
	if userInfo == nil {
		return nil
	}
	return &ApiUserInfoTokens{
		ApiUserInfo: ApiUserInfo{
			User:        *FromUserModel(&userInfo.User),
			Roles:       userInfo.Roles,
			Permissions: userInfo.Permissions,
			Providers:   userInfo.Providers,
		},
		Tokens: TokenDto{
			AccessToken:  userInfo.Tokens.AccessToken,
			ExpiresIn:    userInfo.Tokens.ExpiresIn,
			TokenType:    userInfo.Tokens.TokenType,
			RefreshToken: userInfo.Tokens.RefreshToken,
		},
	}
}

type CallbackOutput struct {
	ApiUserInfoTokens
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
	if parsedState.Type != models.TokenTypesStateToken {
		return nil, fmt.Errorf("invalid token type. want verification_token, got  %v", parsedState.Type)
	}
	authUser, err := action.FetchAuthUser(ctx, input.Code, parsedState)
	if err != nil {
		return nil, fmt.Errorf("error at Oatuh2Callback: %w", err)
	}
	params := &services.AuthenticationInput{
		AvatarUrl:         &authUser.AvatarURL,
		Email:             authUser.Email,
		Name:              &authUser.Username,
		EmailVerifiedAt:   &authUser.Expiry,
		Provider:          models.Providers(parsedState.Provider),
		Type:              models.ProviderTypeOAuth,
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
		ApiUserInfoTokens: *ToApiUserInfoTokens(dto),
		RedirectTo:        parsedState.RedirectTo,
	}, nil
}
