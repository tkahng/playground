package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
)

var _ AuthStorage = (*AuthAdapterBase)(nil)

func NewAuthStorage(dbtx db.Dbx) *AuthAdapterBase {
	return &AuthAdapterBase{db: dbtx}
}

type AuthAdapterBase struct {
	db db.Dbx
}

func (a *AuthAdapterBase) GetToken(ctx context.Context, token string) (*shared.Token, error) {
	res, err := repository.Token.GetOne(ctx,
		a.db,
		&map[string]any{
			"token": map[string]any{
				"_eq": token,
			},
			"expires": map[string]any{
				"_gt": time.Now(),
			},
		})
	if err != nil {
		return nil, fmt.Errorf("error at getting token: %w", err)
	}
	return &shared.Token{
		Type:       shared.TokenType(res.Type),
		Identifier: res.Identifier,
		Expires:    res.Expires,
		Token:      res.Token,
		ID:         res.ID,
		UserID:     res.UserID,
		Otp:        res.Otp,
	}, nil
}

func (a *AuthAdapterBase) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	_, err := repository.Token.PostOne(ctx, a.db, &models.Token{
		Type:       models.TokenTypes(token.Type),
		Identifier: token.Identifier,
		Expires:    token.Expires,
		Token:      token.Token,
		UserID:     token.UserID,
		Otp:        token.Otp,
	})

	if err != nil {
		return fmt.Errorf("error at saving token: %w", err)
	}
	return nil
}

func (a *AuthAdapterBase) DeleteToken(ctx context.Context, token string) error {
	_, err := repository.Token.DeleteReturn(ctx, a.db, &map[string]any{
		"token": map[string]any{
			"_eq": token,
		},
	})
	if err != nil {
		return fmt.Errorf("error at deleting token: %w", err)
	}
	return nil
}

func (a *AuthAdapterBase) VerifyTokenStorage(ctx context.Context, token string) error {
	res, err := a.GetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("error at getting token: %w", err)
	}
	if res == nil {
		return fmt.Errorf("token not found")
	}
	err = a.DeleteToken(ctx, token)
	if err != nil {
		return fmt.Errorf("error at deleting token: %w", err)
	}
	return nil
}

// FindUserByEmail implements AuthAdapter.
func (a *AuthAdapterBase) FindUserByEmail(ctx context.Context, email string) (*shared.User, error) {

	user, err := queries.FindUserByEmail(ctx, a.db, email)

	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return &shared.User{
		ID:              user.ID,
		Email:           user.Email,
		EmailVerifiedAt: user.EmailVerifiedAt,
		Name:            user.Name,
		Image:           user.Image,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}, nil
}

// FindUserAccountByUserIdAndProvider implements AuthAdapter.
func (a *AuthAdapterBase) FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider shared.Providers) (*shared.UserAccount, error) {

	account, err := queries.FindUserAccountByUserIdAndProvider(ctx, a.db, userId, provider)

	if err != nil {
		return nil, fmt.Errorf("error getting user account: %w", err)
	}

	return &shared.UserAccount{
		ID:                account.ID,
		UserID:            account.UserID,
		Provider:          shared.Providers(account.Provider),
		ProviderAccountID: account.ProviderAccountID,
		CreatedAt:         account.CreatedAt,
		UpdatedAt:         account.UpdatedAt,
		Type:              shared.ProviderTypes(account.Type),
		AccessToken:       account.AccessToken,
		RefreshToken:      account.RefreshToken,
		ExpiresAt:         account.ExpiresAt,
		IDToken:           account.IDToken,
		Scope:             account.Scope,
		SessionState:      account.SessionState,
		TokenType:         account.TokenType,
		Password:          account.Password,
	}, nil
}

// UpdateUserAccount implements AuthAdapter.
func (a *AuthAdapterBase) UpdateUserAccount(ctx context.Context, account *shared.UserAccount) error {
	res, err := repository.UserAccount.PutOne(ctx, a.db, &models.UserAccount{
		ID:                account.ID,
		UserID:            account.UserID,
		Provider:          models.Providers(account.Provider),
		ProviderAccountID: account.ProviderAccountID,
		CreatedAt:         account.CreatedAt,
		UpdatedAt:         account.UpdatedAt,
		Type:              models.ProviderTypes(account.Type),
		AccessToken:       account.AccessToken,
		RefreshToken:      account.RefreshToken,
		ExpiresAt:         account.ExpiresAt,
		IDToken:           account.IDToken,
		Scope:             account.Scope,
		SessionState:      account.SessionState,
		TokenType:         account.TokenType,
		Password:          account.Password,
	})
	if err != nil {
		return fmt.Errorf("error updating user account: %w", err)
	}
	if res == nil {
		return fmt.Errorf("user account not found")
	}
	return nil
}

