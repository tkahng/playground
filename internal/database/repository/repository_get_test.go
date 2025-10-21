package repository

import (
	"context"
	"testing"

	"github.com/tkahng/playground/internal/database"
)

type GetScenario[T any] struct {
	// Name is the name of the scenario
	Name string
	// Dbx is the database connection
	Dbx database.Dbx
	// Repo is the repository to be tested
	Repo Repository[T]
	// Args is the arguments to be passed to the repository
	Args []QueryOptionFunc
	// ArgsFunc returns the arguments to be passed to the repository. this is for arguments that need some computation.
	ArgsFunc func(t testing.TB, ctx context.Context, scenario *GetScenario[T]) []QueryOptionFunc
	// SetupFunc is the function to setup the test
	SetupFunc func(t testing.TB, ctx context.Context, scenario *GetScenario[T])
	// TestFunc is the function to verify the post result
	TestFunc func(t testing.TB, ctx context.Context, scenario *GetScenario[T], res []*T)
	// ㅉantErr indicates whether the test should expect an error
	WantErr bool
	// CauseErr enables manual transaction failure for testing
	CauseErr error
}

// PutTestScenarioFunc runs a single test scenario
func GetTestScenarioFunc[T any](t testing.TB, ctx context.Context, scenario *GetScenario[T]) {
	t.Helper()
	if scenario.SetupFunc != nil {
		scenario.SetupFunc(t, ctx, scenario)
	}
	dbx := scenario.Dbx
	repo := scenario.Repo
	var args []QueryOptionFunc
	args = scenario.Args
	if scenario.ArgsFunc != nil {
		args = scenario.ArgsFunc(t, ctx, scenario)
	}
	var txRes, err = repo.GetWithOptions(ctx, dbx, args...)
	if err != nil {
		if scenario.WantErr {
			return
		}
		t.Fatal(err)
	}
	scenario.TestFunc(t, ctx, scenario, txRes)
}

// func TestPostgresRepository_GetWithOptions(t *testing.T) {
// 	scenarios := []*GetScenario[models.RolePermission]{
// 		{
// 			Name: "",
// 			Dbx:  nil,
// 			Repo: RolePermission,
// 			Args: nil,
// 			TestFunc: func(t testing.TB, ctx context.Context, scenario *GetScenario[models.RolePermission], res []*models.RolePermission) {
// 				panic("TODO")
// 			},
// 			WantErr:  false,
// 			CauseErr: nil,
// 		},
// 	}
// 	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
// 		CreateRolesAndPermissions(t, ctx, db, knownRoleNamesPermissionsMap)
// 		for _, scenario := range scenarios {
// 			t.Run(scenario.Name, func(t *testing.T) {
// 				scenario.Dbx = db
// 				GetTestScenarioFunc(t, ctx, scenario)
// 			})
// 		}
// 	})
// }
