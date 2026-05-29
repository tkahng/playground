package conf

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tkahng/playground/internal/models"
)

type TokenOption struct {
	Type     models.TokenTypes `form:"type" json:"type" enum:"authentication_token,password_reset_token,verification_token,refresh_token,state_token"`
	Secret   string            `form:"secret" json:"secret,omitempty"`
	Duration int64             `form:"duration" json:"duration"`
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

// tokenSecret derives a per-token-type HMAC secret from the master JWT secret.
// Using distinct per-type secrets prevents a token of one type being validated as another type
// even if the "type" claim check were ever bypassed.
func tokenSecret(base string, tokenType models.TokenTypes) string {
	return base + ":" + string(tokenType)
}

func NewTokenOptions(jwtSecret string) AuthOptions {
	return AuthOptions{
		VerificationToken: TokenOption{
			Type:     models.TokenTypesVerificationToken,
			Secret:   tokenSecret(jwtSecret, models.TokenTypesVerificationToken),
			Duration: 259200, // 3 days
		},
		AccessToken: TokenOption{
			Type:     models.TokenTypesAccessToken,
			Secret:   tokenSecret(jwtSecret, models.TokenTypesAccessToken),
			Duration: 3600, // 1 hr
		},
		PasswordResetToken: TokenOption{
			Type:     models.TokenTypesPasswordResetToken,
			Secret:   tokenSecret(jwtSecret, models.TokenTypesPasswordResetToken),
			Duration: 1800, // 30 min
		},
		RefreshToken: TokenOption{
			Type:     models.TokenTypesRefreshToken,
			Secret:   tokenSecret(jwtSecret, models.TokenTypesRefreshToken),
			Duration: 604800, // 7 days
		},
		StateToken: TokenOption{
			Type:     models.TokenTypesStateToken,
			Secret:   tokenSecret(jwtSecret, models.TokenTypesStateToken),
			Duration: 1800, // 30 min
		},
		InviteToken: TokenOption{
			Type:     models.TokenTypesInviteToken,
			Secret:   tokenSecret(jwtSecret, models.TokenTypesInviteToken),
			Duration: 604800, // 7 days
		},
	}
}