// GetUserInfo implements AuthAdapter.
func (a *AuthAdapterBase) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	user, err := repository.User.GetOne(ctx, a.db, &map[string]any{"email": map[string]any{"_eq": email}})
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	result := &shared.UserInfo{
		User: shared.User{
			ID:              user.ID,
			Email:           user.Email,
			EmailVerifiedAt: user.EmailVerifiedAt,
			Name:            user.Name,
			Image:           user.Image,
			CreatedAt:       user.CreatedAt,
			UpdatedAt:       user.UpdatedAt,
		},
	}
	roles, err := queries.FindUserWithRolesAndPermissionsByEmail(ctx, a.db, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user roles and permissions: %w", err)
	}
	var providers []shared.Providers
	for _, provider := range roles.Providers {
		providers = append(providers, shared.Providers(provider))
	}
	result.Roles = roles.Roles
	result.Permissions = roles.Permissions
	result.Providers = providers

	return result, nil
}

// CreateUser implements AuthAdapter.
func (a *AuthAdapterBase) CreateUser(ctx context.Context, user *shared.User) (*shared.User, error) {
	res, err := repository.User.PostOne(ctx, a.db, &models.User{
		Email:           user.Email,
		Name:            user.Name,
		Image:           user.Image,
		EmailVerifiedAt: user.EmailVerifiedAt,
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("user not found")
	}
	return &shared.User{
		ID:              res.ID,
		Email:           res.Email,
		EmailVerifiedAt: res.EmailVerifiedAt,
		Name:            res.Name,
		Image:           res.Image,
		CreatedAt:       res.CreatedAt,
		UpdatedAt:       res.UpdatedAt,
	}, nil
}

// DeleteUser implements AuthAdapter.
func (a *AuthAdapterBase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	res, err := repository.User.DeleteReturn(ctx, a.db, &map[string]any{
		"id": map[string]any{"_eq": id.String()},
	})
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("user not found")
	}
	return nil
}
func (a *AuthAdapterBase) LinkAccount(ctx context.Context, account *shared.UserAccount) error {
	if account == nil {
		return errors.New("account is nil")
	}
	_, err := repository.UserAccount.PostOne(ctx,
		a.db,
		&models.UserAccount{
			ID:                account.ID,
			UserID:            account.UserID,
			Provider:          models.Providers(account.Provider),
			ProviderAccountID: account.ProviderAccountID,
			CreatedAt:         account.CreatedAt,
			UpdatedAt:         account.UpdatedAt,
			Type:              models.ProviderTypes(account.Type),
			AccessToken:       account.AccessToken,
			RefreshToken:      account.RefreshToken,
			ExpiresAt:         account.ExpiresAt,
			IDToken:           account.IDToken,
			Scope:             account.Scope,
			SessionState:      account.SessionState,
			TokenType:         account.TokenType,
			Password:          account.Password,
		})
	if err != nil {
		return err
	}
	return nil
}

// UnlinkAccount implements AuthAdapter.
func (a *AuthAdapterBase) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider shared.Providers) error {
	// providerModel := shared.ToModelProvider(provider)
	// _, err := repository.DeleteAccount(ctx, a.db, userId, providerModel)
	// if err != nil {
	// 	return err
	// }
	return nil
}

// UpdateUser implements AuthAdapter.
func (a *AuthAdapterBase) UpdateUser(ctx context.Context, user *shared.User) error {
	_, err := repository.User.PutOne(ctx, a.db, &models.User{
		ID:              user.ID,
		Email:           user.Email,
		Name:            user.Name,
		Image:           user.Image,
		EmailVerifiedAt: user.EmailVerifiedAt,
		UpdatedAt:       time.Now(),
		CreatedAt:       user.CreatedAt,
	})
	if err != nil {
		return err
	}
	return nil
}

// AssignUserRoles implements AuthAdapter.
func (a *AuthAdapterBase) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	if len(roleNames) > 0 {
		user, err := repository.User.GetOne(
			ctx,
			a.db,
			&map[string]any{
				"id": map[string]any{
					"_eq": userId.String(),
				},
			},
		)
		if err != nil {
			return fmt.Errorf("error finding user while assigning roles: %w", err)
		}
		if user == nil {
			return fmt.Errorf("user not found while assigning roles")
		}
		roles, err := repository.Role.Get(
			ctx,
			a.db,
			&map[string]any{
				"name": map[string]any{
					"_in": roleNames,
				},
			},
			nil,
			types.Pointer(10),
			nil,
		)
		if err != nil {
			return fmt.Errorf("error finding user role while assigning roles: %w", err)
		}
		if len(roles) > 0 {
			var userRoles []models.UserRole
			for _, role := range roles {
				userRoles = append(userRoles, models.UserRole{
					UserID: user.ID,
					RoleID: role.ID,
				})
			}
			_, err = repository.UserRole.Post(ctx, a.db, userRoles)
			if err != nil {
				return fmt.Errorf("error assigning user role while assigning roles: %w", err)
			}
		}
	}
	return nil
}
