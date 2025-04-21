package core

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/cookie"
	"github.com/tkahng/authgo/internal/tools/security"
)

func VerifyAndParseToken(token string, config TokenOption, data any) error {
	claims, err := security.ParseJWTMapClaims(token, config.Secret)
	if err != nil {
		return fmt.Errorf("error while parsing token string: %w", err)
	}
	if !CheckTokenType(claims, config.Type) {
		return fmt.Errorf("invalid token type")
	}
	// Convert the JSON to a struct
	_, err = security.MarshalToken(claims, data)
	if err != nil {
		return fmt.Errorf("error at error: %w", err)
	}
	return nil
}

// Token Storage ----------------------------------------------------------------------------------------------

// TokenStorage persists and verifies tokens through queries
type TokenStorage struct {
	// authOpts *AuthOptions
}

// persist verification token to db
func (storage *TokenStorage) PersistVerificationToken(ctx context.Context, db bob.Executor, payload *OtpPayload, config TokenOption) error {
	return PersistOtpToken(ctx, db, payload, config)
}

// gets and expires give token, then updates user email confirmed when token is valid
func (storage *TokenStorage) UseVerificationTokenAndUpdateUser(ctx context.Context, db bob.Executor, token string) error {
	return UseVerificationTokenAndUpdateUser(ctx, db, token)
}

// Token Verifier ----------------------------------------------------------------------------------------------

// Token Verifier verifies and signs tokens
type TokenVerifier struct {
}

// create a new verification payload for use in creating verification token and email
func (t *TokenVerifier) CreateVerificationPayload(user *models.User, redirectTo string) *OtpPayload {
	otp := security.GenerateOtp(6)
	token := security.GenerateTokenKey()
	payload := &OtpPayload{
		UserId:     user.ID,
		Email:      user.Email,
		Type:       shared.VerificationTokenType,
		Token:      token,
		Otp:        otp,
		RedirectTo: redirectTo,
	}
	return payload
}
func (t *TokenVerifier) CreateResetPasswordPayload(user *models.User, redirectTo string) *OtpPayload {
	otp := security.GenerateOtp(6)
	token := security.GenerateTokenKey()
	payload := &OtpPayload{
		UserId:     user.ID,
		Email:      user.Email,
		Type:       shared.PasswordResetTokenType,
		Token:      token,
		Otp:        otp,
		RedirectTo: redirectTo,
	}
	return payload
}

// create a new verification token from claims
func (t *TokenVerifier) CreateVerificationToken(payload *OtpPayload, config TokenOption) (string, error) {
	return CreateOtpToken(payload, config)
}

// parse and verify token string to claims
func (t *TokenVerifier) ParseVerificationToken(token string, config TokenOption) (*EmailVerificationClaims, error) {
	return ParseVerificationToken(token, config)
}

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

func CreateAuthenticationToken(payload *AuthenticationPayload, config TokenOption) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	claims := AuthenticationClaims{
		Type: shared.AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		AuthenticationPayload: *payload,
	}
	token, err := security.NewJWTWithClaims(claims, config.Secret)
	if err != nil {
		return token, fmt.Errorf("error at error: %w", err)
	}

	return token, nil
}

func VerifyAuthenticationToken(token string, config TokenOption) (*AuthenticationClaims, error) {
	claims, err := security.ParseJWTMapClaims(token, config.Secret)
	if err != nil {
		return nil, fmt.Errorf("error while parsing token string: %w", err)
	}
	if !CheckTokenType(claims, shared.AccessTokenType) {
		return nil, fmt.Errorf("invalid token type")
	}
	var structData AuthenticationClaims
	// Convert the JSON to a struct
	jsond, err := security.MarshalToken(claims, &structData)
	if err != nil {
		return nil, fmt.Errorf("error at error: %w", err)
	}
	return jsond, nil
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

func CreateAndPersistRefreshToken(ctx context.Context, db bob.Executor, payload *RefreshTokenPayload, config TokenOption) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	claims := RefreshTokenClaims{
		Type: shared.RefreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		RefreshTokenPayload: *payload,
	}
	dto := &repository.TokenDTO{
		Type:       models.TokenTypesRefreshToken,
		Identifier: payload.Email,
		Expires:    config.Expires(),
		Token:      payload.Token,
		UserID:     &payload.UserId,
	}
	token, err := security.NewJWTWithClaims(claims, config.Secret)
	if err != nil {
		return token, fmt.Errorf("error at error: %w", err)
	}

	_, err = repository.CreateToken(ctx, db, dto)
	if err != nil {
		return token, fmt.Errorf("error at error: %w", err)
	}
	return token, nil
}

