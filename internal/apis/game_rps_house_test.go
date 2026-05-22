//go:build integration

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
	"github.com/tkahng/playground/internal/test"
)

// mustSeedHouse seeds the house player for API tests (startup seed not called in test env).
func mustSeedHouse(t testing.TB, ctx context.Context, testApi *apis.TestApi) {
	t.Helper()
	if err := services.SeedHousePlayer(ctx, testApi.App.Adapter()); err != nil {
		t.Fatalf("SeedHousePlayer: %v", err)
	}
}

func TestApi_GetPlayers_ExcludesHousePlayer(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		mustSeedHouse(t, ctx, testApi)

		p1 := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		header := core.CreateTokenHeader(t, testApi.App, p1.Email)

		// Verify house email is never returned regardless of what other players exist.
		scenario := apis.ApiScenario{
			Name:           "GET /players excludes house player",
			Method:         http.MethodGet,
			URL:            "/players",
			ExpectedStatus: http.StatusOK,
			Headers:        []string{header},
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
				var body apis.ApiPaginatedResponse[*apis.Player]
				if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
					t.Fatalf("decode: %v", err)
				}
				for _, p := range body.Data {
					if p.Email == services.HousePlayerEmail {
						t.Errorf("GET /players returned house player — it must be excluded")
					}
				}
			},
		}
		scenario.Test(t)
	})
}

func TestApi_GetPlayersByEmail_ExcludesHousePlayer(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		mustSeedHouse(t, ctx, testApi)

		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		header := core.CreateTokenHeader(t, testApi.App, requester.Email)

		scenario := apis.ApiScenario{
			Name:           "email search never returns house player",
			Method:         http.MethodGet,
			URL:            "/players/registered/email/" + services.HousePlayerEmail,
			ExpectedStatus: http.StatusOK,
			Headers:        []string{header},
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
				var body apis.ApiSingleResponse[*apis.Player]
				if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if body.Data != nil {
					t.Errorf("email search returned house player — it should be excluded")
				}
			},
		}
		scenario.Test(t)
	})
}

func TestApi_ChallengeHouse_Success(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		mustSeedHouse(t, ctx, testApi)

		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		header := core.CreateTokenHeader(t, testApi.App, player.Email)

		body, _ := json.Marshal(map[string]any{"move": "rock"})

		scenario := apis.ApiScenario{
			Name:           "POST /games/rps/house returns completed game",
			Method:         http.MethodPost,
			URL:            "/games/rps/house",
			ExpectedStatus: http.StatusOK,
			Headers:        []string{header},
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, s *apis.ApiScenario) {
				s.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
				var out apis.ApiSingleResponse[*apis.ChallengeHouseResponse]
				if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if out.Data == nil {
					t.Fatal("data is nil")
				}
				if out.Data.RpsGame == nil {
					t.Fatal("rps_game is nil")
				}
				if out.Data.RpsGame.Status != apis.RpsGameStatusCompleted {
					t.Errorf("game status = %v, want completed", out.Data.RpsGame.Status)
				}
				if out.Data.RequestingParticipant == nil || out.Data.InvitedParticipant == nil {
					t.Error("participants are nil")
				}
				if out.Data.CooldownEndsAt.IsZero() {
					t.Error("cooldown_ends_at is zero")
				}
			},
		}
		scenario.Test(t)
	})
}

func TestApi_ChallengeHouse_Cooldown_Enforced(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		mustSeedHouse(t, ctx, testApi)

		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		header := core.CreateTokenHeader(t, testApi.App, player.Email)

		body, _ := json.Marshal(map[string]any{"move": "rock"})

		// First call succeeds.
		first := apis.ApiScenario{
			Name:           "first challenge succeeds",
			Method:         http.MethodPost,
			URL:            "/games/rps/house",
			ExpectedStatus: http.StatusOK,
			Headers:        []string{header},
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, s *apis.ApiScenario) {
				s.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
				var out apis.ApiSingleResponse[*apis.ChallengeHouseResponse]
				if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
					t.Fatalf("decode first challenge: %v", err)
				}
				if out.Data == nil || out.Data.RpsGame == nil {
					t.Fatal("first challenge returned no game")
				}
			},
		}
		first.Test(t)

		// Immediate second call hits cooldown.
		second := apis.ApiScenario{
			Name:            "second immediate challenge blocked by cooldown",
			Method:          http.MethodPost,
			URL:             "/games/rps/house",
			ExpectedStatus:  http.StatusTooManyRequests,
			Headers:         []string{header},
			ExpectedContent: []string{"cooldown"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, s *apis.ApiScenario) {
				s.Body = strings.NewReader(string(body))
			},
		}
		second.Test(t)
	})
}

func TestApi_ChallengeHouse_Disabled_Returns403(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		mustSeedHouse(t, ctx, testApi)

		// Disable the house player.
		house, _ := testApi.App.Adapter().Gaming().FindHousePlayer(ctx)
		disabledMeta, _ := services.SetHouseEnabled(house.Metadata, false)
		house.Metadata = disabledMeta
		if _, err := testApi.App.Adapter().Gaming().UpdatePlayer(ctx, house); err != nil {
			t.Fatalf("UpdatePlayer: %v", err)
		}

		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		header := core.CreateTokenHeader(t, testApi.App, player.Email)

		body, _ := json.Marshal(map[string]any{"move": "rock"})

		scenario := apis.ApiScenario{
			Name:            "challenge disabled house returns 403",
			Method:          http.MethodPost,
			URL:             "/games/rps/house",
			ExpectedStatus:  http.StatusForbidden,
			Headers:         []string{header},
			ExpectedContent: []string{"disabled"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, s *apis.ApiScenario) {
				s.Body = strings.NewReader(string(body))
			},
		}
		scenario.Test(t)
	})
}

func TestApi_ChallengeHouse_Unauthenticated_Returns401(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		mustSeedHouse(t, ctx, testApi)

		body, _ := json.Marshal(map[string]any{"move": "rock"})

		scenario := apis.ApiScenario{
			Name:            "unauthenticated challenge returns 401",
			Method:          http.MethodPost,
			URL:             "/games/rps/house",
			ExpectedStatus:  http.StatusUnauthorized,
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			ExpectedContent: []string{"Unauthorized"},
			BeforeTestFunc: func(t testing.TB, _ *core.BaseApp, s *apis.ApiScenario) {
				s.Body = strings.NewReader(string(body))
			},
		}
		scenario.Test(t)
	})
}
