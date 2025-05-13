package auth

import (
	"context"
	"encoding/json"
	"io"
	"strconv"

	"golang.org/x/oauth2"
)

// func init() {
// 	Providers[NameGithub] = wrapFactory(NewGithubProvider)
// }

// func NewGithubProvider() *GithubConfig {
// 	return &GithubConfig{OAuth2ProviderConfig{
// 		// ctx:         context.Background(),
// 		// displayName: "GitHub",
// 		// pkce:        true, // technically is not supported yet but it is safe as the PKCE params are just ignored
// 		// scopes:      []string{"read:user", "user:email"},
// 		// authURL:     github.Endpoint.AuthURL,
// 		// tokenURL:    github.Endpoint.TokenURL,
// 		// userInfoURL: "https://api.github.com/user",
// 		Name:        "GitHub",
// 		Enabled:     true,
// 		PKCE:        true, // technically is not supported yet but it is safe as the PKCE params are just ignored
// 		Scopes:      []string{"read:user", "user:email"},
// 		AuthURL:     github.Endpoint.AuthURL,
// 		TokenURL:    github.Endpoint.TokenURL,
// 		UserInfoURL: "https://api.github.com/user",
// 	}}
// }

type GithubConfig struct {
	OAuth2ProviderConfig
}

var _ ProviderConfig = (*GithubConfig)(nil)

func (p *GithubConfig) Active() bool {
	return p.Enabled
}

// FetchAuthUser implements Provider.FetchAuthUser() interface method.
func (p *GithubConfig) FetchAuthUser(ctx context.Context, token *oauth2.Token) (*AuthUser, error) {
	data, err := p.FetchRawUserInfo(ctx, token)

	if err != nil {
		return nil, err
	}

	rawUser := map[string]any{}
	if err := json.Unmarshal(data, &rawUser); err != nil {
		return nil, err
	}
	extracted := struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Id        int64  `json:"id"`
	}{}
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil, err
	}

	user := &AuthUser{
		Id:           strconv.FormatInt(extracted.Id, 10),
		Name:         extracted.Name,
		Username:     extracted.Login,
		Email:        extracted.Email,
		AvatarURL:    extracted.AvatarURL,
		RawUser:      rawUser,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	// in case user has set "Keep my email address private", send an
	// **optional** API request to retrieve the verified primary email
	if user.Email == "" {
		email, err := p.fetchPrimaryEmail(ctx, token)
		if err != nil {
			return nil, err
		}
		user.Email = email
	}

	return user, nil
}
func (p *GithubConfig) fetchPrimaryEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	client := p.Client(ctx, token)

	response, err := client.Get(p.UserInfoURL + "/emails")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// ignore common http errors caused by insufficient scope permissions
	// (the email field is optional, aka. return the auth user without it)
	if response.StatusCode == 401 || response.StatusCode == 403 || response.StatusCode == 404 {
		return "", nil
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	emails := []struct {
		Email    string
		Verified bool
		Primary  bool
	}{}
	if err := json.Unmarshal(content, &emails); err != nil {
		return "", err
	}

	// extract the verified primary email
	for _, email := range emails {
		if email.Verified && email.Primary {
			return email.Email, nil
		}
	}

	return "", nil
}
