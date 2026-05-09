package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_AdminStripeSyncProductsAndPrices(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("sync-admin@k2dv.io"),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		regularUser := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("sync-user@k2dv.io"),
		)
		adminHeader := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		regularHeader := core.CreateTokenHeader(t, testApi.App, regularUser.User.Email)

		tests := []apis.ApiScenario{
			{
				Name:           "sync - ok as superuser",
				Method:         http.MethodPost,
				URL:            "/admin/stripe/sync",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{adminHeader},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.AdminStripeSyncOutput
					if err := json.NewDecoder(res.Body).Decode(&body.Body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if !body.Body.ProductsSynced {
						t.Error("expected products_synced=true")
					}
					if !body.Body.PricesSynced {
						t.Error("expected prices_synced=true")
					}
					if body.Body.Message == "" {
						t.Error("expected non-empty message")
					}
				},
			},
			{
				Name:            "sync - unauthorized without token",
				Method:          http.MethodPost,
				URL:             "/admin/stripe/sync",
				ExpectedStatus:  http.StatusUnauthorized,
				ExpectedContent: []string{"Unauthorized"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			},
			{
				Name:            "sync - forbidden for regular user",
				Method:          http.MethodPost,
				URL:             "/admin/stripe/sync",
				ExpectedStatus:  http.StatusForbidden,
				Headers:         []string{regularHeader},
				ExpectedContent: []string{"Forbidden"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_AdminStripeSyncProductsAndPrices_ClientError(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("sync-admin-err@k2dv.io"),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		adminHeader := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)

		mockClient := core.ExtractTestPaymentClient(t, testApi.App)
		mockClient.FindAllProductsFunc = func() ([]*stripe.Product, error) {
			return nil, &stripe.Error{Msg: "stripe is down"}
		}

		tests := []apis.ApiScenario{
			{
				Name:            "sync - 500 when stripe client fails",
				Method:          http.MethodPost,
				URL:             "/admin/stripe/sync",
				ExpectedStatus:  http.StatusInternalServerError,
				Headers:         []string{adminHeader},
				ExpectedContent: []string{"Failed to sync"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
