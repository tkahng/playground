package core

import (
	"github.com/tkahng/authgo/internal/shared"
)

type BaseOAuth2ProviderEnv struct {
	ClientID     string                `form:"client_id" json:"client_id"`
	ClientSecret string                `form:"client_secret" json:"client_secret"`
	Type         shared.OAuthProviders `form:"type" json:"type"`
}

type OAuth2ProviderEnv struct {
	BaseOAuth2ProviderEnv
	AuthURL     string `form:"auth_url" json:"auth_url"`
	TokenURL    string `form:"token_url" json:"token_url"`
	UserInfoURL string `form:"user_info_url" json:"user_info_url"`
	DisplayName string `form:"display_name" json:"display_name"`
	// Extra       map[string]any
}
