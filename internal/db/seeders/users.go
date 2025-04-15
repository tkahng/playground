package seeders

import (
	"context"
	"log/slog"
	"time"

	"github.com/aarondl/opt/null"
	"github.com/alexedwards/argon2id"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/db/models/factory"
	"github.com/tkahng/authgo/internal/tools/security"
)

func UserCredentialsFactory(ctx context.Context, dbx bob.DB, count int) error {
	f := factory.New()
	// fake := faker.New()
	hash, _ := security.CreateHash("password", argon2id.DefaultParams)
	f.AddBaseUserMod(
		factory.UserMods.RandomEmail(nil),
		factory.UserMods.AddNewRoles(1,
			factory.RoleMods.AddNewPermissions(1,
				factory.PermissionMods.RandomName(nil),
			),
		),
		factory.UserMods.AddNewUserAccounts(1,
			factory.UserAccountMods.Provider(models.ProvidersCredentials),
			factory.UserAccountMods.RandomProviderAccountID(nil),
			factory.UserAccountMods.Password(null.From(hash)),
			factory.UserAccountMods.Type(models.ProviderTypesCredentials),
		),
	)
	usertemplate := f.NewUser()
	_, err := usertemplate.CreateMany(ctx, dbx, count)
	return err
}
func UserCredentialsRolesFactory(ctx context.Context, dbx bob.Executor, count int, roles ...*models.Role) (models.UserSlice, error) {
	f := factory.New()
	hash, err := security.CreateHash("Password123!", argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}
	f.AddBaseUserMod(
		factory.UserMods.RandomEmail(nil),
		factory.UserMods.AddNewUserAccounts(1,
			factory.UserAccountMods.Provider(models.ProvidersCredentials),
			factory.UserAccountMods.RandomProviderAccountID(nil),
			factory.UserAccountMods.Password(null.From(hash)),
			factory.UserAccountMods.Type(models.ProviderTypesCredentials),
		),
	)
	usertemplate := f.NewUser()
	data, err := usertemplate.CreateMany(ctx, dbx, count)
	if err != nil {
		return nil, err
	}
	for _, user := range data {
		err = user.AttachRoles(ctx, dbx, roles...)
		if err != nil {
			slog.Error("error attaching roles", "error", err)
		}
	}
	return data, err
}

func UserTokenFactory(user *models.User, tokenType models.TokenTypes, expires int64) *factory.TokenTemplate {
	f := factory.New()
	// fake := faker.New()
	// hash, _ := security.CreateHash("password", argon2id.DefaultParams)

	f.AddBaseTokenMod(
		factory.TokenMods.UserID(null.From(user.ID)),
		factory.TokenMods.Identifier(user.Email),
		factory.TokenMods.Type(tokenType),
		factory.TokenMods.ExpiresFunc(func() time.Time { return core.Expires(expires) }),
	)
	tokenTemplate := f.NewToken()
	return tokenTemplate
	// _, err := usertemplate.CreateMany(ctx, dbx, count)
	// return err/
}

func UserOauthFactory(ctx context.Context, dbx bob.DB, count int, provider models.Providers) error {
	f := factory.New()
	// fake := faker.New()
	// hash, _ := security.CreateHash("password", argon2id.DefaultParams)
	f.AddBaseUserMod(
		factory.UserMods.RandomEmail(nil),
		factory.UserMods.AddNewRoles(1,
			factory.RoleMods.AddNewPermissions(1,
				factory.PermissionMods.RandomName(nil),
			),
		),
		factory.UserMods.AddNewUserAccounts(1,
			factory.UserAccountMods.Provider(provider),
			factory.UserAccountMods.RandomProviderAccountID(nil),
			// factory.UserAccountMods.Password(null.From(hash)),
			factory.UserAccountMods.Type(models.ProviderTypesOauth),
		),
	)
	usertemplate := f.NewUser()
	_, err := usertemplate.CreateMany(ctx, dbx, count)
	return err
}
