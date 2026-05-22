//go:build integration

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
)

func TestApi_IssueSSETicket(t *testing.T) {
	type ticketBody struct {
		Ticket string `json:"ticket"`
	}

	tests := []apis.ApiScenario{
		{
			Name:           "owner receives a valid ticket",
			Method:         http.MethodPost,
			URL:            "/team-members/{team-member-id}/sse/ticket",
			ExpectedStatus: http.StatusOK,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				team := CreateTeamAndOwner(t, app)
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
				sc.Headers = []string{header}
				sc.URL = fmt.Sprintf("/team-members/%s/sse/ticket", team.Member.ID)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario, res *httptest.ResponseRecorder) {
				body := test.MustUnMarshal[ticketBody](t, res.Body.Bytes())
				require.NotEmpty(t, body.Ticket, "response must contain a non-empty ticket")
				_, _, ok := app.SseTickets().Validate(body.Ticket)
				assert.True(t, ok, "issued ticket must be immediately valid in the store")
			},
		},
		{
			Name:            "unauthenticated request is rejected",
			Method:          http.MethodPost,
			URL:             "/team-members/{team-member-id}/sse/ticket",
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"you are not authenticated"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				team := CreateTeamAndOwner(t, app)
				sc.URL = fmt.Sprintf("/team-members/%s/sse/ticket", team.Member.ID)
				// no Authorization header
			},
		},
		{
			Name:            "cross-member ticket request is rejected",
			Method:          http.MethodPost,
			URL:             "/team-members/{team-member-id}/sse/ticket",
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"team info not found"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				team1 := CreateTeamAndOwner(t, app)
				team2 := CreateTeamAndOwner(t, app)
				// team2's user tries to issue a ticket for team1's member
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team2.User.Email)
				sc.Headers = []string{header}
				sc.URL = fmt.Sprintf("/team-members/%s/sse/ticket", team1.Member.ID)
			},
		},
	}

	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *apis.TestApi { return testApi }
			tt.Test(t)
		})
	}
}

func TestApi_SSETicketAuth_RejectsInvalidTicket(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:            "invalid ticket string returns 401",
			Method:          http.MethodGet,
			URL:             "/team-members/{team-member-id}/sse?ticket=bogus-ticket",
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"invalid or expired SSE ticket"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				team := CreateTeamAndOwner(t, app)
				sc.URL = fmt.Sprintf("/team-members/%s/sse?ticket=bogus-ticket", team.Member.ID)
				// no Authorization header — ticket path only
			},
		},
		{
			Name:            "ticket issued for member A is rejected on member B's SSE URL",
			Method:          http.MethodGet,
			URL:             "/team-members/{team-member-id}/sse",
			ExpectedStatus:  http.StatusUnauthorized,
			ExpectedContent: []string{"SSE ticket does not match"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				team1 := CreateTeamAndOwner(t, app)
				team2 := CreateTeamAndOwner(t, app)

				// Issue a ticket scoped to team1's member
				tok := app.SseTickets().Issue(team1.User.ID, team1.Member.ID)

				// Try to use it on team2's SSE URL
				sc.URL = fmt.Sprintf("/team-members/%s/sse?ticket=%s", team2.Member.ID, tok)
			},
		},
	}

	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *apis.TestApi { return testApi }
			tt.Test(t)
		})
	}
}
