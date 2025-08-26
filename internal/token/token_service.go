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
	GenerateToken(ctx context.Context, email string, emailType models.TokenTypes) (string, error)
	ValidateToken(ctx context.Context, token string, emailType models.TokenTypes) (string, error)
}

type TokenServiceImpl struct {
	opts  *conf.EnvConfig
	store stores.DbTokenStoreInterface
}

// ValidateToken implements TokenService.
func (t *TokenServiceImpl) ValidateToken(ctx context.Context, token string, emailType models.TokenTypes) (string, error) {
	dbtoken, err := t.store.GetTokenByValueTypeExpires(ctx, token, emailType, time.Now())
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

// GenerateToken implements TokenService.
func (t *TokenServiceImpl) GenerateToken(ctx context.Context, email string, emailType models.TokenTypes) (string, error) {
	token := security.GenerateTokenKey()
	var opt conf.TokenOption
	switch emailType {
	case models.TokenTypesVerificationToken:
		opt = t.opts.AuthOptions.VerificationToken
	case models.TokenTypesPasswordResetToken:
		opt = t.opts.AuthOptions.PasswordResetToken
	case models.TokenTypesRefreshToken:
		opt = t.opts.AuthOptions.RefreshToken
	default:
		return "", fmt.Errorf("invalid email type %v", emailType)
	}
	expiry := opt.Duration
	err := t.store.SaveToken(ctx, &stores.CreateTokenDTO{
		Type:       emailType,
		Token:      token,
		Identifier: email,
		Expires:    time.Now().Add(time.Duration(expiry) * time.Second),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

var _ TokenService = (*TokenServiceImpl)(nil)
