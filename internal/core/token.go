package core

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/cookie"
)

// ----------- Authentication Claims -----------------

type AuthenticationClaims struct {
	jwt.RegisteredClaims
	Type shared.TokenType `json:"type"`
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
	Type shared.TokenType `json:"type"`
	RefreshTokenPayload
}

type RefreshTokenPayload struct {
	UserId uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Token  string    `json:"token"`
}

func CheckTokenType(claims jwt.MapClaims, tokenType shared.TokenType) bool {
	if claimType, ok := claims["type"].(string); ok && claimType == string(tokenType) {
		return true
	} else {
		return false
	}
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
	UserId     uuid.UUID        `json:"user_id,omitempty"`
	Email      string           `json:"email,omitempty"`
	Token      string           `json:"token"`
	Type       shared.TokenType `json:"type"`
	Otp        string           `json:"otp,omitempty"`
	RedirectTo string           `json:"redirect_to,omitempty"`
}

// ----------- Provider State Claims -----------------

type ProviderStateClaims struct {
	jwt.RegisteredClaims
	ProviderStatePayload
}

type ProviderStatePayload struct {
	// UserId              uuid.UUID        `json:"user_id,omitempty"`
	// Email               string           `json:"email,omitempty"`
	Token               string                `json:"token"`
	Type                shared.TokenType      `json:"type"`
	Provider            shared.OAuthProviders `json:"provider"`
	CodeVerifier        string                `json:"code_verifier,omitempty"`
	CodeChallenge       string                `json:"code_challenge,omitempty"`
	CodeChallengeMethod string                `json:"code_challenge_method,omitempty"`
	RedirectTo          string                `json:"redirect_to,omitempty"`
}

// ----------- Password Reset Claims -----------------

type PasswordResetClaims struct {
	jwt.RegisteredClaims
	OtpPayload
}

// ---------COOKIES------------

func SetTokenCookies(w http.ResponseWriter, tokens shared.TokenDto, config AuthOptions) {
	cookie.SetTokenCookie(w, cookie.AccessTokenCookieName, tokens.AccessToken, config.AccessToken.Expires())
	cookie.SetTokenCookie(w, cookie.RefreshTokenCookieName, tokens.RefreshToken, config.RefreshToken.Expires())
}
