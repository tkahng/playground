package repository_test

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/types"
)

var (
	roleNames, permissionNames = []string{"admin", "advanced", "pro", "basic"}, []string{"admin", "advanced", "pro", "basic"}
)

func TestAuth_UserAccountRbac(t *testing.T) {
	t.Parallel()
	database.WithNewTx(t, func(ctx context.Context, dbx database.Dbx) {
		// init
		repository.InitRbac(t, dbx)
		// find all
		roles := repository.MustFindAll(t, repository.Role, dbx, nil)
		if len(roles) != 4 {
			t.Errorf("expected 4 roles, got %d", len(roles))
		}
		permissions := repository.MustFindAll(t, repository.Permission, dbx, nil)
		if len(permissions) != 4 {
			t.Errorf("expected 4 permissions, got %d", len(permissions))
		}
		// find roles with permissions
		rolesWithPermissions := repository.MustFindAll(t, repository.Role, dbx, &map[string]any{
			"permissions": map[string]any{
				"name": map[string]any{
					"_eq": "admin",
				},
			},
		})
		if len(rolesWithPermissions) != 1 && rolesWithPermissions[0].Name != "admin" {
			t.Errorf("expected 1 roles with permissions, got %d", len(rolesWithPermissions))
		}
		rolesWithPermissions = repository.MustFindAll(t, repository.Role, dbx, &map[string]any{
			"permissions": map[string]any{
				"name": map[string]any{
					"_eq": "basic",
				},
			},
		})
		if len(rolesWithPermissions) != 4 {
			t.Errorf("expected 4 roles with permissions, got %d", len(rolesWithPermissions))
		}
		rolesWithPermissions = repository.MustFindAll(t, repository.Role, dbx, &map[string]any{
			"_or": []map[string]any{
				{
					"name": map[string]any{
						"_eq": "admin",
					},
				},
				{
					"permissions": map[string]any{
						"name": map[string]any{
							"_eq": "pro",
						},
					},
				},
			},
		})
		if len(rolesWithPermissions) != 3 {
			t.Errorf("expected 2 roles with permissions, got %d", len(rolesWithPermissions))
		}
	})
}

func TestMustCreateUserAndAccount(t *testing.T) {
	database.WithNewTx(t, func(ctx context.Context, db database.Dbx) {
		tests := []struct {
			name string // description of this test case
			// Named input parameters for target function.
			db        database.Dbx
			fns       []func(*models.User, *models.UserAccount)
			predicate func(*models.User, *models.UserAccount)
		}{
			{
				name: "Create user and credential account",
				db:   db,
				fns: []func(user *models.User, account *models.UserAccount){
					func(user *models.User, account *models.UserAccount) {
						user.Name = types.Pointer("credential")
						account.Type = models.ProviderTypeCredentials
						account.Provider = models.ProvidersCredentials
					},
				},
				predicate: func(user *models.User, account *models.UserAccount) {
					if user == nil {
						t.Errorf("expected user, got nil")
					}
					if account == nil {
						t.Errorf("expected account, got nil")
					}
					if user.Name != types.Pointer("credential") {
						t.Errorf("expected user name to be credential, got %s", *user.Name)
					}
					if account.Type != models.ProviderTypeCredentials {
						t.Errorf("expected account type to be credentials, got %s", account.Type)
					}
					if account.Provider != models.ProvidersCredentials {
						t.Errorf("expected account provider to be credentials, got %s", account.Provider)
					}
				},
			},
			{
				name: "create google account",
				db:   db,
				fns: []func(user *models.User, account *models.UserAccount){
					func(user *models.User, account *models.UserAccount) {
						user.Name = types.Pointer("google")
						account.Type = models.ProviderTypeOAuth
						account.Provider = models.ProvidersGoogle
					},
				},
				predicate: func(user *models.User, account *models.UserAccount) {
					if user == nil {
						t.Errorf("expected user, got nil")
					}
					if account == nil {
						t.Errorf("expected account, got nil")
					}
					if user.Name != types.Pointer("google") {
						t.Errorf("expected user name to be google, got %s", *user.Name)
					}
					if account.Type != models.ProviderTypeOAuth {
						t.Errorf("expected account type to be oauth, got %s", account.Type)
					}
					if account.Provider != models.ProvidersGoogle {
						t.Errorf("expected account provider to be google, got %s", account.Provider)
					}
				},
			},
			{
				name: "create github account",
				db:   db,
				fns: []func(user *models.User, account *models.UserAccount){
					func(user *models.User, account *models.UserAccount) {
						user.Name = types.Pointer("github")
						account.Type = models.ProviderTypeOAuth
						account.Provider = models.ProvidersGithub
					},
				},
				predicate: func(user *models.User, account *models.UserAccount) {
					if user == nil {
						t.Errorf("expected user, got nil")
					}
					if account == nil {
						t.Errorf("expected account, got nil")
					}
					if user.Name != types.Pointer("github") {
						t.Errorf("expected user name to be github, got %s", *user.Name)
					}
					if account.Type != models.ProviderTypeOAuth {
						t.Errorf("expected account type to be oauth, got %s", account.Type)
					}
					if account.Provider != models.ProvidersGithub {
						t.Errorf("expected account provider to be github, got %s", account.Provider)
					}
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				user, account := repository.MustCreateUserAndAccount(t, tt.db, tt.fns...)
				tt.predicate(user, account)
			})
		}
	})
}

func TestMustCreateUserAndAccount_Randomize(t *testing.T) {
	database.WithNewTx(t, func(ctx context.Context, db database.Dbx) {
		var count int64 = 100
		for range count {
			repository.MustCreateUserAndAccount(t, db)
		}
		userCount := repository.MustCountAll(t, repository.User, db, &map[string]any{})
		if userCount != count {
			t.Errorf("expected at least 10 users, got %d", userCount)
		}
		countAcc := repository.MustCountAll(t, repository.User, db, &map[string]any{})
		if countAcc != count {
			t.Errorf("expected at least 10 accounts, got %d", countAcc)
		}
	})
}
