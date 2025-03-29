package core

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tkahng/authgo/internal/shared"
)

type TokenConfig struct {
	Name        string           `form:"name" json:"name"`
	DisplayName string           `form:"display_name" json:"display_name"`
	Type        shared.TokenType `form:"type" json:"type" enum:"authentication_token,password_reset_token,verification_token,refresh_token,state_token"`
	Secret      string           `form:"secret" json:"secret,omitempty"`
	// Duration specifies how long an issued token to be valid (in seconds)
	Duration int64 `form:"duration" json:"duration"`
}

func (c *TokenConfig) DurationTime() time.Duration {
	return time.Duration(c.Duration) * time.Second
}

func (c *TokenConfig) ExpiresAt() *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(c.DurationTime()))
}

func (c *TokenConfig) Expires() time.Time {
	return time.Now().Add(c.DurationTime())
}

func DurationTime(c int64) time.Duration {
	return time.Duration(c) * time.Second
}

func Expires(c int64) time.Time {
	return time.Now().Add(DurationTime(c))
}

func ExpiresAt(c int64) *jwt.NumericDate {
	return jwt.NewNumericDate(Expires(c))
}
