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
	"github.com/tkahng/playground/internal/shared"
)

type testCase struct {
	name string // description of this test case
	// Named input parameters for target function.
	app           *core.BaseApp
	args          []string
	wantErr       bool
	afterTestFunc func(t testing.TB, app *core.BaseApp)
}

func TestCreateSuperuser(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		app := core.NewTestBaseApp(conf.ZeroEnvConfig(), db)
		repository.CreateRolesAndPermissions(t, ctx, db, shared.KnownRoleNamesPermissionsMap)
		tests := []testCase{
			{
				name: "should create superuser",
				app:  app,
				args: []string{
					"admin@k2dv.io",
					"Password123!",
				},
				wantErr: false,
				afterTestFunc: func(t testing.TB, app *core.BaseApp) {
					userInfo, err := app.Auth().Signin(ctx, &auth.SigninInput{
						Email:    "admin@k2dv.io",
						Password: "Password123!",
					})
					if err != nil {
						t.Errorf("Error getting user: %v", err)
					}
					if userInfo.User.EmailVerifiedAt == nil {
						t.Error("email not verified")
					}
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				gotErr := cmd.CreateSuperuser(context.Background(), tt.app, tt.args)
				if gotErr != nil {
					if !tt.wantErr {
						t.Errorf("CreateSuperuser() failed: %v", gotErr)
					}
				}
				if tt.afterTestFunc != nil {
					tt.afterTestFunc(t, tt.app)
				}
			})
		}
	})
}
