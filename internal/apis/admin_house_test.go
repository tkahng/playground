package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/test"
)

func TestAdminHouseStats_OK(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		if err := services.SeedHousePlayer(ctx, testApi.App.Adapter()); err != nil {
			t.Fatalf("SeedHousePlayer: %v", err)
		}

		adminUser := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithEmail("admin_hs@example.com"),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)

		scenarios := []apis.ApiScenario{
			{
				Name:           "stats returns 200 with correct shape",
				Method:         http.MethodGet,
				URL:            "/admin/house/stats",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiSingleResponse[*apis.HouseStatsResponse]
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode: %v", err)
					}
					if body.Data == nil {
						t.Fatal("data is nil")
					}
					if body.Data.TotalGames != 0 {
						t.Errorf("TotalGames = %d, want 0 for fresh house", body.Data.TotalGames)
					}
					if !body.Data.Enabled {
						t.Error("Enabled = false, want true for freshly seeded house")
					}
				},
			},
			{
				Name:            "stats rejected without auth",
				Method:          http.MethodGet,
				URL:             "/admin/house/stats",
				ExpectedStatus:  http.StatusForbidden,
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				ExpectedContent: []string{"Forbidden"},
			},
		}
		for _, s := range scenarios {
			s.Test(t)
		}
	})
}

func TestAdminHouseToggle_EnableDisable(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		if err := services.SeedHousePlayer(ctx, testApi.App.Adapter()); err != nil {
			t.Fatalf("SeedHousePlayer: %v", err)
		}

		adminUser := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithEmail("admin_ht@example.com"),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)

		disableBody, _ := json.Marshal(map[string]any{"enabled": false})
		enableBody, _ := json.Marshal(map[string]any{"enabled": true})

		scenarios := []apis.ApiScenario{
			{
				Name:           "disable house returns enabled=false",
				Method:         http.MethodPut,
				URL:            "/admin/house/enabled",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, s *apis.ApiScenario) {
					s.Body = strings.NewReader(string(disableBody))
				},
				AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiSingleResponse[*apis.HouseStatsResponse]
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode: %v", err)
					}
					if body.Data.Enabled {
						t.Error("Enabled = true after disable, want false")
					}
				},
			},
			{
				Name:           "re-enable house returns enabled=true",
				Method:         http.MethodPut,
				URL:            "/admin/house/enabled",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, s *apis.ApiScenario) {
					s.Body = strings.NewReader(string(enableBody))
				},
				AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiSingleResponse[*apis.HouseStatsResponse]
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode: %v", err)
					}
					if !body.Data.Enabled {
						t.Error("Enabled = false after re-enable, want true")
					}
				},
			},
			{
				Name:            "toggle rejected without auth",
				Method:          http.MethodPut,
				URL:             "/admin/house/enabled",
				ExpectedStatus:  http.StatusForbidden,
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				ExpectedContent: []string{"Forbidden"},
			},
		}
		for _, s := range scenarios {
			s.Test(t)
		}
	})
}
