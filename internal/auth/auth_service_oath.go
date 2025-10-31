package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

type (
	OAuth2SigninInput struct {
		Email             string
		Name              *string
		AvatarUrl         *string
		EmailVerifiedAt   *time.Time
		Provider          models.Providers
		ProviderAccountID string
		UserId            *uuid.UUID
		AccessToken       *string
		RefreshToken      *string
	}
	OAuth2UrlInput struct {
		Provider models.Providers
	}
)

type Oauth2Authenticator interface {
	OAuth2Url(ctx context.Context, provider *OAuth2SigninInput) (string, error)
	// OAuth2Signin user.
	// the callback handlers will call this method
	//
	// - if user with email does not exist, it will create a new user and a oauth account.
	//
	// - if user with email exists, and they have another oauth account, it will update the oauth account.
	OAuth2Signin(ctx context.Context, params *OAuth2SigninInput) (*models.UserInfoTokens, error)
}

// OAuth2Url implements AuthService.
func (a *AuthServiceImpl) OAuth2Url(ctx context.Context, provider *OAuth2SigninInput) (string, error) {
	return "", nil
}

// OAuth2Signin implements AuthService.
func (a *AuthServiceImpl) OAuth2Signin(ctx context.Context, params *OAuth2SigninInput) (*models.UserInfoTokens, error) {
	return nil, errors.ErrUnsupported
}

type ProviderStateClaims struct {
	jwt.RegisteredClaims
	ProviderStatePayload
}

type ProviderStatePayload struct {
	Token               string            `json:"token"`
	Type                models.TokenTypes `json:"type"`
	Provider            models.Providers  `json:"provider"`
	CodeVerifier        string            `json:"code_verifier,omitempty"`
	CodeChallenge       string            `json:"code_challenge,omitempty"`
	CodeChallengeMethod string            `json:"code_challenge_method,omitempty"`
	RedirectTo          string            `json:"redirect_to,omitempty"`
}

// CreateAndPersistStateToken implements AuthActions.
func (a *AuthServiceImpl) CreateAndPersistStateToken(ctx context.Context, payload *ProviderStatePayload) (string, error) {
	if payload == nil {
		return "", fmt.Errorf("payload is nil")
	}
	config := a.config.StateToken
	claims := ProviderStateClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: config.ExpiresAt(),
		},
		ProviderStatePayload: *payload,
	}
	token, err := a.jwt.CreateJwtToken(claims, config.Secret)
	if err != nil {
		return token, err
	}

	err = a.adapter.Token().SaveToken(ctx, &stores.CreateTokenDTO{
		Type:       models.TokenTypesStateToken,
		Identifier: payload.Token,
		Expires:    config.Expires(),
		Token:      payload.Token,
	})
	if err != nil {
		return token, err
	}
	return token, nil
}
