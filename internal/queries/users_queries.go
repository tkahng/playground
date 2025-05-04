package queries

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/stephenafamo/scan"
	"github.com/stephenafamo/scan/pgxscan"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	crudModels "github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/security"
)

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

func FindUserWithRolesAndPermissionsByEmail(ctx context.Context, db db.Dbx, email string) (*RolePermissionClaims, error) {
	res, err := pgxscan.One(ctx, db, scan.StructMapper[RolePermissionClaims](), RawGetUserWithAllRolesAndPermissionsByEmail, email)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func FindUserAccountByUserIdAndProvider(ctx context.Context, db db.Dbx, userId uuid.UUID, provider shared.Providers) (*crudModels.UserAccount, error) {
	return repository.UserAccount.GetOne(ctx, db, &map[string]any{
		"user_id": map[string]any{
			"_eq": userId.String(),
		},
		"provider": map[string]any{
			"_eq": provider.String(),
		},
	})
}

func LoadUsersByUserIds(ctx context.Context, db db.Dbx, userIds ...uuid.UUID) ([]*crudModels.User, error) {
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
	return mapper.MapToPointer(users, userIds, func(a *crudModels.User) uuid.UUID {
		if a == nil {
			return uuid.UUID{}
		}
		return a.ID
	}), nil
}

func CreateUser(ctx context.Context, db db.Dbx, params *shared.AuthenticationInput) (*crudModels.User, error) {
	// return models.Users.Insert(&models.UserSetter{
	// 	Email:           omit.From(params.Email),
	// 	Name:            omitnull.FromPtr(params.Name),
	// 	Image:           omitnull.FromPtr(params.AvatarUrl),
	// 	EmailVerifiedAt: omitnull.FromPtr(params.EmailVerifiedAt),
	// }, im.Returning("*")).One(ctx, db)
	return repository.User.PostOne(ctx, db, &crudModels.User{
		Email:           params.Email,
		Name:            params.Name,
		Image:           params.AvatarUrl,
		EmailVerifiedAt: params.EmailVerifiedAt,
	})
}

func CreateUserRoles(ctx context.Context, db db.Dbx, userId uuid.UUID, roleIds ...uuid.UUID) error {
	var dtos []crudModels.UserRole
	for _, id := range roleIds {
		dtos = append(dtos, crudModels.UserRole{
			UserID: userId,
			RoleID: id,
		})
	}
	q := squirrel.Insert("user_roles").Columns("user_id", "role_id")
	for _, perm := range dtos {
		q = q.Values(perm.UserID, perm.RoleID)
	}
	q = q.Suffix("RETURNING *")
	sql, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}
	fmt.Println(sql, args)
	_, err = QueryAll[crudModels.UserRole](ctx, db, sql, args...)
	if err != nil {
		return err
	}
	return nil
	// return models.UserRoles.Insert(&models.UserRoleSetter{
	// 	UserID:  omit.From(userId),
	// 	RoleIDs: omit.From(params.RoleIds),
	// }).Exec(ctx, db)
}
func CreateUserPermissions(ctx context.Context, db db.Dbx, userId uuid.UUID, permissionIds ...uuid.UUID) error {
	var dtos []crudModels.UserPermission
	for _, id := range permissionIds {
		dtos = append(dtos, crudModels.UserPermission{
			UserID:       userId,
			PermissionID: id,
		})
	}
	q := squirrel.Insert("user_permissions").Columns("user_id", "permission_id")
	for _, perm := range dtos {
		q = q.Values(perm.UserID, perm.PermissionID)
	}
	q = q.Suffix("RETURNING *")
	sql, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return err
	}
	fmt.Println(sql, args)
	_, err = QueryAll[crudModels.UserPermission](ctx, db, sql, args...)
	if err != nil {
		return err
	}
	return nil

}

func CreateAccount(ctx context.Context, db db.Dbx, userId uuid.UUID, params *shared.AuthenticationInput) (*crudModels.UserAccount, error) {
	r, err := repository.UserAccount.PostOne(ctx, db, &crudModels.UserAccount{
		UserID:            userId,
		Type:              crudModels.ProviderTypes(params.Type),
		Password:          params.HashPassword,
		Provider:          crudModels.Providers(params.Provider),
		ProviderAccountID: params.ProviderAccountID,
		AccessToken:       params.AccessToken,
		RefreshToken:      params.RefreshToken,
	})
	return OptionalRow(r, err)
}

func FindUserByEmail(ctx context.Context, db db.Dbx, email string) (*crudModels.User, error) {
	a, err := repository.User.GetOne(
		ctx,
		db,
		&map[string]any{
			"email": map[string]any{
				"_eq": email,
			},
		},
	)
	return OptionalRow(a, err)
}
func FindUserById(ctx context.Context, db db.Dbx, userId uuid.UUID) (*crudModels.User, error) {
	a, err := repository.User.GetOne(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_eq": userId,
			},
		},
	)
	return OptionalRow(a, err)
}

func UpdateUserPassword(ctx context.Context, db db.Dbx, userId uuid.UUID, password string) error {
	account, err := repository.UserAccount.GetOne(
		ctx,
		db,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": userId,
			},
			"provider": map[string]any{
				"_eq": crudModels.ProvidersCredentials,
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
	_, err = repository.UserAccount.PutOne(
		ctx,
		db,
		account,
	)
	if err != nil {
		return err
	}
	return nil
}

func UpdateMe(ctx context.Context, db db.Dbx, userId uuid.UUID, input *shared.UpdateMeInput) error {
	_, err := repository.User.PutOne(
		ctx,
		db,
		&crudModels.User{
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

func GetUserAccounts(ctx context.Context, db db.Dbx, userIds ...uuid.UUID) ([][]*crudModels.UserAccount, error) {
	// var results []JoinedResult[*crudModels.Permission, uuid.UUID]
	ids := []string{}
	for _, id := range userIds {
		ids = append(ids, id.String())
	}
	data, err := repository.UserAccount.Get(
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
	return mapper.MapToManyPointer(data, userIds, func(a *crudModels.UserAccount) uuid.UUID {
		return a.UserID
	}), nil
}
