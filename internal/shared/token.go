package shared

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

// ----------- Authentication Claims -----------------

type AuthenticationClaims struct {
	jwt.RegisteredClaims
	Type models.TokenTypes `json:"type"`
	AuthenticationPayload
}

type AuthenticationPayload struct {
	UserId      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
}

// ----------- Refresh Token Claims -----------------

type RefreshTokenClaims struct {
	jwt.RegisteredClaims
	Type models.TokenTypes `json:"type"`
	RefreshTokenPayload
}

type RefreshTokenPayload struct {
	UserId uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Token  string    `json:"token"`
}

// ----------- Email Verification Claims -----------------

type OtpClaims struct {
	jwt.RegisteredClaims
	OtpPayload
}

type EmailVerificationClaims struct {
	jwt.RegisteredClaims
	OtpPayload
}

type OtpPayload struct {
	UserId     uuid.UUID         `json:"user_id,omitempty"`
	Email      string            `json:"email,omitempty"`
	Token      string            `json:"token"`
	Type       models.TokenTypes `json:"type"`
	Otp        string            `json:"otp,omitempty"`
	RedirectTo string            `json:"redirect_to,omitempty"`
}

// ----------- Provider State Claims -----------------

type ProviderStateClaims struct {
	jwt.RegisteredClaims
	ProviderStatePayload
}

type ProviderStatePayload struct {
	// UserId              uuid.UUID        `json:"user_id,omitempty"`
	// Email               string           `json:"email,omitempty"`
	Token               string            `json:"token"`
	Type                models.TokenTypes `json:"type"`
	Provider            Providers         `json:"provider"`
	CodeVerifier        string            `json:"code_verifier,omitempty"`
	CodeChallenge       string            `json:"code_challenge,omitempty"`
	CodeChallengeMethod string            `json:"code_challenge_method,omitempty"`
	RedirectTo          string            `json:"redirect_to,omitempty"`
}

// ----------- Password Reset Claims -----------------

type PasswordResetClaims struct {
	jwt.RegisteredClaims
	OtpPayload
}
