package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/test"
)

var (
	knownRoleNames, knwonPermissionNames                     = []string{"superuser", "advanced", "pro", "basic"}, []string{"superuser", "advanced", "pro", "basic"}
	knownRoleNamesPermissionsMap         map[string][]string = map[string][]string{
		"basic":     {"basic"},
		"pro":       {"basic", "pro"},
		"advanced":  {"basic", "pro", "advanced"},
		"superuser": {"basic", "pro", "advanced", "superuser"},
	}
)

func TestApi_AdminRolesList(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		repository.CreateRolesAndPermissions(t, ctx, db, knownRoleNamesPermissionsMap)
		testApi := SetupApi(t, ctx, db)
		adminUser := CreateUserWithOptions(
			t,
			testApi.App,
			UserWithEmail("admin@k2dv.io"),
			UserWithRoles(shared.PermissionNameAdmin),
			UserWithProvider(models.ProvidersCredentials),
			UserWithProviderType(models.ProviderTypeCredentials),
		)
		header := createTokenHeader(t, testApi.App, adminUser.User.Email)
		tests := []ApiScenario{
			{
				Name:           "admin roles list all",
				Method:         http.MethodGet,
				URL:            "/admin/roles",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.Role]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != 4 {
						t.Errorf("Expected 4 roles, got %d", len(body.Data))
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
