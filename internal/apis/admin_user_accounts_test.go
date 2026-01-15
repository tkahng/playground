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
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_AdminUserAccounts(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("admin@k2dv.io"),
			core.UserWithPermission(shared.PermissionNameAdmin),
			core.UserWithProvider(models.ProvidersApple),
			core.UserWithProviderType(models.ProviderTypeOAuth),
		)
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		_ = core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("user1@example.com"),
			core.UserWithProvider(models.ProvidersCredentials),
			core.UserWithProviderType(models.ProviderTypeCredentials),
		)
		_ = core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("user2@example.com"),
			core.UserWithProvider(models.ProvidersGoogle),
			core.UserWithProviderType(models.ProviderTypeOAuth),
		)
		tests := []apis.ApiScenario{
			{
				Name:           "admin user accounts list all",
				Method:         http.MethodGet,
				URL:            "/admin/user-accounts",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.UserAccountOutput]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != 3 {
						t.Errorf("Expected 3 user, got %d", len(body.Data))
					}
				},
			},
			{
				Name:           "admin user accounts list filter provider google",
				Method:         http.MethodGet,
				URL:            "/admin/user-accounts?providers=google",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.UserAccountOutput]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != 1 {
						t.Errorf("Expected 1 user, got %d", len(body.Data))
					}
					if body.Data[0].Provider != models.ProvidersGoogle {
						t.Errorf("Expected account provider to be google, got %s", body.Data[0].Provider)
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
