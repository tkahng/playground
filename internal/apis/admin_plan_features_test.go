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
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_AdminPlanFeaturesList(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("admin@k2dv.io"),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)

		// Seed a product and a plan features row
		product := &models.StripeProduct{ID: "prod_pf_list_test", Active: true, Name: "Pro", Metadata: map[string]string{}}
		if err := testApi.App.Adapter().Product().UpsertProduct(ctx, product); err != nil {
			t.Fatalf("UpsertProduct() error = %v", err)
		}
		if _, err := testApi.App.Adapter().PlanFeatures().Upsert(ctx, &models.PlanFeatures{
			StripeProductID: product.ID,
			DailyAiTokens:   50_000,
		}); err != nil {
			t.Fatalf("PlanFeatures.Upsert() error = %v", err)
		}

		tests := []apis.ApiScenario{
			{
				Name:           "list plan features - ok",
				Method:         http.MethodGet,
				URL:            "/admin/plan-features",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.PlanFeature]
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if len(body.Data) == 0 {
						t.Error("expected at least one plan feature row")
					}
				},
			},
			{
				Name:            "list plan features - unauthorized",
				Method:          http.MethodGet,
				URL:             "/admin/plan-features",
				ExpectedStatus:  http.StatusUnauthorized,
				Headers:         []string{},
				ExpectedContent: []string{"Unauthorized"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_AdminPlanFeaturesGet(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("admin@k2dv.io"),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)

		product := &models.StripeProduct{ID: "prod_pf_get_test", Active: true, Name: "Pro", Metadata: map[string]string{}}
		if err := testApi.App.Adapter().Product().UpsertProduct(ctx, product); err != nil {
			t.Fatalf("UpsertProduct() error = %v", err)
		}
		if _, err := testApi.App.Adapter().PlanFeatures().Upsert(ctx, &models.PlanFeatures{
			StripeProductID: product.ID,
			DailyAiTokens:   75_000,
		}); err != nil {
			t.Fatalf("PlanFeatures.Upsert() error = %v", err)
		}

		tests := []apis.ApiScenario{
			{
				Name:           "get plan features - found",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/admin/plan-features/%s", product.ID),
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.PlanFeature
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if body.StripeProductID != product.ID {
						t.Errorf("StripeProductID = %q, want %q", body.StripeProductID, product.ID)
					}
					if body.DailyAiTokens != 75_000 {
						t.Errorf("DailyAiTokens = %d, want 75000", body.DailyAiTokens)
					}
				},
			},
			{
				Name:            "get plan features - not found",
				Method:          http.MethodGet,
				URL:             "/admin/plan-features/prod_does_not_exist",
				ExpectedStatus:  http.StatusNotFound,
				Headers:         []string{header},
				ExpectedContent: []string{"Not Found"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_AdminPlanFeaturesUpsert(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(
			t,
			testApi.App,
			core.UserWithEmail("admin@k2dv.io"),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)

		product := &models.StripeProduct{ID: "prod_pf_upsert_test", Active: true, Name: "Pro", Metadata: map[string]string{}}
		if err := testApi.App.Adapter().Product().UpsertProduct(ctx, product); err != nil {
			t.Fatalf("UpsertProduct() error = %v", err)
		}

		tests := []apis.ApiScenario{
			{
				Name:           "upsert plan features - creates new",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/admin/plan-features/%s", product.ID),
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					scenario.Body = apis.JsonToReader(t, &apis.PlanFeaturesUpsertBody{DailyAiTokens: 100_000})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.PlanFeature
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if body.DailyAiTokens != 100_000 {
						t.Errorf("DailyAiTokens = %d, want 100000", body.DailyAiTokens)
					}
					if body.StripeProductID != product.ID {
						t.Errorf("StripeProductID = %q, want %q", body.StripeProductID, product.ID)
					}
				},
			},
			{
				Name:           "upsert plan features - updates existing",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/admin/plan-features/%s", product.ID),
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					scenario.Body = apis.JsonToReader(t, &apis.PlanFeaturesUpsertBody{DailyAiTokens: 500_000})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.PlanFeature
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if body.DailyAiTokens != 500_000 {
						t.Errorf("DailyAiTokens after update = %d, want 500000", body.DailyAiTokens)
					}
				},
			},
			{
				Name:           "upsert plan features - arbitrary product id succeeds",
				Method:         http.MethodPut,
				URL:            "/admin/plan-features/prod_nonexistent",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					scenario.Body = apis.JsonToReader(t, &apis.PlanFeaturesUpsertBody{DailyAiTokens: 10_000})
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.PlanFeature
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if body.StripeProductID != "prod_nonexistent" {
						t.Errorf("StripeProductID = %q, want %q", body.StripeProductID, "prod_nonexistent")
					}
					if body.DailyAiTokens != 10_000 {
						t.Errorf("DailyAiTokens = %d, want 10000", body.DailyAiTokens)
					}
				},
			},
			{
				Name:            "upsert plan features - unauthorized",
				Method:          http.MethodPut,
				URL:             fmt.Sprintf("/admin/plan-features/%s", product.ID),
				ExpectedStatus:  http.StatusUnauthorized,
				Headers:         []string{},
				ExpectedContent: []string{"Unauthorized"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					scenario.Body = apis.JsonToReader(t, &apis.PlanFeaturesUpsertBody{DailyAiTokens: 10_000})
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
