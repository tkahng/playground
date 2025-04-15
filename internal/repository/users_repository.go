package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/scan"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
)

type RolesMap map[string]*models.Role

type PermissionsMap map[string]*models.Permission

type RoleStructTree map[string]RoleDto

type RoleDto struct {
	Role        *models.Role
	Permissions []*models.Permission
}

func CreateUser(ctx context.Context, db bob.Executor, params *shared.AuthenticateUserParams) (*models.User, error) {
	return models.Users.Insert(&models.UserSetter{
		Email:           omit.From(params.Email),
		Name:            omitnull.FromPtr(params.Name),
		Image:           omitnull.FromPtr(params.AvatarUrl),
		EmailVerifiedAt: omitnull.FromPtr(params.EmailVerifiedAt),
	}, im.Returning("*")).One(ctx, db)
}

func UpdateUserEmailConfirm(ctx context.Context, db bob.Executor, userId uuid.UUID) (*models.User, error) {
	user, err := GetUserById(ctx, db, userId)
	if err != nil {
		return nil, err
	}
	if user.EmailVerifiedAt.IsSet() {
		return user, nil
	}
	err = user.Update(ctx, db, &models.UserSetter{
		EmailVerifiedAt: omitnull.From(time.Now()),
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func AssignRoles(ctx context.Context, db bob.Executor, user *models.User, params ...*models.Role) error {
	return user.AttachRoles(ctx, db, params...)
}

func AssignPermissions(ctx context.Context, db bob.Executor, user *models.User, params *models.Permission) error {
	return user.AttachPermissions(ctx, db, params)
}

func CreateAccount(ctx context.Context, db bob.Executor, user *models.User, params *shared.AuthenticateUserParams) (*models.UserAccount, error) {
	r, err := models.UserAccounts.Insert(&models.UserAccountSetter{
		UserID:            omit.From(user.ID),
		Type:              omit.From(params.Type),
		Provider:          omit.From(params.Provider),
		ProviderAccountID: omit.From(params.ProviderAccountID),
		Password:          omitnull.FromPtr(params.HashPassword),
		AccessToken:       omitnull.FromPtr(params.AccessToken),
		RefreshToken:      omitnull.FromPtr(params.RefreshToken),
	}, im.Returning("*")).One(ctx, db)
	return OptionalRow(r, err)
}

func GetUserAccountByProviderAndEmail(ctx context.Context, db bob.Executor, email string, provider models.Providers) (*shared.AuthenticateUserState, error) {
	user, err := models.Users.Query(models.SelectWhere.Users.Email.EQ(email)).One(ctx, db)
	user, err = OptionalRow(user, err)
	if err != nil {
		return nil, fmt.Errorf("error while to get user: %w", err)
	}
	if user == nil {
		return &shared.AuthenticateUserState{}, nil
	}

	acc, err := models.
		UserAccounts.
		Query(
			psql.WhereAnd(
				models.SelectWhere.UserAccounts.UserID.EQ(
					user.ID,
				),
				models.SelectWhere.UserAccounts.Provider.EQ(
					provider,
				),
			),
		).One(ctx, db)
	acc, err = OptionalRow(acc, err)
	if err != nil {
		return nil, err
	}
	return &shared.AuthenticateUserState{
		User:    user,
		Account: acc,
	}, nil
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
        JOIN public.permissions p ON rp.permission_id = p.id -- WHERE ur.user_id = '575a9f91-159e-4680-ba8b-3fc4db40d194'
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
        JOIN public.permissions p ON rp.permission_id = p.id -- WHERE u.id = '575a9f91-159e-4680-ba8b-3fc4db40d194'
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
WHERE u.email = ?
GROUP BY u.id
LIMIT 1;
`
)

type rolePermissionClaims struct {
	UserID      uuid.UUID      `json:"user_id" db:"user_id"`
	Email       string         `json:"email" db:"email"`
	Roles       pq.StringArray `json:"roles" db:"roles"`
	Permissions pq.StringArray `json:"permissions" db:"permissions"`
	Providers   pq.StringArray `json:"providers" db:"providers"`
}
type RolePermissionClaims struct {
	UserID      uuid.UUID          `json:"user_id" db:"user_id"`
	Email       string             `json:"email" db:"email"`
	Roles       []string           `json:"roles" db:"roles"`
	Permissions []string           `json:"permissions" db:"permissions"`
	Providers   []models.Providers `json:"providers" db:"providers"`
}

func GetUserWithRolesAndPermissions(ctx context.Context, db bob.Executor, email string) (*RolePermissionClaims, error) {
	query := psql.RawQuery(RawGetUserWithAllRolesAndPermissionsByEmail, email)

	res, err := bob.One(ctx, db, query, scan.StructMapper[rolePermissionClaims]())
	if err != nil {
		return nil, err
	}
	var prov []models.Providers
	all := models.AllProviders()
	for _, provider := range res.Providers {
		for _, p := range all {
			if provider == string(p) {
				prov = append(prov, p)
			}
		}
	}
	return &RolePermissionClaims{
		UserID:      res.UserID,
		Email:       res.Email,
		Roles:       res.Roles,
		Permissions: res.Permissions,
		Providers:   prov,
	}, nil
}

// func GetUsersWithRolesAndPermissions(ctx context.Context, db bob.Executor, ids ...uuid.UUID) ([]RolePermissionClaims, error) {
// 	var input []any
// 	for _, id := range ids {
// 		input = append(input, id)
// 	}

// 	query := psql.RawQuery(rawGetUsersWithPermissionsByIds, psql.Arg(input...))
// 	res, err := bob.All(ctx, db, query, scan.StructMapper[rolePermissionClaims]())
// 	if err != nil {
// 		return nil, err
// 	}
// 	var claims []RolePermissionClaims
// 	all := models.AllProviders()
// 	for _, r := range res {
// 		var prov []models.Providers
// 		var claim RolePermissionClaims = RolePermissionClaims{
// 			UserID:      r.UserID,
// 			Email:       r.Email,
// 			Roles:       r.Roles,
// 			Permissions: r.Permissions,
// 		}
// 		for _, provider := range r.Providers {
// 			for _, p := range all {
// 				if provider == string(p) {
// 					prov = append(prov, p)
// 				}
// 			}
// 		}
// 		claim.Providers = prov
// 		claims = append(claims, claim)
// 	}
// 	claims = dataloader.MapTo(claims, ids, func(c RolePermissionClaims) uuid.UUID {
// 		return c.UserID
// 	})
// 	return claims, nil

// }

func GetUserByEmail(ctx context.Context, db bob.Executor, email string) (*models.User, error) {
	a, err := models.Users.Query(models.SelectWhere.Users.Email.EQ(email)).One(ctx, db)
	return OptionalRow(a, err)
}
func GetUserById(ctx context.Context, db bob.Executor, userId uuid.UUID) (*models.User, error) {
	a, err := models.Users.Query(models.SelectWhere.Users.ID.EQ(userId)).One(ctx, db)
	return OptionalRow(a, err)
}

// // func GetUserInfo(ctx context.Context, db bob.Executor,email string) (*models.User, error) {
func GetAdminUserCount(ctx context.Context, db bob.Executor) (int64, error) {
	a, err := models.Roles.Query(models.SelectWhere.Roles.Name.EQ("admin")).One(ctx, db)
	a, err = OptionalRow(a, err)
	if err != nil {
		panic(err)
	}
	usecount, err := a.Users().Count(ctx, db)
	if err != nil {
		panic(err)
	}
	if usecount == 0 {
		// user, err := CreateUser()
	}
	return usecount, nil
}

func UpdateUserPassword(ctx context.Context, db bob.Executor, userId uuid.UUID, password string) error {
	user, err := models.FindUser(ctx, db, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	account, err := user.UserAccounts(
		models.SelectWhere.UserAccounts.Provider.EQ(models.ProvidersCredentials),
	).One(ctx, db)
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
	return account.Update(ctx, db, &models.UserAccountSetter{
		Password: omitnull.From(hash),
	})
}
