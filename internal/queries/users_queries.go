package queries

import (
	"context"
	"errors"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	crudModels "github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/crud/crudrepo"
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
func CreateUserPermissions(ctx context.Context, db Queryer, userId uuid.UUID, permissionIds ...uuid.UUID) error {
	var dtos []crudModels.UserPermission
	for _, id := range permissionIds {
		dtos = append(dtos, crudModels.UserPermission{
			UserID:       userId,
			PermissionID: id,
		})
	}
	_, err := crudrepo.UserPermission.Post(ctx, db, dtos)
	return err

}

func CreateAccount(ctx context.Context, db Queryer, userId uuid.UUID, params *shared.AuthenticateUserParams) (*crudModels.UserAccount, error) {
	r, err := crudrepo.UserAccount.PostOne(ctx, db, &crudModels.UserAccount{
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
func FindUserById(ctx context.Context, db Queryer, userId uuid.UUID) (*crudModels.User, error) {
	a, err := crudrepo.User.GetOne(
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

func UpdateUserPassword(ctx context.Context, db Queryer, userId uuid.UUID, password string) error {
	account, err := crudrepo.UserAccount.GetOne(
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

func UpdateMe(ctx context.Context, db Queryer, userId uuid.UUID, input *shared.UpdateMeInput) error {
	_, err := crudrepo.User.PutOne(
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
