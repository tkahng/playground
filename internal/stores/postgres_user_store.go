package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"

	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/crudrepo"
)

type PostgresUserStore struct {
	db database.Dbx
}

// AssignUserRoles implements UserStore.
func (a *PostgresUserStore) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	if len(roleNames) > 0 {
		user, err := crudrepo.User.GetOne(
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
		roles, err := crudrepo.Role.Get(
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
			_, err = crudrepo.UserRole.Post(ctx, a.db, userRoles)
			if err != nil {
				return fmt.Errorf("error assigning user role while assigning roles: %w", err)
			}
		}
	}
	return nil
}

// DeleteUser implements UserStore.
func (p *PostgresUserStore) DeleteUser(ctx context.Context, userId uuid.UUID) error {
	_, err := crudrepo.User.Delete(ctx, p.db, &map[string]any{
		"id": map[string]any{"_eq": userId.String()},
	})
	if err != nil {
		return err
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
-- user_sub_role_permissions AS (
--     SELECT u.id AS user_id,
--         p.name AS permission,
--         r.name AS role
--     FROM public.stripe_subscriptions s
--         JOIN public.users u ON s.user_id = u.id
--         JOIN public.stripe_prices price ON s.price_id = price.id
--         JOIN public.stripe_products product ON price.product_id = product.id
--         JOIN public.product_roles pr ON product.id = pr.product_id
--         JOIN public.roles r ON pr.role_id = r.id
--         JOIN public.role_permissions rp ON r.id = rp.role_id
--         JOIN public.permissions p ON rp.permission_id = p.id
-- ),
combined_permissions AS (
    SELECT *
    FROM user_role_permissions
    UNION ALL
    SELECT *
    FROM user_direct_permissions
    -- UNION ALL
    -- SELECT *
    -- FROM user_sub_role_permissions
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

// GetUserInfo implements UserStore.
func (p *PostgresUserStore) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	type rolePermissionClaims struct {
		UserID      uuid.UUID          `json:"user_id" db:"user_id"`
		Email       string             `json:"email" db:"email"`
		Roles       []string           `json:"roles" db:"roles"`
		Permissions []string           `json:"permissions" db:"permissions"`
		Providers   []models.Providers `json:"providers" db:"providers"`
	}
	user, err := crudrepo.User.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"email": map[string]any{
				"_eq": email,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, shared.ErrUserNotFound
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
	roles, err := func() (*rolePermissionClaims, error) {
		res, err := pgxscan.One(ctx, p.db, scan.StructMapper[rolePermissionClaims](), RawGetUserWithAllRolesAndPermissionsByEmail, email)
		if err != nil {
			return nil, err
		}
		return &res, nil
	}()
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

// UpdateUser implements UserStore.
func (p *PostgresUserStore) UpdateUser(ctx context.Context, user *models.User) error {
	_, err := crudrepo.User.PutOne(ctx, p.db, &models.User{
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

// var _ UserStore = (*PostgresUserStore)(nil)

func NewPostgresUserStore(db database.Dbx) *PostgresUserStore {
	return &PostgresUserStore{
		db: db,
	}
}

// FindUserByEmail implements UserStore.
func (p *PostgresUserStore) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	a, err := crudrepo.User.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"email": map[string]any{
				"_eq": email,
			},
		},
	)
	return database.OptionalRow(a, err)
}

// CreateUser implements UserStore.
func (p *PostgresUserStore) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	return crudrepo.User.PostOne(
		ctx,
		p.db,
		user,
	)
}
