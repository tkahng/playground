package apis_test

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestApi_AdminPermissionsList(t *testing.T) {
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

		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		tests := []apis.ApiScenario{
			{
				Name:           "admin permission list get by basic role id reversed",
				Method:         http.MethodGet,
				URL:            "/admin/permissions",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					role := repository.MustFindOneCtx(t, ctx, repository.Role, app.Db(), &map[string]any{
						"name": map[string]any{
							"_eq": shared.PermissionNameBasic,
						},
					})
					scenario.URL = fmt.Sprintf("/admin/permissions?role_id=%s&role_reverse=true", role.ID.String())
				},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var expectedCount int = 3
					var body apis.ApiPaginatedResponse[*apis.Permission]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d permissions, got %d", expectedCount, len(body.Data))
					}
				},
			},
			{
				Name:           "admin permission list get by advanced role id",
				Method:         http.MethodGet,
				URL:            "/admin/permissions",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					role := repository.MustFindOneCtx(t, ctx, repository.Role, app.Db(), &map[string]any{
						"name": map[string]any{
							"_eq": shared.PermissionNameAdvanced,
						},
					})
					scenario.URL = fmt.Sprintf("/admin/permissions?role_id=%s", role.ID.String())
				},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var expectedCount int = 3
					var body apis.ApiPaginatedResponse[*apis.Permission]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d permissions, got %d", expectedCount, len(body.Data))
					}
				},
			},
			{
				Name:           "admin permission list get by admin role id",
				Method:         http.MethodGet,
				URL:            "/admin/permissions",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					role := repository.MustFindOneCtx(t, ctx, repository.Role, app.Db(), &map[string]any{
						"name": map[string]any{
							"_eq": shared.PermissionNameAdmin,
						},
					})
					scenario.URL = fmt.Sprintf("/admin/permissions?role_id=%s", role.ID.String())
				},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var expectedCount int = 4
					var body apis.ApiPaginatedResponse[*apis.Permission]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d permissions, got %d", expectedCount, len(body.Data))
					}
				},
			},
			{
				Name:           "admin permission list get superuser, basic by name",
				Method:         http.MethodGet,
				URL:            "/admin/permissions?names=" + shared.PermissionNameAdmin + "," + shared.PermissionNameBasic,
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var expectedCount int = 2
					var body apis.ApiPaginatedResponse[*apis.Permission]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d permissions, got %d", expectedCount, len(body.Data))
					}
				},
			},
			{
				Name:           "admin permission list get superuser by name",
				Method:         http.MethodGet,
				URL:            "/admin/permissions?names=" + shared.PermissionNameAdmin,
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var expectedCount int = 1
					var body apis.ApiPaginatedResponse[*apis.Permission]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d permissions, got %d", expectedCount, len(body.Data))
					}
				},
			},
			{
				Name:           "admin permission list get named",
				Method:         http.MethodGet,
				URL:            "/admin/permissions?names=" + shared.PermissionNameAdmin,
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var expectedCount int = 1
					var body apis.ApiPaginatedResponse[*apis.Permission]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d permissions, got %d", expectedCount, len(body.Data))
					}
				},
			},
			{
				Name:           "admin permission list all",
				Method:         http.MethodGet,
				URL:            "/admin/permissions",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var expectedCount int = 4
					var body apis.ApiPaginatedResponse[*apis.Permission]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != expectedCount {
						t.Errorf("Expected %d permissions, got %d", expectedCount, len(body.Data))
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
