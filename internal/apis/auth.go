package apis

import "github.com/tkahng/playground/internal/models"

type ApiUserInfoTokens struct { // size=360 (0x168), class=384 (0x180)
	ApiUserInfo
	Tokens TokenDto `json:"tokens"`
}
type ApiUserInfo struct { // size=360 (0x168), class=384 (0x180)
	User        ApiUser        `db:"user" json:"user"`
	Roles       []string       `db:"roles" json:"roles"`
	Permissions []string       `db:"permissions" json:"permissions"`
	Providers   []models.Providers `db:"providers" json:"providers" enum:"google,apple,facebook,github,credentials"`
}

type TokenDto struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in" example:"3600"`
	TokenType   string `json:"token_type" example:"Bearer"`
	// Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}
