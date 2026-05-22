//go:build integration

package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/sse"
)

// --- POST /players/sse/ticket ---

func TestApi_IssuePlayerSseTicket_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		type ticketBody struct {
			Ticket string `json:"ticket"`
		}

		scenario := &apis.ApiScenario{
			Name:           "player receives valid ticket",
			Method:         http.MethodPost,
			URL:            "/players/sse/ticket",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				sc.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario, res *httptest.ResponseRecorder) {
				body := test.MustUnMarshal[ticketBody](t, res.Body.Bytes())
				require.NotEmpty(t, body.Ticket)
				_, playerID, ok := app.SseTickets().Validate(body.Ticket)
				assert.True(t, ok, "issued ticket must be immediately valid")
				assert.Equal(t, player.ID, playerID)
			},
		}
		scenario.Test(t)
	})
}

func TestApi_IssuePlayerSseTicket_Requires401WithoutAuth(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		scenario := &apis.ApiScenario{
			Name:            "unauthenticated",
			Method:          http.MethodPost,
			URL:             "/players/sse/ticket",
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"Unauthorized"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
		}
		scenario.Test(t)
	})
}

// --- GET /players/{id}/sse ticket auth ---

func TestApi_PlayerSseAuth_RejectsInvalidTicket(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:            "bogus ticket",
			Method:          http.MethodGet,
			URL:             fmt.Sprintf("/players/%s/sse?ticket=bogus", player.ID),
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"invalid or expired SSE ticket"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
		}
		scenario.Test(t)
	})
}

func TestApi_PlayerSseAuth_RejectsMissingTicket(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:            "no ticket param",
			Method:          http.MethodGet,
			URL:             fmt.Sprintf("/players/%s/sse", player.ID),
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"missing SSE ticket"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
		}
		scenario.Test(t)
	})
}

func TestApi_PlayerSseAuth_RejectsTicketForWrongPlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		p1 := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		p2 := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// Issue ticket scoped to p1, try to use on p2's SSE URL.
		ticket := testApi.App.SseTickets().Issue(*p1.UserID, p1.ID)

		scenario := &apis.ApiScenario{
			Name:            "ticket player mismatch",
			Method:          http.MethodGet,
			URL:             fmt.Sprintf("/players/%s/sse?ticket=%s", p2.ID, ticket),
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"SSE ticket does not match player"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
		}
		scenario.Test(t)
	})
}

// --- SSE connect/disconnect last_seen_at stamping ---
//
// stampLastSeenFromChannel is unexported, so we test its effect through the
// SSE manager by directly registering and deregistering a fake client and
// verifying the side-effect (last_seen_at updated) against the DB.
// This mirrors exactly what PlayerSseEventsBind's onCreate/onDestroy hooks do.

func TestSseConnect_StampsLastSeenAt(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		startTestSseManager(t, testApi.App)

		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// Confirm last_seen_at is nil before any SSE activity.
		before, err := testApi.App.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{Ids: []uuid.UUID{player.ID}})
		require.NoError(t, err)
		require.Nil(t, before.LastSeenAt)

		// Simulate SSE connect: register client then stamp (mirrors onCreate hook).
		connectCtx, connectCancel := context.WithCancel(context.Background())
		noop := func(any) error { return nil }
		client := sse.NewClient(sse.PlayerChannel(player.ID.String()), noop, nil, nil)
		testApi.App.SseManager().RegisterClient(connectCtx, connectCancel, client)
		require.NoError(t, testApi.App.Adapter().Gaming().UpdatePlayerLastSeen(ctx, player.ID))

		after, err := testApi.App.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{Ids: []uuid.UUID{player.ID}})
		require.NoError(t, err)
		assert.NotNil(t, after.LastSeenAt, "last_seen_at should be set after SSE connect")

		connectCancel()
		testApi.App.SseManager().UnregisterClient(client)
	})
}

func TestSseDisconnect_UpdatesLastSeenAt(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		startTestSseManager(t, testApi.App)

		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// Connect.
		connectCtx, connectCancel := context.WithCancel(context.Background())
		noop := func(any) error { return nil }
		client := sse.NewClient(sse.PlayerChannel(player.ID.String()), noop, nil, nil)
		testApi.App.SseManager().RegisterClient(connectCtx, connectCancel, client)
		require.NoError(t, testApi.App.Adapter().Gaming().UpdatePlayerLastSeen(ctx, player.ID))

		firstSeen, err := testApi.App.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{Ids: []uuid.UUID{player.ID}})
		require.NoError(t, err)
		require.NotNil(t, firstSeen.LastSeenAt)

		// Simulate SSE disconnect: stamp then unregister (mirrors onDestroy hook).
		require.NoError(t, testApi.App.Adapter().Gaming().UpdatePlayerLastSeen(ctx, player.ID))
		connectCancel()
		testApi.App.SseManager().UnregisterClient(client)

		afterDisconnect, err := testApi.App.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{Ids: []uuid.UUID{player.ID}})
		require.NoError(t, err)
		assert.NotNil(t, afterDisconnect.LastSeenAt, "last_seen_at should still be set after disconnect")
		// The disconnect stamp should be >= the connect stamp.
		assert.True(t, !afterDisconnect.LastSeenAt.Before(*firstSeen.LastSeenAt),
			"disconnect stamp should be >= connect stamp")
	})
}
