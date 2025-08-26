package token

import (
	"context"
	"time"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/security"
)

type TokenService interface {
	GenerateEmailVerificationToken(ctx context.Context, email string) (string, error)
	VerifyEmailVerificationToken(ctx context.Context, token string) (string, error)
	GenerateResetPasswordToken(ctx context.Context, email string) (string, error)
	VerifyResetPasswordToken(ctx context.Context, token string) (string, error)
}

type TokenServiceImpl struct {
	opts  *conf.EnvConfig
	store stores.DbTokenStoreInterface
}

// GenerateEmailVerificationToken implements TokenService.
func (t *TokenServiceImpl) GenerateEmailVerificationToken(ctx context.Context, email string) (string, error) {
	token := security.GenerateTokenKey()
	expiry := t.opts.AuthOptions.VerificationToken.Duration
	err := t.store.SaveToken(ctx, &stores.CreateTokenDTO{
		Type:       models.TokenTypesVerificationToken,
		Token:      token,
		Identifier: email,
		Expires:    time.Now().Add(time.Duration(expiry) * time.Second),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

// GenerateResetPasswordToken implements TokenService.
func (t *TokenServiceImpl) GenerateResetPasswordToken(ctx context.Context, email string) (string, error) {
	token := security.GenerateTokenKey()
	expiry := t.opts.AuthOptions.PasswordResetToken.Duration
	err := t.store.SaveToken(ctx, &stores.CreateTokenDTO{
		Type:       models.TokenTypesPasswordResetToken,
		Token:      token,
		Identifier: email,
		Expires:    time.Now().Add(time.Duration(expiry) * time.Second),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

// VerifyEmailVerificationToken implements TokenService.
func (t *TokenServiceImpl) VerifyEmailVerificationToken(ctx context.Context, token string) (string, error) {
	dbtoken, err := t.store.GetToken(ctx, token)
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

// VerifyResetPasswordToken implements TokenService.
func (t *TokenServiceImpl) VerifyResetPasswordToken(ctx context.Context, token string) (string, error) {
	panic("unimplemented")
}

var _ TokenService = (*TokenServiceImpl)(nil)
