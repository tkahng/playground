package stores

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/security"
	"github.com/tkahng/authgo/internal/tools/types"

	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/crudrepo"
)

type DbUserStore struct {
	db database.Dbx
}

func NewDbUserStore(db database.Dbx) *DbUserStore {
	return &DbUserStore{
		db: db,
	}
}
func (s *DbUserStore) WithTx(tx database.Dbx) *DbUserStore {
	return &DbUserStore{
		db: tx,
	}
}

func (*DbUserStore) UserWhere(user *models.User) *map[string]any {
	if user == nil {
		return nil
	}
	where := map[string]any{}
	if user.ID != uuid.Nil {
		where[models.UserTable.ID] = map[string]any{
			"_eq": user.ID,
		}
	}
	if user.Name != nil {
		where[models.UserTable.Name] = map[string]any{
			"_like": fmt.Sprintf("%%%s%%", *user.Name),
		}
	}
	if user.Email != "" {
		where[models.UserTable.Email] = map[string]any{
			"_eq": user.Email,
		}
	}
	if user.EmailVerifiedAt != nil {
		if user.EmailVerifiedAt.IsZero() {
			where[models.UserTable.EmailVerifiedAt] = map[string]any{
				"_neq": nil,
			}
		} else {
			where[models.UserTable.EmailVerifiedAt] = map[string]any{
				"_gte": user.EmailVerifiedAt,
			}
		}
	}

	return &where
}

func (s *DbUserStore) FindUser(ctx context.Context, user *models.User) (*models.User, error) {
	where := s.UserWhere(user)
	return crudrepo.User.GetOne(
		ctx,
		s.db,
		where,
	)
}

func (s *DbUserStore) FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	return s.FindUser(
		ctx,
		&models.User{
			ID: userId,
		},
	)
}

