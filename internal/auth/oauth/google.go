package oauth

import (
	"context"
	"encoding/json"

	"golang.org/x/oauth2"
)

type GoogleConfig struct {
	OAuth2ProviderConfig
}

var _ ProviderConfig = (*GoogleConfig)(nil)

func (p *GoogleConfig) Active() bool {
	return p.Enabled
}

func (p *GoogleConfig) FetchAuthUser(ctx context.Context, token *oauth2.Token) (*AuthUser, error) {
	data, err := p.FetchRawUserInfo(ctx, token)
	if err != nil {
		return nil, err
	}

	rawUser := map[string]any{}
	if err := json.Unmarshal(data, &rawUser); err != nil {
		return nil, err
	}
	extracted := struct {
		Id            string `json:"sub"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}{}
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil, err
	}

	user := &AuthUser{
		Id:           extracted.Id,
		Name:         extracted.Name,
		AvatarURL:    extracted.Picture,
		RawUser:      rawUser,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	if extracted.EmailVerified {
		user.Email = extracted.Email
	}

	return user, nil
}
