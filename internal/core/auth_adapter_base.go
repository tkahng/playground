package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
)

type AuthAdapter interface {
	Db() bob.Executor
	GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error)
	CreateUser(ctx context.Context, user *shared.User) (*shared.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*shared.User, error)
	GetUserByEmail(ctx context.Context, email string) (*shared.User, error)
	GetUserAccount(ctx context.Context, userId uuid.UUID, provider shared.Providers) (*shared.UserAccount, error)
	UpdateUser(ctx context.Context, user *shared.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	LinkAccount(ctx context.Context, account *shared.UserAccount) error
	UnlinkAccount(ctx context.Context, userId uuid.UUID, provider shared.Providers) error
}

var _ AuthAdapter = (*AuthAdapterBase)(nil)

type AuthAdapterBase struct {
	db bob.Executor
}

func (a *AuthAdapterBase) Db() bob.Executor {
	return a.db
}

// GetUserInfo implements AuthAdapter.
func (a *AuthAdapterBase) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	db := a.Db()
	user, err := a.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	result := &shared.UserInfo{
		User: *user,
	}
	roles, err := repository.FindUserWithRolesAndPermissionsByEmail(ctx, db, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user roles and permissions: %w", err)
	}
	if roles == nil {
		return result, nil
	}
	providers := shared.ToProvidersArray(roles.Providers)
	return &shared.UserInfo{
		User:        *user,
		Roles:       roles.Roles,
		Permissions: roles.Permissions,
		Providers:   providers,
	}, nil
}

// CreateUser implements AuthAdapter.
func (a *AuthAdapterBase) CreateUser(ctx context.Context, user *shared.User) (*shared.User, error) {
	res, err := repository.CreateUser(ctx, a.db, &shared.AuthenticateUserParams{
		Email:           user.Email,
		Name:            user.Name,
		AvatarUrl:       user.Image,
		EmailVerifiedAt: user.EmailVerifiedAt,
	})
	if err != nil {
		return nil, err
	}
	newUser := shared.ToUser(res)
	return newUser, nil
}

// DeleteUser implements AuthAdapter.
func (a *AuthAdapterBase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// GetUser implements AuthAdapter.
func (a *AuthAdapterBase) GetUser(ctx context.Context, id uuid.UUID) (*shared.User, error) {
	res, err := repository.FindUserById(ctx, a.db, id)
	if err != nil {
		return nil, err
	}
	return shared.ToUser(res), nil
}

// GetUserByAccount implements AuthAdapter.
func (a *AuthAdapterBase) GetUserAccount(ctx context.Context, userId uuid.UUID, provider shared.Providers) (*shared.UserAccount, error) {
	providerModel := shared.ToModelProvider(provider)
	res, err := repository.FindUserAccountByUserIdAndProvider(ctx, a.db, userId, providerModel)
	if err != nil {
		return nil, err
	}
	return shared.ToUserAccount(res), nil
}

// GetUserByEmail implements AuthAdapter.
func (a *AuthAdapterBase) GetUserByEmail(ctx context.Context, email string) (*shared.User, error) {
	res, err := repository.FindUserByEmail(ctx, a.db, email)
	if err != nil {
		return nil, err
	}
	return shared.ToUser(res), nil
}

// LinkAccount implements AuthAdapter.
func (a *AuthAdapterBase) LinkAccount(ctx context.Context, account *shared.UserAccount) error {
	if account == nil {
		return errors.New("account is nil")
	}
	providerModel := shared.ToModelProvider(account.Provider)
	providerTypeModel := shared.ToModelProviderType(account.Type)
	_, err := repository.CreateAccount(ctx, a.db, account.UserID, &shared.AuthenticateUserParams{
		UserId:            &account.UserID,
		Type:              providerTypeModel,
		Provider:          providerModel,
		ProviderAccountID: account.ProviderAccountID,
		Password:          account.Password,
		AccessToken:       account.AccessToken,
		RefreshToken:      account.RefreshToken,
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
	err := repository.UpdateUser(ctx, a.db, user.ID, &repository.UpdateUserInput{
		Email:           user.Email,
		Name:            user.Name,
		AvatarUrl:       user.Image,
		EmailVerifiedAt: user.EmailVerifiedAt,
	})
	if err != nil {
		return err
	}
	return nil
}