func CheckTokenType(claims jwt.MapClaims, tokenType shared.TokenType) bool {
	if claimType, ok := claims["type"].(string); ok && claimType == string(tokenType) {
		return true
	} else {
		return false
	}
}

func VerifyRefreshToken(ctx context.Context, db bob.Executor, token string, config TokenOption) (*RefreshTokenClaims, error) {
	// parse token
	claims, err := security.ParseJWTMapClaims(token, config.Secret)
	if err != nil {
		return nil, fmt.Errorf("error while parsing token string: %w", err)
	}
	if !CheckTokenType(claims, shared.RefreshTokenType) {
		return nil, fmt.Errorf("invalid token %v", claims)
	}
	var structData RefreshTokenClaims
	// Convert the JSON to a struct
	jsond, err := security.MarshalToken(claims, &structData)
	if err != nil {
		return nil, fmt.Errorf("error at error: %w", err)
	}
	// get and expire token
	dbToken, err := repository.UseToken(ctx, db, jsond.Token)
	if err != nil {
		return nil, fmt.Errorf("error at error: %w", err)
	}

	if dbToken == nil {
		return nil, fmt.Errorf("token not found")
	}
	return jsond, nil
}

// ----------- Email Verification Claims -----------------

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

// create a new verification token from claims
func CreateOtpToken(payload *OtpPayload, config TokenOption) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	claims := EmailVerificationClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		OtpPayload: *payload,
	}
	token, err := security.NewJWTWithClaims(claims, config.Secret)
	if err != nil {
		return "", fmt.Errorf("error at creating verification token: %w", err)
	}
	return token, nil
}

// persist verification token to db
func PersistOtpToken(ctx context.Context, db bob.Executor, payload *OtpPayload, config TokenOption) error {

	//  clear any existing verification tokens
	_ = repository.DeleteTokensByUser(ctx, db, &repository.OtpDto{
		Type:       models.TokenTypes(payload.Type),
		Identifier: payload.Email,
		Otp:        &payload.Otp,
		UserID:     &payload.UserId,
	})
	// save new verification token
	dto := &repository.TokenDTO{
		Type:       models.TokenTypes(payload.Type),
		Identifier: payload.Email,
		Expires:    config.Expires(),
		Token:      payload.Token,
		Otp:        &payload.Otp,
		UserID:     &payload.UserId,
	}
	_, err := repository.CreateToken(ctx, db, dto)
	if err != nil {
		return fmt.Errorf("error at storing verification token: %w", err)
	}
	return nil
}

// parse and verify token string to claims
func ParseVerificationToken(token string, config TokenOption) (*EmailVerificationClaims, error) {
	claims, err := security.ParseJWTMapClaims(token, config.Secret)
	if err != nil {
		return nil, fmt.Errorf("error while parsing token string: %w", err)
	}
	if !CheckTokenType(claims, shared.VerificationTokenType) {
		return nil, fmt.Errorf("invalid token. want verification, got %v", claims)
	}
	var structData EmailVerificationClaims
	// Convert the JSON to a struct
	jsond, err := security.MarshalToken(claims, &structData)
	if err != nil {
		return nil, fmt.Errorf("error at marshaling token: %w", err)
	}
	return jsond, nil
}

