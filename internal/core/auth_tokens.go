package core

import (
	"context"
	"fmt"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/shared"
)

// HandleAuthToken implements App.
func (a *BaseApp) HandleAuthToken(ctx context.Context, token string) (*shared.UserInfoDto, error) {
	db := a.Db()
	opts := a.Settings().Auth
	claims, err := VerifyAuthenticationToken(token, opts.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying refresh token: %w", err)
	}
	user, err := GetUserInfoDTO(ctx, db, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}
	return user, nil
}

func (a *BaseApp) CreateAuthDto(ctx context.Context, email string) (*shared.AuthenticatedDTO, error) {
	db := a.Db()
	info, err := GetUserInfoDTO(ctx, db, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}
	if info == nil {
		return nil, fmt.Errorf("user not found")
	}
	tokens, err := a.CreateAuthTokens(ctx, db, info)
	if err != nil || tokens == nil {
		return nil, fmt.Errorf("error creating auth tokens: %w", err)
	}
	bod := shared.AuthenticatedDTO{
		User:        shared.ToUser(&info.User),
		Tokens:      *tokens,
		Roles:       info.Roles,
		Permissions: info.Permissions,
		Providers:   info.Providers,
	}
	return &bod, nil
}

func (a *BaseApp) RefreshTokens(ctx context.Context, db bob.Executor, refreshToken string) (*shared.AuthenticatedDTO, error) {
	opts := a.Settings().Auth
	claims, err := VerifyRefreshToken(ctx, db, refreshToken, opts.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying refresh token: %w", err)
	}
	info, err := GetUserInfoDTO(ctx, db, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("error getting user info: %w", err)
	}

	tokens, err := a.CreateAuthTokens(ctx, db, info)
	if err != nil || tokens == nil {
		return nil, fmt.Errorf("error creating auth tokens: %w", err)
	}
	return &shared.AuthenticatedDTO{
		User:        shared.ToUser(&info.User),
		Tokens:      *tokens,
		Roles:       info.Roles,
		Permissions: info.Permissions,
		Providers:   info.Providers,
	}, nil
}

func (a *BaseApp) Signout(ctx context.Context, db bob.Executor, refreshToken string) error {
	opts := a.Settings().Auth
	_, err := VerifyRefreshToken(ctx, db, refreshToken, opts.RefreshToken)
	if err != nil {
		return fmt.Errorf("error verifying refresh token: %w", err)
	}

	return nil
}
