package stores

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/internal/tools/utils"

	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/playground/internal/repository"
)

type UserFilter struct {
	PaginatedInput
	SortParams
	Providers     []models.Providers        `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	Q             string                    `query:"q,omitempty" required:"false"`
	Ids           []uuid.UUID               `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Emails        []string                  `query:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	RoleIds       []uuid.UUID               `query:"role_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	EmailVerified types.OptionalParam[bool] `query:"email_verified,omitempty" required:"false"`
}

type DbUserStoreInterface interface {
	FindUser(ctx context.Context, user *UserFilter) (*models.User, error)
	FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error)
	AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	DeleteUser(ctx context.Context, userId uuid.UUID) error
	GetUserInfo(ctx context.Context, email string) (*models.UserInfo, error)
	UpdateUser(ctx context.Context, user *models.User) error
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	LoadUsersByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.User, error)
	FindUsers(ctx context.Context, filter *UserFilter) ([]*models.User, error)
	CountUsers(ctx context.Context, filter *UserFilter) (int64, error)
}

type DbUserStore struct {
	db database.Dbx
}

// CountUsers implements DbUserStoreInterface.
func (s *DbUserStore) CountUsers(ctx context.Context, filter *UserFilter) (int64, error) {
	where := s.filter(filter)
	if where == nil {
		return 0, nil // no filter, return 0
	}
	count, err := repository.User.Count(ctx, s.db, where)
	if err != nil {
		return 0, fmt.Errorf("error counting users: %w", err)
	}
	return count, nil
}

func pagination(filter Paginable) (limit, offset int) {
	if filter == nil {
		return 10, 0 // default values
	}
	return filter.LimitOffset()
}

func queryPagination(q squirrel.SelectBuilder, filter Paginable) squirrel.SelectBuilder {
	limit, offset := pagination(filter)
	q = q.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}

func (s *DbUserStore) sort(filter Sortable) *map[string]string {
	if filter == nil {
		return nil // return nil if no filter is provided
	}

	sortBy, sortOrder := filter.Sort()
	if sortBy != "" && slices.Contains(repository.UserBuilder.ColumnNames(), utils.Quote(sortBy)) {
		return &map[string]string{
			sortBy: sortOrder,
		}
	} else {
		slog.Info("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder, "columns", repository.UserBuilder.ColumnNames())
	}

	return nil // default no sorting
}

// FindUsers implements DbUserStoreInterface.
func (s *DbUserStore) FindUsers(ctx context.Context, filter *UserFilter) ([]*models.User, error) {
	where := s.filter(filter)
	if where == nil {
		return nil, nil // no filter, return empty slice
	}
	sort := s.sort(filter)
	limit, offset := pagination(filter)
	users, err := repository.User.Get(
		ctx,
		s.db,
		where,
		sort,
		types.Pointer(limit),
		types.Pointer(offset),
	)
	if err != nil {
		return nil, fmt.Errorf("error finding users: %w", err)
	}
	return users, nil
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

func (s *DbUserStore) filter(filter *UserFilter) *map[string]any {
	where := make(map[string]any)
	if filter == nil {
		return &where // return empty map if no filter is provided
	}

	if filter.EmailVerified.IsSet {
		emailverified := filter.EmailVerified.Value
		if emailverified {
			where[models.UserTable.EmailVerifiedAt] = map[string]any{
				repository.IsNotNull: nil,
			}
		} else {
			where[models.UserTable.EmailVerifiedAt] = map[string]any{
				repository.IsNull: nil,
			}
		}
	}
	if len(filter.Emails) > 0 {
		where["email"] = map[string]any{
			"_in": filter.Emails,
		}
	}
	if len(filter.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.Providers) > 0 {
		where["accounts"] = map[string]any{
			"provider": map[string]any{
				"_in": filter.Providers,
			},
		}
	}
	if len(filter.RoleIds) > 0 {
		where["roles"] = map[string]any{
			"id": map[string]any{
				"_in": filter.RoleIds,
			},
		}
	}
	if filter.Q != "" {
		where["_or"] = []map[string]any{
			{
				"email": map[string]any{
					"_ilike": "%" + filter.Q + "%",
				},
			},
			{
				"name": map[string]any{
					"_ilike": "%" + filter.Q + "%",
				},
			},
		}
	}
	if len(where) == 0 {
		return nil
	}
	return &where
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

func (s *DbUserStore) FindUser(ctx context.Context, user *UserFilter) (*models.User, error) {
	where := s.filter(user)
	return repository.User.GetOne(
		ctx,
		s.db,
		where,
	)
}

func (s *DbUserStore) FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	return s.FindUser(
		ctx,
		&UserFilter{
			Ids: []uuid.UUID{userId},
		},
	)
}

// AssignUserRoles implements UserStore.
func (s *DbUserStore) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	if len(roleNames) > 0 {
		user, err := repository.User.GetOne(
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
		roles, err := repository.Role.Get(
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
			_, err = repository.UserRole.Post(ctx, s.db, userRoles)
			if err != nil {
				return fmt.Errorf("error assigning user role while assigning roles: %w", err)
			}
		}
	}
	return nil
}

// DeleteUser implements UserStore.
func (s *DbUserStore) DeleteUser(ctx context.Context, userId uuid.UUID) error {
	_, err := repository.User.Delete(ctx, s.db, &map[string]any{
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
func (s *DbUserStore) GetUserInfo(ctx context.Context, email string) (*models.UserInfo, error) {
	type rolePermissionClaims struct {
		UserID      uuid.UUID          `json:"user_id" db:"user_id"`
		Email       string             `json:"email" db:"email"`
		Roles       []string           `json:"roles" db:"roles"`
		Permissions []string           `json:"permissions" db:"permissions"`
		Providers   []models.Providers `json:"providers" db:"providers"`
	}
	user, err := repository.User.GetOne(
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
		return nil, errors.New("user not found")
	}
	result := &models.UserInfo{
		User: *user,
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

	result.Roles = roles.Roles
	result.Permissions = roles.Permissions
	result.Providers = roles.Providers

	return result, nil
}

// UpdateUser implements UserStore.
func (s *DbUserStore) UpdateUser(ctx context.Context, user *models.User) error {
	_, err := repository.User.PutOne(ctx, s.db, &models.User{
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
	return repository.User.PostOne(
		ctx,
		s.db,
		user,
	)
}

func (s *DbUserStore) LoadUsersByUserIds(ctx context.Context, userIds ...uuid.UUID) ([]*models.User, error) {
	users, err := repository.User.Get(
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

var _ DbUserStoreInterface = &DbUserStore{}

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

func LoadUsersByUserIds(ctx context.Context, db database.Dbx, userIds ...uuid.UUID) ([]*models.User, error) {
	users, err := repository.User.Get(
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
