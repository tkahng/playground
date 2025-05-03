package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	crud "github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/types"
)

type AuthAdapter interface {
	GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error)
	CreateUser(ctx context.Context, user *shared.User) (*shared.User, error)
	AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	FindUser(ctx context.Context, where *map[string]any) (*shared.User, error)
	FindUserAccount(ctx context.Context, where *map[string]any) (*shared.UserAccount, error)
	UpdateUser(ctx context.Context, user *shared.User) error
	UpdateUserAccount(ctx context.Context, account *shared.UserAccount) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	LinkAccount(ctx context.Context, account *shared.UserAccount) error
	UnlinkAccount(ctx context.Context, userId uuid.UUID, provider shared.Providers) error
}

var _ AuthAdapter = (*AuthAdapterBase)(nil)

func NewAuthAdapter(dbtx *pgxpool.Pool) *AuthAdapterBase {
	return &AuthAdapterBase{db: dbtx}
}

type AuthAdapterBase struct {
	db *pgxpool.Pool
}

// FindUser implements AuthAdapter.
func (a *AuthAdapterBase) FindUser(ctx context.Context, where *map[string]any) (*shared.User, error) {

	user, err := repository.User.GetOne(ctx, a.db, where)

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

// FindUserAccount implements AuthAdapter.
func (a *AuthAdapterBase) FindUserAccount(ctx context.Context, where *map[string]any) (*shared.UserAccount, error) {

	account, err := repository.UserAccount.GetOne(ctx, a.db, where)

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

const (
	RawGetUserWithAllRolesAndPermissionsByEmail string = `--sql
WITH -- Get permissions assigned through roles
user_role_permissions AS (
    SELECT ur.user_id AS user_id,
        p.name AS permission,
        r.name AS role
    FROM public.user_roles ur
        JOIN public.roles r ON ur.role_id = r.id
        JOIN public.role_permissions rp ON ur.role_id = rp.role_id
        JOIN public.permissions p ON rp.permission_id = p.id
),
user_direct_permissions AS (
    SELECT up.user_id AS user_id,
        p.name AS permission,
        NULL::text AS role
    FROM public.user_permissions up
        JOIN public.permissions p ON up.permission_id = p.id
),
user_sub_role_permissions AS (
    SELECT u.id AS user_id,
        p.name AS permission,
        r.name AS role
    FROM public.stripe_subscriptions s
        JOIN public.users u ON s.user_id = u.id
        JOIN public.stripe_prices price ON s.price_id = price.id
        JOIN public.stripe_products product ON price.product_id = product.id
        JOIN public.product_roles pr ON product.id = pr.product_id
        JOIN public.roles r ON pr.role_id = r.id
        JOIN public.role_permissions rp ON r.id = rp.role_id
        JOIN public.permissions p ON rp.permission_id = p.id
),
combined_permissions AS (
    SELECT *
    FROM user_role_permissions
    UNION ALL
    SELECT *
    FROM user_direct_permissions
    UNION ALL
    SELECT *
    FROM user_sub_role_permissions
)
SELECT u.id AS user_id,
    u.email AS email,
    array_remove(ARRAY_AGG(DISTINCT p.role), NULL)::text [] AS roles,
    array_remove(ARRAY_AGG(DISTINCT p.permission), NULL)::text [] AS permissions,
    array_remove(ARRAY_AGG(DISTINCT ua.provider), NULL)::public.providers [] AS providers
FROM public.users u
    LEFT JOIN combined_permissions p ON u.id = p.user_id
    LEFT JOIN public.user_accounts ua ON u.id = ua.user_id
WHERE u.email = $1
GROUP BY u.id
LIMIT 1;
`
)

type RolePermissionClaims struct {
	UserID      uuid.UUID          `json:"user_id" db:"user_id"`
	Email       string             `json:"email" db:"email"`
	Roles       []string           `json:"roles" db:"roles"`
	Permissions []string           `json:"permissions" db:"permissions"`
	Providers   []models.Providers `json:"providers" db:"providers"`
}

func FindUserWithRolesAndPermissionsByEmail(ctx context.Context, db crud.DBTX, email string) (*RolePermissionClaims, error) {
	res, err := pgxscan.One(ctx, db, scan.StructMapper[RolePermissionClaims](), RawGetUserWithAllRolesAndPermissionsByEmail, email)
	if err != nil {
		return nil, err
	}

	return &res, nil
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
	roles, err := pgxscan.One(ctx, a.db, scan.StructMapper[RolePermissionClaims](), RawGetUserWithAllRolesAndPermissionsByEmail, email)
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
