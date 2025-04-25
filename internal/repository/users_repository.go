package repository

import (
	"context"
	"errors"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
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

func UpdateUserEmailConfirm(ctx context.Context, db bob.Executor, userId uuid.UUID, emailVerifiedAt time.Time) (*models.User, error) {
	user, err := FindUserById(ctx, db, userId)
	if err != nil {
		return nil, err
	}
	if user.EmailVerifiedAt.IsSet() {
		return user, nil
	}
	err = user.Update(ctx, db, &models.UserSetter{
		EmailVerifiedAt: omitnull.From(emailVerifiedAt),
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

func CreateAccount(ctx context.Context, db bob.Executor, userId uuid.UUID, params *shared.AuthenticateUserParams) (*models.UserAccount, error) {
	r, err := models.UserAccounts.Insert(&models.UserAccountSetter{
		UserID:            omit.From(userId),
		Type:              omit.From(params.Type),
		Provider:          omit.From(params.Provider),
		ProviderAccountID: omit.From(params.ProviderAccountID),
		Password:          omitnull.FromPtr(params.HashPassword),
		AccessToken:       omitnull.FromPtr(params.AccessToken),
		RefreshToken:      omitnull.FromPtr(params.RefreshToken),
	}, im.Returning("*")).One(ctx, db)
	return OptionalRow(r, err)
}
func FindUserAccountByUserIdAndProvider(ctx context.Context, db bob.Executor, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
	acc, err := models.UserAccounts.Query(
		models.SelectWhere.UserAccounts.Provider.EQ(provider),
		models.SelectWhere.UserAccounts.UserID.EQ(userId),
	).One(ctx, db)
	return OptionalRow(acc, err)
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
WHERE u.email = ?
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

func FindUserWithRolesAndPermissionsByEmail(ctx context.Context, db bob.Executor, email string) (*RolePermissionClaims, error) {
	query := psql.RawQuery(RawGetUserWithAllRolesAndPermissionsByEmail, email)

	res, err := bob.One(ctx, db, query, scan.StructMapper[RolePermissionClaims]())
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func FindUserByEmail(ctx context.Context, db bob.Executor, email string) (*models.User, error) {
	a, err := models.Users.Query(models.SelectWhere.Users.Email.EQ(email)).One(ctx, db)
	return OptionalRow(a, err)
}
func FindUserById(ctx context.Context, db bob.Executor, userId uuid.UUID) (*models.User, error) {
	a, err := models.Users.Query(models.SelectWhere.Users.ID.EQ(userId)).One(ctx, db)
	return OptionalRow(a, err)
}

func UpdateUserAccount(ctx context.Context, db bob.Executor, account *models.UserAccount) error {
	return account.Update(ctx, db, &models.UserAccountSetter{
		UserID:            omit.From(account.UserID),
		Type:              omit.From(account.Type),
		Provider:          omit.From(account.Provider),
		ProviderAccountID: omit.From(account.ProviderAccountID),
		Password:          omitnull.FromNull(account.Password),
		RefreshToken:      omitnull.FromNull(account.RefreshToken),
		AccessToken:       omitnull.FromNull(account.AccessToken),
		ExpiresAt:         omitnull.FromNull(account.ExpiresAt),
		IDToken:           omitnull.FromNull(account.IDToken),
		Scope:             omitnull.FromNull(account.Scope),
		SessionState:      omitnull.FromNull(account.SessionState),
		TokenType:         omitnull.FromNull(account.TokenType),
		CreatedAt:         omit.From(account.CreatedAt),
		UpdatedAt:         omit.From(account.UpdatedAt),
	})
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

func UpdateMe(ctx context.Context, db bob.Executor, userId uuid.UUID, input *shared.UpdateMeInput) error {
	q := models.Users.Update(
		models.UpdateWhere.Users.ID.EQ(userId),
		models.UserSetter{
			Name:      omitnull.FromPtr(input.Name),
			Image:     omitnull.FromPtr(input.Image),
			UpdatedAt: omit.From(time.Now()),
		}.UpdateMod(),
	)
	_, err := q.Exec(ctx, db)
	if err != nil {
		return err
	}
	return nil
}
