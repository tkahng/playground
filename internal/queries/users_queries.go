package queries

import (
	"context"
	"errors"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/tkahng/authgo/internal/crud/crudrepo"
	crudModels "github.com/tkahng/authgo/internal/crud/models"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
)

func CreateUser(ctx context.Context, db Queryer, params *shared.AuthenticateUserParams) (*crudModels.User, error) {
	// return models.Users.Insert(&models.UserSetter{
	// 	Email:           omit.From(params.Email),
	// 	Name:            omitnull.FromPtr(params.Name),
	// 	Image:           omitnull.FromPtr(params.AvatarUrl),
	// 	EmailVerifiedAt: omitnull.FromPtr(params.EmailVerifiedAt),
	// }, im.Returning("*")).One(ctx, db)
	return crudrepo.User.PostOne(ctx, db, &crudModels.User{
		Email:           params.Email,
		Name:            params.Name,
		Image:           params.AvatarUrl,
		EmailVerifiedAt: params.EmailVerifiedAt,
	})
}

func CreateUserRoles(ctx context.Context, db Queryer, userId uuid.UUID, roleIds ...uuid.UUID) error {
	var dtos []crudModels.UserRole
	for _, id := range roleIds {
		dtos = append(dtos, crudModels.UserRole{
			UserID: userId,
			RoleID: id,
		})
	}
	_, err := crudrepo.UserRole.Post(ctx, db, dtos)
	return err
	// return models.UserRoles.Insert(&models.UserRoleSetter{
	// 	UserID:  omit.From(userId),
	// 	RoleIDs: omit.From(params.RoleIds),
	// }).Exec(ctx, db)
}

func UpdateUserEmailConfirm(ctx context.Context, db Queryer, userId uuid.UUID, emailVerifiedAt time.Time) (*models.User, error) {
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

func CreateAccount(ctx context.Context, db Queryer, userId uuid.UUID, params *shared.AuthenticateUserParams) (*models.UserAccount, error) {
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

func FindUserByEmail(ctx context.Context, db Queryer, email string) (*crudModels.User, error) {
	a, err := crudrepo.User.GetOne(
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
func FindUserById(ctx context.Context, db Queryer, userId uuid.UUID) (*models.User, error) {
	a, err := models.Users.Query(models.SelectWhere.Users.ID.EQ(userId)).One(ctx, db)
	return OptionalRow(a, err)
}

func UpdateUserPassword(ctx context.Context, db Queryer, userId uuid.UUID, password string) error {
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

func UpdateMe(ctx context.Context, db Queryer, userId uuid.UUID, input *shared.UpdateMeInput) error {
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
