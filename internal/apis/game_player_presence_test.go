package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/sse"
)

// presenceBody is the subset of the response body we care about in these tests.
type presenceBody struct {
	PlayerID    string  `json:"player_id"`
	IsConnected bool    `json:"is_connected"`
	LastSeenAt  *string `json:"last_seen_at"`
}

// startTestSseManager starts the SSE manager's Run loop for the duration of
// the test. Call once per test that needs SSE registration.
func startTestSseManager(t testing.TB, app *core.BaseApp) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	go app.SseManager().Run(ctx)
	t.Cleanup(cancel)
}

// registerTestSseClient adds a fake SSE client for playerID to the app's SSE
// manager and returns a cleanup func that deregisters it.
// startTestSseManager must be called before this.
func registerTestSseClient(t testing.TB, app *core.BaseApp, playerID string) func() {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	// noopSend discards every message — we only care about the registration.
	noopSend := func(_ any) error { return nil }
	client := sse.NewClient(sse.PlayerChannel(playerID), noopSend, nil, nil)
	app.SseManager().RegisterClient(ctx, cancel, client)
	return func() {
		cancel()
		app.SseManager().UnregisterClient(client)
	}
}

// --- GET /players/{player-id}/online-status ---

func Test_GetPlayerPresence_IsConnected_WhenSseClientRegistered(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		startTestSseManager(t, testApi.App)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// Register a live SSE client for the target player.
		cleanup := registerTestSseClient(t, testApi.App, target.ID.String())
		defer cleanup()

		scenario := &apis.ApiScenario{
			Name:           "connected",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/online-status", target.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, requester.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[presenceBody](t, res.Body.Bytes())
				assert.True(t, result.IsConnected)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetPlayerPresence_NotConnected_WhenNoSseClient(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:           "not connected",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/online-status", target.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, requester.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[presenceBody](t, res.Body.Bytes())
				assert.False(t, result.IsConnected)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetPlayerPresence_NotConnected_AfterClientDeregistered(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		startTestSseManager(t, testApi.App)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		cleanup := registerTestSseClient(t, testApi.App, target.ID.String())
		cleanup() // deregister immediately before the request

		scenario := &apis.ApiScenario{
			Name:           "deregistered",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/online-status", target.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, requester.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[presenceBody](t, res.Body.Bytes())
				assert.False(t, result.IsConnected)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetPlayerPresence_LastSeenAt_NilWhenNeverSeen(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:           "never seen",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/online-status", target.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, requester.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[presenceBody](t, res.Body.Bytes())
				assert.Nil(t, result.LastSeenAt)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetPlayerPresence_LastSeenAt_SetAfterActivity(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		require.NoError(t, testApi.App.Adapter().Gaming().UpdatePlayerLastSeen(ctx, target.ID))

		scenario := &apis.ApiScenario{
			Name:           "has last_seen_at",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/online-status", target.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, requester.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[presenceBody](t, res.Body.Bytes())
				assert.NotNil(t, result.LastSeenAt)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetPlayerPresence_Returns404ForUnknownPlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:            "not found",
			Method:          http.MethodGet,
			URL:             "/players/00000000-0000-0000-0000-000000000001/online-status",
			ExpectedStatus:  http.StatusNotFound,
			ExpectedContent: []string{"Not Found"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, requester.Email)
				scenario.Headers = []string{header}
			},
		}
		scenario.Test(t)
	})
}

func Test_GetPlayerPresence_Requires401WithoutAuth(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:            "unauthenticated",
			Method:          http.MethodGet,
			URL:             fmt.Sprintf("/players/%s/online-status", target.ID),
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"Unauthorized"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
		}
		scenario.Test(t)
	})
}
