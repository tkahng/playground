package conf

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tkahng/authgo/internal/shared"
)

type TokenOption struct {
	Type     shared.TokenType `form:"type" json:"type" enum:"authentication_token,password_reset_token,verification_token,refresh_token,state_token"`
	Secret   string           `form:"secret" json:"secret,omitempty"`
	Duration int64            `form:"duration" json:"duration"`
}

func (c *TokenOption) durationTime() time.Duration {
	return time.Duration(c.Duration) * time.Second
}

func (c *TokenOption) ExpiresAt() *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(c.durationTime()))
}

func (c *TokenOption) Expires() time.Time {
	return time.Now().Add(c.durationTime())
}

type AuthOptions struct {
	AccessToken        TokenOption `form:"access_token" json:"access_token"`
	PasswordResetToken TokenOption `form:"password_reset_token" json:"password_reset_token"`
	VerificationToken  TokenOption `form:"verification_token" json:"verification_token"`
	RefreshToken       TokenOption `form:"refresh_token" json:"refresh_token"`
	StateToken         TokenOption `form:"state_token" json:"state_token"`
	InviteToken        TokenOption `form:"invite_token" json:"invite_token"`
}

func NewTokenOptions() AuthOptions {
	return AuthOptions{
		VerificationToken: TokenOption{
			Type:     shared.TokenTypesVerificationToken,
			Secret:   string(shared.TokenTypesVerificationToken),
			Duration: 259200, // 3days
		},
		AccessToken: TokenOption{
			Type:     shared.TokenTypesAccessToken,
			Secret:   string(shared.TokenTypesAccessToken),
			Duration: 3600, // 1hr
		},
		PasswordResetToken: TokenOption{
			Type:     shared.TokenTypesPasswordResetToken,
			Secret:   string(shared.TokenTypesPasswordResetToken),
			Duration: 1800, // 30min
		},
		RefreshToken: TokenOption{
			Type:     shared.TokenTypesRefreshToken,
			Secret:   string(shared.TokenTypesRefreshToken),
			Duration: 604800, // 7days
		},
		StateToken: TokenOption{
			Type:     shared.TokenTypesStateToken,
			Secret:   string(shared.TokenTypesStateToken),
			Duration: 1800, // 30min
		},
		InviteToken: TokenOption{
			Type:     shared.TokenTypesInviteToken,
			Secret:   string(shared.TokenTypesInviteToken),
			Duration: 604800, // 7days
		},
	}
}

type AppOptions struct {
	Auth AuthOptions `form:"auth" json:"auth"`
	Smtp SmtpConfig  `form:"smtp" json:"smtp"`
	Meta AppConfig   `form:"meta" json:"meta"`
}
