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
	// ValidateToken uses the token, tokenType and the current time to validate the token.
	// If the token is valid, it delete the token, and return the email asscoiated with the token.
	// If the token is not valid, it will return an error.
	ValidateToken(ctx context.Context, token string, tokenType models.TokenTypes) (string, error)
	CheckToken(ctx context.Context, token string, tokenType models.TokenTypes) error
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

// ValidateToken uses the token, tokenType and the current time to validate the token.
// If the token is valid, it delete the token, and return the email asscoiated with the token.
// If the token is not valid, it will return an error.
func (t *TokenServiceImpl) ValidateToken(ctx context.Context, token string, tokenType models.TokenTypes) (string, error) {
	dbtoken, err := t.store.GetTokenByValueTypeExpires(ctx, token, tokenType, time.Now())
	if err != nil {
		return "", err
	}
	if dbtoken == nil {
		return "", nil
	}
	err = t.store.DeleteToken(ctx, token)
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

func NewTokenService(opts *conf.EnvConfig, store stores.DbTokenStoreInterface) TokenService {
	return &TokenServiceImpl{
		opts:  opts,
		store: store,
	}
}

var _ TokenService = (*TokenServiceImpl)(nil)
