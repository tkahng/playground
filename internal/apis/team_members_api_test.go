package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/test"
)

// find-team-team-member-by-id
func TestApi_FindTeamMemberByID(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		team1 := CreateTeamAndOwner(t, testApi.App)
		team1Member1 := CreateTeamMember(t, testApi.App, &team1.Team)
		team1Member2 := CreateTeamMember(t, testApi.App, &team1.Team, core.TeamWithActive(false))
		assert.False(t, team1Member2.Member.Active)
		team2 := CreateTeamAndOwner(t, testApi.App)
		// testMailer := ExtractTestMailer(t, testApi.App)
		tests := []ApiScenario{
			{
				Name:           "fail: team2owner find team1member1",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/team-members/{team-member-id}",
				ExpectedStatus: http.StatusUnprocessableEntity,
				ExpectedContent: []string{
					"team member's team_id does not match team_id in path",
				},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team2.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/team-members/%s", team2.Team.ID.String(), team1Member1.Member.ID.String())
				},
			},
			{
				Name:           "failed: inactive team1member2 find team1owner",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/team-members/{team-member-id}",
				ExpectedStatus: http.StatusUnauthorized,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				ExpectedContent: []string{
					"team info not found",
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1Member2.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/team-members/%s", team1Member2.Team.ID.String(), team1.Member.ID.String())
				},
			},
			{
				Name:           "success: team1member1 find team1owner",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/team-members/{team-member-id}",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1Member1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/team-members/%s", team1Member1.Team.ID.String(), team1.Member.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					teamMember := test.MustUnMarshal[apis.TeamMember](t, res.Body.Bytes())
					assert.Equal(t, team1.Member.ID, teamMember.ID)
					assert.Equal(t, team1.Member.TeamID, teamMember.TeamID)
					assert.Equal(t, team1.Member.UserID, teamMember.UserID)
					assert.Equal(t, string(team1.Member.Role), string(teamMember.Role))
					assert.Equal(t, team1.Member.Active, teamMember.Active)
				},
			},
			{
				Name:           "success: owner find member 1",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/team-members/{team-member-id}",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/team-members/%s", team1.Team.ID.String(), team1Member1.Member.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					teamMember := test.MustUnMarshal[apis.TeamMember](t, res.Body.Bytes())
					assert.Equal(t, team1Member1.Member.ID, teamMember.ID)
					assert.Equal(t, team1Member1.Member.TeamID, teamMember.TeamID)
					assert.Equal(t, team1Member1.Member.UserID, teamMember.UserID)
					assert.Equal(t, string(team1Member1.Member.Role), string(teamMember.Role))
					assert.Equal(t, team1Member1.Member.Active, teamMember.Active)
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})

}
