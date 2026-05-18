package cmd_test

import (
	"context"
	"testing"

	"github.com/tkahng/playground/cmd"
	"github.com/tkahng/playground/internal/auth"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
)

func TestSeedRoles(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		if err := cmd.SeedRoles(ctx, db); err != nil {
			t.Fatalf("SeedRoles() error = %v", err)
		}
		rbacStore := stores.NewDbRBACStore(db)
		for roleName := range shared.KnownRoleNamesPermissionsMap {
			role, err := rbacStore.FindRoleByName(ctx, roleName)
			if err != nil {
				t.Fatalf("FindRoleByName(%q) error = %v", roleName, err)
			}
			if role == nil {
				t.Errorf("role %q not found after SeedRoles", roleName)
			}
		}
		for roleName, permissions := range shared.KnownTeamRolePermissionsMap {
			for _, permission := range permissions {
				allowed, err := rbacStore.HasTeamRolePermission(ctx, models.TeamMemberRole(roleName), permission)
				if err != nil {
					t.Fatalf("HasTeamRolePermission(%q, %q) error = %v", roleName, permission, err)
				}
				if !allowed {
					t.Errorf("team role permission %q/%q not found after SeedRoles", roleName, permission)
				}
			}
		}
	})
}

func TestSeedUser(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "error on missing args",
			args:    []string{"admin@example.com", "Password123!"},
			wantErr: true,
		},
		{
			name:    "error on invalid email",
			args:    []string{"not-an-email", "Password123!", "true"},
			wantErr: true,
		},
		{
			name:    "error on empty email",
			args:    []string{"", "Password123!", "true"},
			wantErr: true,
		},
		{
			name:    "creates user with verified email",
			args:    []string{"verified@example.com", "Password123!", "true"},
			wantErr: false,
		},
		{
			name:    "creates user with unverified email",
			args:    []string{"unverified@example.com", "Password123!", "false"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
				repository.CreateRolesAndPermissions(t, ctx, db, shared.KnownRoleNamesPermissionsMap)
				app := core.NewTestBaseApp(conf.ZeroEnvConfig(), db)

				err := cmd.SeedUser(ctx, app, tt.args)
				if (err != nil) != tt.wantErr {
					t.Errorf("SeedUser() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErr || len(tt.args) != 3 {
					return
				}

				userInfo, err := app.Auth().Signin(ctx, &auth.SigninInput{
					Email:    tt.args[0],
					Password: tt.args[1],
				})
				if err != nil {
					t.Fatalf("Signin() after SeedUser() error = %v", err)
				}
				wantVerified := tt.args[2] == "true"
				gotVerified := userInfo.User.EmailVerifiedAt != nil
				if gotVerified != wantVerified {
					t.Errorf("EmailVerifiedAt verified = %v, want %v", gotVerified, wantVerified)
				}
			})
		})
	}
}

func TestSeedTeam(t *testing.T) {
	t.Run("error on missing args", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			app := core.NewTestBaseApp(conf.ZeroEnvConfig(), db)
			err := cmd.SeedTeam(ctx, app, []string{"admin@example.com"})
			if err == nil {
				t.Error("SeedTeam() expected error for missing args, got nil")
			}
		})
	})

	t.Run("error on invalid email", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			app := core.NewTestBaseApp(conf.ZeroEnvConfig(), db)
			err := cmd.SeedTeam(ctx, app, []string{"not-an-email", "my-team"})
			if err == nil {
				t.Error("SeedTeam() expected error for invalid email, got nil")
			}
		})
	})

	t.Run("error when user does not exist", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			app := core.NewTestBaseApp(conf.ZeroEnvConfig(), db)
			err := cmd.SeedTeam(ctx, app, []string{"nobody@example.com", "my-team"})
			if err == nil {
				t.Error("SeedTeam() expected error for missing user, got nil")
			}
		})
	})

	t.Run("creates team with owner", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			repository.CreateRolesAndPermissions(t, ctx, db, shared.KnownRoleNamesPermissionsMap)
			app := core.NewTestBaseApp(conf.ZeroEnvConfig(), db)

			const email = "owner@example.com"
			if _, err := app.Auth().Signup(ctx, &auth.SignupInput{
				Email:    email,
				Password: "Password123!",
				Verified: true,
			}); err != nil {
				t.Fatalf("Signup() error = %v", err)
			}

			if err := cmd.SeedTeam(ctx, app, []string{email, "my-team"}); err != nil {
				t.Fatalf("SeedTeam() error = %v", err)
			}

			user, err := app.Adapter().User().FindUser(ctx, &stores.UserFilter{
				Emails: []string{email},
			})
			if err != nil {
				t.Fatalf("FindUser() error = %v", err)
			}
			count := repository.MustCountAllCtx(t, ctx, repository.Team, db, nil)
			if count != 1 {
				t.Errorf("expected 1 team, got %d", count)
			}
			memberCount := repository.MustCountAllCtx(t, ctx, repository.TeamMember, db, &map[string]any{
				"user_id": map[string]any{"_eq": user.ID},
			})
			if memberCount != 1 {
				t.Errorf("expected 1 team member for owner, got %d", memberCount)
			}
		})
	})
}