// gets and expires give token, then updates user email confirmed when token is valid
func UseVerificationTokenAndUpdateUser(ctx context.Context, db bob.Executor, token string) error {
	//  get and delete token
	dbToken, err := repository.UseToken(ctx, db, token)
	if err != nil {
		return fmt.Errorf("error at use token: %w", err)
	}
	if dbToken == nil {
		return fmt.Errorf("token not found")
	}
	if dbToken.Type != models.TokenTypesVerificationToken {
		return fmt.Errorf("invalid token type. want verification_token, got  %v", dbToken.Type)
	}
	if dbToken.UserID.IsNull() {
		return fmt.Errorf("found a verification token, but it cannot have userId nil")
	}
	//  if token is valid, update user email confirmed
	_, err = repository.UpdateUserEmailConfirm(ctx, db, dbToken.UserID.MustGet(), time.Now())
	if err != nil {
		return fmt.Errorf("eror updating user: %w", err)
	}
	return nil
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

func CreateAndPersistStateToken(ctx context.Context, db bob.Executor, payload *ProviderStatePayload, config TokenOption) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	claims := ProviderStateClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		ProviderStatePayload: *payload,
	}
	token, err := security.NewJWTWithClaims(claims, config.Secret)
	if err != nil {
		return "", fmt.Errorf("error at creating verification token: %w", err)
	}
	dto := &repository.TokenDTO{
		Type:       models.TokenTypes(payload.Type),
		Identifier: string(payload.Type),
		Expires:    config.Expires(),
		Token:      payload.Token,
	}
	_, err = repository.CreateToken(ctx, db, dto)
	if err != nil {
		return "", fmt.Errorf("error at storing verification token: %w", err)
	}
	return token, nil
}

// parse and verify token string to claims
func ParseProviderStateToken(token string, config TokenOption) (*ProviderStateClaims, error) {
	claims, err := security.ParseJWTMapClaims(token, config.Secret)
	if err != nil {
		return nil, fmt.Errorf("error while parsing token string: %w", err)
	}
	if !CheckTokenType(claims, shared.StateTokenType) {
		return nil, fmt.Errorf("invalid token. want provider_state, got %v", claims)
	}
	var structData ProviderStateClaims
	// Convert the JSON to a struct
	jsond, err := security.MarshalToken(claims, &structData)
	if err != nil {
		return nil, fmt.Errorf("error at marshaling token: %w", err)
	}
	return jsond, nil
}

// ----------- Password Reset Claims -----------------

type PasswordResetClaims struct {
	jwt.RegisteredClaims
	OtpPayload
}

// create a new verification token from claims
func CreatePasswordResetToken(payload *OtpPayload, config TokenOption) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	claims := PasswordResetClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		OtpPayload: *payload,
	}
	token, err := security.NewJWTWithClaims(claims, config.Secret)
	if err != nil {
		return "", fmt.Errorf("error at creating verification token: %w", err)
	}
	return token, nil
}

// parse and verify token string to claims
func ParseResetToken(token string, config TokenOption) (*PasswordResetClaims, error) {
	claims, err := security.ParseJWTMapClaims(token, config.Secret)
	if err != nil {
		return nil, fmt.Errorf("error while parsing token string: %w", err)
	}
	if !CheckTokenType(claims, shared.PasswordResetTokenType) {
		return nil, fmt.Errorf("invalid token. want password_reset, got %v", claims)
	}
	var structData PasswordResetClaims
	// Convert the JSON to a struct
	jsond, err := security.MarshalToken(claims, &structData)
	if err != nil {
		return nil, fmt.Errorf("error at marshaling token: %w", err)
	}
	return jsond, nil
}

// ---------COOKIES------------

func SetTokenCookies(w http.ResponseWriter, tokens shared.TokenDto, config AuthOptions) {
	cookie.SetTokenCookie(w, cookie.AccessTokenCookieName, tokens.AccessToken, config.AccessToken.Expires())
	cookie.SetTokenCookie(w, cookie.RefreshTokenCookieName, tokens.RefreshToken, config.RefreshToken.Expires())
}
