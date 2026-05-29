package token

import (
	"context"
	"fmt"
	"time"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/security"
)

type TokenService interface {
	GenerateToken(ctx context.Context, options *GenerateTokenOptions) (string, error)
	// ValidateToken validates the token, deletes it, and returns the associated email.
	// Returns an error if the token is missing, expired, or of the wrong type.
	ValidateToken(ctx context.Context, options *ValidateTokenOptions) (string, error)
	CheckToken(ctx context.Context, token string, tokenType models.TokenTypes) error
	// RevokeToken deletes a token by its value, used for signout.
	RevokeToken(ctx context.Context, tokenValue string) error
}

type TokenServiceImpl struct {
	opts  *conf.EnvConfig
	store stores.DbTokenStoreInterface
}

// CheckToken implements TokenService.
func (t *TokenServiceImpl) CheckToken(ctx context.Context, token string, tokenType models.TokenTypes) error {
	dbtoken, err := t.store.GetTokenByValueTypeExpires(ctx, token, tokenType, time.Now())
	if err != nil {
		return err
	}
	if dbtoken == nil {
		return fmt.Errorf("token not found")
	}
	return nil
}

type ValidateTokenOptions struct {
	Email string
	Otp   string
	Token string
	Type  models.TokenTypes
}

// ValidateToken uses the token, tokenType and the current time to validate the token.
// If the token is valid, it delete the token, and return the email asscoiated with the token.
// If the token is not valid, it will return an error.
func (t *TokenServiceImpl) ValidateToken(ctx context.Context, options *ValidateTokenOptions) (string, error) {
	var tokenToValidate string
	if options.Otp != "" && options.Email != "" {
		tokenToValidate = security.GenerateTokenHash(options.Email, options.Otp)
	} else {
		tokenToValidate = options.Token
	}
	if tokenToValidate == "" {
		return "", fmt.Errorf("token to validate is empty")
	}
	dbtoken, err := t.store.GetTokenByValueTypeExpires(ctx, tokenToValidate, options.Type, time.Now())
	if err != nil {
		return "", err
	}
	if dbtoken == nil {
		return "", fmt.Errorf("token not found or expired")
	}
	err = t.store.DeleteToken(ctx, tokenToValidate)
	if err != nil {
		return "", err
	}
	return dbtoken.Identifier, nil
}

type GenerateTokenOptions struct {
	Email string
	Otp   string
	Type  models.TokenTypes
}

func (t *TokenServiceImpl) GenerateToken(ctx context.Context, options *GenerateTokenOptions) (string, error) {
	var token string

	if options.Otp != "" {
		token = security.GenerateTokenHash(options.Email, options.Otp)
	} else {
		token = security.GenerateTokenKey()
	}
	var opt conf.TokenOption
	switch options.Type {
	case models.TokenTypesVerificationToken:
		opt = t.opts.VerificationToken
	case models.TokenTypesPasswordResetToken:
		opt = t.opts.PasswordResetToken
	case models.TokenTypesRefreshToken:
		opt = t.opts.RefreshToken
	default:
		return "", fmt.Errorf("invalid email type %v", options.Type)
	}
	expiry := opt.Duration
	err := t.store.SaveToken(ctx, &stores.CreateTokenDTO{
		Type:       options.Type,
		Token:      token,
		Identifier: options.Email,
		Expires:    time.Now().UTC().Add(time.Duration(expiry) * time.Second),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

// RevokeToken deletes a token by its raw value, used during signout.
func (t *TokenServiceImpl) RevokeToken(ctx context.Context, tokenValue string) error {
	return t.store.DeleteToken(ctx, tokenValue)
}

func NewTokenService(opts *conf.EnvConfig, store stores.DbTokenStoreInterface) TokenService {
	return &TokenServiceImpl{
		opts:  opts,
		store: store,
	}
}

var _ TokenService = (*TokenServiceImpl)(nil)
