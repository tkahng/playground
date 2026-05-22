//go:build integration

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
		testApi := apis.SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("admin@k2dv.io"),
			core.UserWithRoles(shared.PermissionNameAdmin),
			core.UserWithProvider(models.ProvidersCredentials),
			core.UserWithProviderType(models.ProviderTypeCredentials),
		)

		basicUser := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("basic@example.com"),
			core.UserWithRoles(shared.PermissionNameBasic),
			core.UserWithProvider(models.ProvidersCredentials),
			core.UserWithProviderType(models.ProviderTypeCredentials),
		)
		doubleRoleUser := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("double@example.com"),
			core.UserWithRoles(shared.PermissionNameBasic, shared.PermissionNamePro),
			core.UserWithProvider(models.ProvidersCredentials),
			core.UserWithProviderType(models.ProviderTypeCredentials),
		)
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		tests := []apis.ApiScenario{
			{
				Name:           "admin roles list get by user_id, pro and basic, reversed",
				Method:         http.MethodGet,
				URL:            "/admin/roles?user_id=" + doubleRoleUser.User.ID.String() + "&reverse=true",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
