package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/test"
)

// --- IsPlayerOnline unit tests (no DB required) ---

func Test_IsPlayerOnline_NilLastSeen(t *testing.T) {
	assert.False(t, apis.IsPlayerOnline(nil))
}

func Test_IsPlayerOnline_JustUnderThreshold(t *testing.T) {
	recent := time.Now().Add(-1*time.Minute - 59*time.Second)
	assert.True(t, apis.IsPlayerOnline(&recent))
}

func Test_IsPlayerOnline_OneSecondBeforeThreshold(t *testing.T) {
	// 1 second inside the 2-minute window — must be online.
	at := time.Now().Add(-2*time.Minute + time.Second)
	assert.True(t, apis.IsPlayerOnline(&at))
}

func Test_IsPlayerOnline_JustOverThreshold(t *testing.T) {
	stale := time.Now().Add(-2*time.Minute - time.Second)
	assert.False(t, apis.IsPlayerOnline(&stale))
}

func Test_IsPlayerOnline_FutureTimestamp(t *testing.T) {
	// Edge case: last_seen in the future (clock skew) — should still be online.
	future := time.Now().Add(time.Minute)
	assert.True(t, apis.IsPlayerOnline(&future))
}

// --- GET /players/{player-id}/online-status endpoint tests ---

func Test_GetPlayerOnlineStatus_ReturnsOfflineWhenNeverSeen(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:           "never seen — offline",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/online-status", target.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, requester.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				type onlineResp struct {
					IsOnline bool `json:"is_online"`
				}
				result := test.MustUnMarshal[onlineResp](t, res.Body.Bytes())
				assert.False(t, result.IsOnline)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetPlayerOnlineStatus_ReturnsOnlineAfterLastSeen(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// Simulate target being seen recently.
		err := testApi.App.Adapter().Gaming().UpdatePlayerLastSeen(ctx, target.ID)
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "recently seen — online",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/online-status", target.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, requester.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				type onlineResp struct {
					IsOnline bool `json:"is_online"`
				}
				result := test.MustUnMarshal[onlineResp](t, res.Body.Bytes())
				assert.True(t, result.IsOnline)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetPlayerOnlineStatus_Returns404ForUnknownPlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:            "player not found",
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

func Test_GetPlayerOnlineStatus_Requires401WithoutAuth(t *testing.T) {
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
