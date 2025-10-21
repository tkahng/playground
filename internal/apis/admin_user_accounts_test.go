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
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		adminUser := CreateUserWithOptions(
			t,
			testApi.App,
			UserWithEmail("admin@k2dv.io"),
			UserWithPermission(shared.PermissionNameAdmin),
			UserWithProvider(models.ProvidersApple),
			UserWithProviderType(models.ProviderTypeOAuth),
		)
		header := createTokenHeader(t, testApi.App, adminUser.User.Email)
		_ = CreateUserWithOptions(
			t,
			testApi.App,
			UserWithEmail("user1@example.com"),
			UserWithProvider(models.ProvidersCredentials),
			UserWithProviderType(models.ProviderTypeCredentials),
		)
		_ = CreateUserWithOptions(
			t,
			testApi.App,
			UserWithEmail("user2@example.com"),
			UserWithProvider(models.ProvidersGoogle),
			UserWithProviderType(models.ProviderTypeOAuth),
		)
		tests := []ApiScenario{
			{
				Name:           "admin user accounts list all",
				Method:         http.MethodGet,
				URL:            "/admin/user-accounts",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