// AssignUserRoles implements UserStore.
func (s *DbUserStore) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	if len(roleNames) > 0 {
		user, err := crudrepo.User.GetOne(
			ctx,
			s.db,
			&map[string]any{
				models.UserTable.ID: map[string]any{
					"_eq": userId,
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
			s.db,
			&map[string]any{
				models.RoleTable.Name: map[string]any{
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
			_, err = crudrepo.UserRole.Post(ctx, s.db, userRoles)
			if err != nil {
				return fmt.Errorf("error assigning user role while assigning roles: %w", err)
			}
		}
	}
	return nil
}

// DeleteUser implements UserStore.
func (s *DbUserStore) DeleteUser(ctx context.Context, userId uuid.UUID) error {
	_, err := crudrepo.User.Delete(ctx, s.db, &map[string]any{
		models.UserTable.ID: map[string]any{"_eq": userId},
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
func (s *DbUserStore) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	type rolePermissionClaims struct {
		UserID      uuid.UUID          `json:"user_id" db:"user_id"`
		Email       string             `json:"email" db:"email"`
		Roles       []string           `json:"roles" db:"roles"`
		Permissions []string           `json:"permissions" db:"permissions"`
		Providers   []models.Providers `json:"providers" db:"providers"`
	}
	user, err := crudrepo.User.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.UserTable.Email: map[string]any{
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
		res, err := pgxscan.One(ctx, s.db, scan.StructMapper[rolePermissionClaims](), RawGetUserWithAllRolesAndPermissionsByEmail, email)
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
func (s *DbUserStore) UpdateUser(ctx context.Context, user *models.User) error {
	_, err := crudrepo.User.PutOne(ctx, s.db, &models.User{
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

// FindUserByEmail implements UserStore.

// CreateUser implements UserStore.
func (s *DbUserStore) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	return crudrepo.User.PostOne(
		ctx,
		s.db,
		user,
	)
}

func (s *DbUserStore) LoadUsersByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.User, error) {
	users, err := crudrepo.User.Get(
		ctx,
		s.db,
		&map[string]any{
			models.UserTable.ID: map[string]any{
				"_in": userIds,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToPointer(users, userIds, func(a *models.User) uuid.UUID {
		if a == nil {
			return uuid.UUID{}
		}
		return a.ID
	}), nil
}

type DbUserStoreInterface interface {
	WithTx(tx database.Dbx) *DbUserStore
	UserWhere(user *models.User) *map[string]any
	FindUser(ctx context.Context, user *models.User) (*models.User, error)
	FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error)
	AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	DeleteUser(ctx context.Context, userId uuid.UUID) error
	GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error)
	UpdateUser(ctx context.Context, user *models.User) error
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	LoadUsersByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.User, error)
}

type RolePermissionClaims struct {
	UserID      uuid.UUID          `json:"user_id" db:"user_id"`
	Email       string             `json:"email" db:"email"`
	Roles       []string           `json:"roles" db:"roles"`
	Permissions []string           `json:"permissions" db:"permissions"`
	Providers   []models.Providers `json:"providers" db:"providers"`
}

// FindUserWithRolesAndPermissionsByEmail retrieves a user's roles and permissions
// from the database based on their email address.
// It expects a database connection (or transaction) `db` and the `email` of the user.
// It returns a pointer to a RolePermissionClaims struct containing the user's
// roles and permissions, or an error if the user is not found or if any other
// database error occurs.
func FindUserWithRolesAndPermissionsByEmail(ctx context.Context, db database.Dbx, email string) (*RolePermissionClaims, error) {
	res, err := pgxscan.One(ctx, db, scan.StructMapper[RolePermissionClaims](), RawGetUserWithAllRolesAndPermissionsByEmail, email)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func FindUserAccountByUserIdAndProvider(ctx context.Context, db database.Dbx, userId uuid.UUID, provider shared.Providers) (*models.UserAccount, error) {
	return crudrepo.UserAccount.GetOne(ctx, db, &map[string]any{
		"user_id": map[string]any{
			"_eq": userId.String(),
		},
		"provider": map[string]any{
			"_eq": provider.String(),
		},
	})
}

func LoadUsersByUserIds(ctx context.Context, db database.Dbx, userIds ...uuid.UUID) ([]*models.User, error) {
	users, err := crudrepo.User.Get(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_in": userIds,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToPointer(users, userIds, func(a *models.User) uuid.UUID {
		if a == nil {
			return uuid.UUID{}
		}
		return a.ID
	}), nil
}

func CreateUser(ctx context.Context, db database.Dbx, params *shared.AuthenticationInput) (*models.User, error) {
	return crudrepo.User.PostOne(ctx, db, &models.User{
		Email:           params.Email,
		Name:            params.Name,
		Image:           params.AvatarUrl,
		EmailVerifiedAt: params.EmailVerifiedAt,
	})
}

func CreateUserRoles(ctx context.Context, db database.Dbx, userId uuid.UUID, roleIds ...uuid.UUID) error {
	var dtos []models.UserRole
	for _, id := range roleIds {
		dtos = append(dtos, models.UserRole{
			UserID: userId,
			RoleID: id,
		})
	}
	_, err := crudrepo.UserRole.Post(
		ctx,
		db,
		dtos,
	)
	if err != nil {

		return err
	}
	return nil
}
func CreateUserPermissions(ctx context.Context, db database.Dbx, userId uuid.UUID, permissionIds ...uuid.UUID) error {
	var dtos []models.UserPermission
	for _, id := range permissionIds {
		dtos = append(dtos, models.UserPermission{
			UserID:       userId,
			PermissionID: id,
		})
	}
	_, err := crudrepo.UserPermission.Post(
		ctx,
		db,
		dtos,
	)
	if err != nil {
		return err
	}
	return nil
}

func CreateAccount(ctx context.Context, db database.Dbx, userId uuid.UUID, params *shared.AuthenticationInput) (*models.UserAccount, error) {
	r, err := crudrepo.UserAccount.PostOne(ctx, db, &models.UserAccount{
		UserID:            userId,
		Type:              models.ProviderTypes(params.Type),
		Password:          params.HashPassword,
		Provider:          models.Providers(params.Provider),
		ProviderAccountID: params.ProviderAccountID,
		AccessToken:       params.AccessToken,
		RefreshToken:      params.RefreshToken,
	})
	return database.OptionalRow(r, err)
}

func FindUserByEmail(ctx context.Context, db database.Dbx, email string) (*models.User, error) {
	a, err := crudrepo.User.GetOne(
		ctx,
		db,
		&map[string]any{
			"email": map[string]any{
				"_eq": email,
			},
		},
	)
	return database.OptionalRow(a, err)
}
func FindUserByID(ctx context.Context, db database.Dbx, userId uuid.UUID) (*models.User, error) {
	a, err := crudrepo.User.GetOne(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
	return database.OptionalRow(a, err)
}

func UpdateUserPassword(ctx context.Context, db database.Dbx, userId uuid.UUID, password string) error {
	account, err := crudrepo.UserAccount.GetOne(
		ctx,
		db,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": userId.String(),
			},
			"provider": map[string]any{
				"_eq": string(models.ProvidersCredentials),
			},
		},
	)
	if err != nil {
		return err
	}
	if account == nil {
		return errors.New("user ProvidersCredentials account not found")
	}
	hash, err := security.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}
	account.Password = &hash
	_, err = crudrepo.UserAccount.PutOne(
		ctx,
		db,
		account,
	)
	if err != nil {
		return err
	}
	return nil
}

func UpdateMe(ctx context.Context, db database.Dbx, userId uuid.UUID, input *shared.UpdateMeInput) error {
	user, err := crudrepo.User.GetOne(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	_, err = crudrepo.User.PutOne(
		ctx,
		db,
		&models.User{
			ID:        userId,
			Name:      input.Name,
			Image:     input.Image,
			UpdatedAt: time.Now(),
		},
	)

	if err != nil {
		return err
	}
	return nil
}

func GetUserAccounts(ctx context.Context, db database.Dbx, userIds ...uuid.UUID) ([][]*models.UserAccount, error) {
	// var results []JoinedResult[*models.Permission, uuid.UUID]
	ids := []string{}
	for _, id := range userIds {
		ids = append(ids, id.String())
	}
	data, err := crudrepo.UserAccount.Get(
		ctx,
		db,
		&map[string]any{
			"user_id": map[string]any{
				"_in": userIds,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToManyPointer(data, userIds, func(a *models.UserAccount) uuid.UUID {
		return a.UserID
	}), nil
}
