package core

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tkahng/authgo/internal/shared"
)

type TokenOption struct {
	Type   shared.TokenType `form:"type" json:"type" enum:"authentication_token,password_reset_token,verification_token,refresh_token,state_token"`
	Secret string           `form:"secret" json:"secret,omitempty"`
	// Duration specifies how long an issued token to be valid (in seconds)
	Duration int64 `form:"duration" json:"duration"`
}

func (c TokenOption) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Secret, validation.Required, validation.Length(10, 255)),
		validation.Field(&c.Duration, validation.Required, validation.Min(10), validation.Max(94670856)), // ~3y max
	)
}

func (o *TokenOption) SetClaims(c *jwt.RegisteredClaims) {
	c.ExpiresAt = o.ExpiresAt()
}

func (o *TokenOption) SetDto(c *shared.CreateTokenDTO) {
	c.Expires = o.Expires()
}

func (c *TokenOption) DurationTime() time.Duration {
	return time.Duration(c.Duration) * time.Second
}

func (c *TokenOption) ExpiresAt() *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(c.DurationTime()))
}

func (c *TokenOption) Expires() time.Time {
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
