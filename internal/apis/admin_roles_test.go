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

func TestApi_AdminRolesList(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		repository.CreateRolesAndPermissions(t, ctx, db, shared.KnownRoleNamesPermissionsMap)
		testApi := SetupApi(t, ctx, db)
		adminUser := CreateUserWithOptions(
			t,
			testApi.App,
			UserWithEmail("admin@k2dv.io"),
			UserWithRoles(shared.PermissionNameAdmin),
			UserWithProvider(models.ProvidersCredentials),
			UserWithProviderType(models.ProviderTypeCredentials),
		)

		basicUser := CreateUserWithOptions(
			t,
			testApi.App,
			UserWithEmail("basic@example.com"),
			UserWithRoles(shared.PermissionNameBasic),
			UserWithProvider(models.ProvidersCredentials),
			UserWithProviderType(models.ProviderTypeCredentials),
		)
		doubleRoleUser := CreateUserWithOptions(
			t,
			testApi.App,
			UserWithEmail("double@example.com"),
			UserWithRoles(shared.PermissionNameBasic, shared.PermissionNamePro),
			UserWithProvider(models.ProvidersCredentials),
			UserWithProviderType(models.ProviderTypeCredentials),
		)
		header := CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		tests := []ApiScenario{
			{
				Name:           "admin roles list get by user_id, pro and basic, reversed",
				Method:         http.MethodGet,
				URL:            "/admin/roles?user_id=" + doubleRoleUser.User.ID.String() + "&reverse=true",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.Role]
					var expectedCount int = 2
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d roles, got %d", expectedCount, len(body.Data))
					}
				},
			},
			{
				Name:           "admin roles list get by user_id, basic reversed",
				Method:         http.MethodGet,
				URL:            "/admin/roles?user_id=" + basicUser.User.ID.String() + "&reverse=true",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.Role]
					var expectedCount int = 3
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d roles, got %d", expectedCount, len(body.Data))
					}
				},
			},
			{
				Name:           "admin roles list get by user_id, basic",
				Method:         http.MethodGet,
				URL:            "/admin/roles?user_id=" + basicUser.User.ID.String(),
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.Role]
					var expectedCount int = 1
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d roles, got %d", expectedCount, len(body.Data))
					}
				},
			},
			{
				Name:           "admin roles list get superuser, basic by name",
				Method:         http.MethodGet,
				URL:            "/admin/roles?names=" + shared.PermissionNamePro + "," + shared.PermissionNameAdvanced,
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.Role]
					var expectedCount int = 2
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d roles, got %d", expectedCount, len(body.Data))
					}
				},
			},
			{
				Name:           "admin roles list get superuser by name",
				Method:         http.MethodGet,
				URL:            "/admin/roles?names=" + shared.PermissionNameAdmin,
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.Role]
					var expectedCount int = 1
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d roles, got %d", expectedCount, len(body.Data))
					}
				},
			},
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
					var expectedCount int = 4
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d roles, got %d", expectedCount, len(body.Data))
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
