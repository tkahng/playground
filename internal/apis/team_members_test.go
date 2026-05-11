package apis_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

// find-team-team-member-by-id
func TestApi_FindTeamMemberByID(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team1 := CreateTeamAndOwner(t, testApi.App)
		team1Member1 := CreateTeamMember(t, testApi.App, &team1.Team)
		team1Member2 := CreateTeamMember(t, testApi.App, &team1.Team, core.TeamWithActive(false))
		assert.False(t, team1Member2.Member.Active)
		team2 := CreateTeamAndOwner(t, testApi.App)
		// testMailer := ExtractTestMailer(t, testApi.App)
		tests := []apis.ApiScenario{
			{
				Name:           "fail: team2owner find team1member1",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/team-members/{team-member-id}",
				ExpectedStatus: http.StatusUnprocessableEntity,
				ExpectedContent: []string{
					"team member's team_id does not match team_id in path",
				},
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team2.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/team-members/%s", team2.Team.ID.String(), team1Member1.Member.ID.String())
				},
			},
			{
				Name:           "failed: inactive team1member2 find team1owner",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/team-members/{team-member-id}",
				ExpectedStatus: http.StatusForbidden,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				ExpectedContent: []string{
					"team info not found",
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
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
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1Member1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/team-members/%s", team1Member1.Team.ID.String(), team1.Member.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/team-members/%s", team1.Team.ID.String(), team1Member1.Member.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
func TestApi_FindTeamTeamMembers(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team1 := CreateTeamAndOwner(t, testApi.App)
		team1Member1 := CreateTeamMember(t, testApi.App, &team1.Team)
		team1Member2 := CreateTeamMember(t, testApi.App, &team1.Team, core.TeamWithActive(false))
		team1Member3 := CreateTeamMember(t, testApi.App, &team1.Team, core.TeamWithRole(models.TeamMemberRoleGuest))
		// testMailer := ExtractTestMailer(t, testApi.App)
		tests := []apis.ApiScenario{
			{
				Name:           "success: team1 owner find everyone",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members", team1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					allTeamMemberIds := []uuid.UUID{team1.Member.ID, team1Member1.Member.ID, team1Member2.Member.ID, team1Member3.Member.ID}
					test.TestSliceEveryFunc(t, "something", result.Data, func(member *apis.TeamMember) bool {
						teamMatch := member.TeamID == team1.Team.ID
						isFound := slices.Contains(allTeamMemberIds, member.ID)
						return teamMatch && isFound
					})
				},
			},
			{
				Name:           "success: team1 owner find active",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members?active=true", team1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					allTeamMemberIds := []uuid.UUID{team1.Member.ID, team1Member1.Member.ID, team1Member3.Member.ID}
					test.TestSliceEveryFunc(t, "something", result.Data, func(member *apis.TeamMember) bool {
						teamMatch := member.TeamID == team1.Team.ID
						isFound := slices.Contains(allTeamMemberIds, member.ID)
						return teamMatch && isFound
					})
				},
			},
			{
				Name:           "success: team1 owner find active guest, member",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members?active=true&roles=guest,member", team1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					allTeamMemberIds := []uuid.UUID{team1Member3.Member.ID, team1Member1.Member.ID}
					test.TestSliceEveryFunc(t, "something", result.Data, func(member *apis.TeamMember) bool {
						teamMatch := member.TeamID == team1.Team.ID
						isFound := slices.Contains(allTeamMemberIds, member.ID)
						return teamMatch && isFound
					})
				},
			},
			{
				Name:           "success: team1member1 find everyone",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1Member1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members", team1Member1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					allTeamMemberIds := []uuid.UUID{team1.Member.ID, team1Member1.Member.ID, team1Member2.Member.ID, team1Member3.Member.ID}
					test.TestSliceEveryFunc(t, "something", result.Data, func(member *apis.TeamMember) bool {
						teamMatch := member.TeamID == team1.Team.ID
						isFound := slices.Contains(allTeamMemberIds, member.ID)
						return teamMatch && isFound
					})
				},
			},
			{
				Name:           "success: guest team1member3 find everyone",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1Member3.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members", team1Member3.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					allTeamMemberIds := []uuid.UUID{team1.Member.ID, team1Member1.Member.ID, team1Member2.Member.ID, team1Member3.Member.ID}
					test.TestSliceEveryFunc(t, "something", result.Data, func(member *apis.TeamMember) bool {
						teamMatch := member.TeamID == team1.Team.ID
						isFound := slices.Contains(allTeamMemberIds, member.ID)
						return teamMatch && isFound
					})
				},
			},
			{
				Name:           "failure: inactive team1member2 find everyone",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/members",
				ExpectedStatus: http.StatusForbidden,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				ExpectedContent: []string{
					"team info not found",
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1Member2.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members", team1Member2.Team.ID.String())
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
func TestApi_FindUserTeamMembers(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		adapter := testApi.App.Adapter()
		user1 := stores.CreateUser(adapter, ctx, "user1@example.com")
		user2 := stores.CreateUser(adapter, ctx, "user2@example.com")
		team1 := stores.CreateTeam(adapter, ctx, "Team1")
		team2 := stores.CreateTeam(adapter, ctx, "Team2")
		team3 := stores.CreateTeam(adapter, ctx, "Team3")
		user1Team1Member := stores.CreateTeamMember(adapter, ctx, team1, user1, models.TeamMemberRoleOwner, true)
		user1Team2Member := stores.CreateTeamMember(adapter, ctx, team2, user1, models.TeamMemberRoleMember, true)
		user1Team3Member := stores.CreateTeamMember(adapter, ctx, team3, user1, models.TeamMemberRoleGuest, true)
		user2Team1Member := stores.CreateTeamMember(adapter, ctx, team1, user2, models.TeamMemberRoleOwner, true)
		user2Team2Member := stores.CreateTeamMember(adapter, ctx, team2, user2, models.TeamMemberRoleMember, true)
		user2Team3Member := stores.CreateTeamMember(adapter, ctx, team3, user2, models.TeamMemberRoleGuest, true)

		user1Team1Member.User = user1
		user1Team1Member.Team = team3
		user1Team2Member.User = user1
		user1Team2Member.Team = team2
		user1Team3Member.User = user1
		user1Team3Member.Team = team1

		user2Team1Member.User = user2
		user2Team1Member.Team = team3
		user2Team2Member.User = user2
		user2Team2Member.Team = team2
		user2Team3Member.User = user2
		user2Team3Member.Team = team1

		err := adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, user1Team2Member.TeamID, *user1Team2Member.UserID)
		assert.NoError(t, err)
		err = adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, user1Team3Member.TeamID, *user1Team3Member.UserID)
		assert.NoError(t, err)
		err = adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, user1Team1Member.TeamID, *user1Team1Member.UserID)
		assert.NoError(t, err)
		// testMailer := ExtractTestMailer(t, testApi.App)
		tests := []apis.ApiScenario{
			{
				Name:           "user1 teamMembers by team name asc",
				Method:         http.MethodGet,
				URL:            "/team-members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, user1.Email)
					scenario.Headers = []string{tokenHeader}
					u, err := url.Parse("/team-members")
					assert.NoError(t, err)
					q := u.Query()
					q.Add("sort_by", "team.name")
					u.RawQuery = q.Encode()
					scenario.URL = u.RequestURI()
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					members := result.Data
					assert.Equal(t, 3, len(members))
					for _, member := range result.Data {
						assert.Equal(t, user1.ID, *member.UserID)
						assert.NotNil(t, member.User)
						assert.NotNil(t, member.Team)
					}
					test.TestSliceItemsOrderByFunc(t, members, func(a, b *apis.TeamMember) bool {
						return a.Team.Name < b.Team.Name
					})
				},
			},
			{
				Name:           "user1 teamMembers by team name desc",
				Method:         http.MethodGet,
				URL:            "/team-members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, user1.Email)
					scenario.Headers = []string{tokenHeader}
					u, err := url.Parse("/team-members")
					assert.NoError(t, err)
					q := u.Query()
					q.Add("sort_by", "team.name")
					q.Add("sort_order", "desc")
					u.RawQuery = q.Encode()
					scenario.URL = u.RequestURI()
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					members := result.Data
					assert.Equal(t, 3, len(members))
					for _, member := range result.Data {
						assert.Equal(t, user1.ID, *member.UserID)
						assert.NotNil(t, member.User)
						assert.NotNil(t, member.Team)
					}
					test.TestSliceItemsOrderByFunc(t, members, func(a, b *apis.TeamMember) bool {
						return a.Team.Name > b.Team.Name
					})
				},
			},
			{
				Name:           "user1 teamMembers by member last selected at asc",
				Method:         http.MethodGet,
				URL:            "/team-members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, user1.Email)
					scenario.Headers = []string{tokenHeader}
					u, err := url.Parse("/team-members")
					assert.NoError(t, err)
					q := u.Query()
					q.Add("sort_by", "last_selected_at")
					u.RawQuery = q.Encode()
					scenario.URL = u.RequestURI()
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					members := result.Data
					assert.Equal(t, 3, len(members))
					for _, member := range result.Data {
						assert.Equal(t, user1.ID, *member.UserID)
						assert.NotNil(t, member.User)
						assert.NotNil(t, member.Team)
					}
					test.TestSliceItemsOrderByFunc(t, members, func(a, b *apis.TeamMember) bool {
						return a.LastSelectedAt.Before(b.LastSelectedAt)
					})
				},
			},
			{
				Name:           "user1 teamMembers by member last selected at desc",
				Method:         http.MethodGet,
				URL:            "/team-members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, user1.Email)
					scenario.Headers = []string{tokenHeader}
					u, err := url.Parse("/team-members")
					assert.NoError(t, err)
					q := u.Query()
					q.Add("sort_by", "last_selected_at")
					q.Add("sort_order", "desc")
					u.RawQuery = q.Encode()
					scenario.URL = u.RequestURI()
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					members := result.Data
					assert.Equal(t, 3, len(members))
					for _, member := range result.Data {
						assert.Equal(t, user1.ID, *member.UserID)
						assert.NotNil(t, member.User)
						assert.NotNil(t, member.Team)
					}
					test.TestSliceItemsOrderByFunc(t, members, func(a, b *apis.TeamMember) bool {
						return a.LastSelectedAt.After(b.LastSelectedAt)
					})
				},
			},
			{
				Name:           "user1 teamMembers by owner role",
				Method:         http.MethodGet,
				URL:            "/team-members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, user1.Email)
					scenario.Headers = []string{tokenHeader}
					u, err := url.Parse("/team-members")
					assert.NoError(t, err)
					q := u.Query()
					q.Add("roles", "owner")
					u.RawQuery = q.Encode()
					scenario.URL = u.RequestURI()
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					members := result.Data
					assert.Equal(t, 1, len(members))
					member := members[0]
					assert.Equal(t, member.Role, apis.TeamMemberRoleOwner)
				},
			},
			{
				Name:           "user1 teamMembers by member and guest role",
				Method:         http.MethodGet,
				URL:            "/team-members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, user1.Email)
					scenario.Headers = []string{tokenHeader}
					u, err := url.Parse("/team-members")
					assert.NoError(t, err)
					q := u.Query()
					q.Add("roles", "member,guest")
					u.RawQuery = q.Encode()
					scenario.URL = u.RequestURI()
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.TeamMember]](t, res.Body.Bytes())
					members := result.Data
					assert.Equal(t, 2, len(members))
					for _, m := range members {
						assert.Contains(t, []apis.TeamMemberRole{apis.TeamMemberRoleMember, apis.TeamMemberRoleGuest}, m.Role)
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_UpdateTeamMember(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "success: team owner update member to guest",
			Method:         http.MethodPut,
			URL:            "/team-members/{team-member-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				team1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1.Team, core.TeamWithRole(models.TeamMemberRoleMember))
				assert.Equal(t, models.TeamMemberRoleMember, team1Member1.Member.Role)
				scenario.Store.Set("team1", team1)
				scenario.Store.Set("team1Member1", team1Member1)
				scenario.URL = fmt.Sprintf("/team-members/%s", team1Member1.Member.ID.String())
				body := apis.UpdateTeamMemberDto{
					Role: apis.TeamMemberRoleGuest,
				}
				scenario.Body = apis.JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, team1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				ctx := t.Context()
				team1Member1, ok := scenario.Store.Get("team1Member1").(*models.TeamInfoModel)
				assert.True(t, ok)
				updatedMember := repository.MustFindOneCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"id": map[string]any{
						"_eq": team1Member1.Member.ID,
					},
				})
				assert.Equal(t, models.TeamMemberRoleGuest, updatedMember.Role)
			},
		},
		{
			Name:           "fail: dont have roles to update",
			Method:         http.MethodPut,
			URL:            "/team-members/{team-member-id}",
			ExpectedStatus: http.StatusForbidden,
			ExpectedContent: []string{
				"You do not have the required team permission: team.members.manage",
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				team1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1.Team, core.TeamWithRole(models.TeamMemberRoleMember))
				team1Member2 := CreateTeamMember(t, app, &team1.Team, core.TeamWithRole(models.TeamMemberRoleMember))
				assert.Equal(t, models.TeamMemberRoleMember, team1Member1.Member.Role)
				assert.Equal(t, models.TeamMemberRoleMember, team1Member2.Member.Role)
				scenario.Store.Set("team1", team1)
				scenario.Store.Set("team1Member1", team1Member1)
				scenario.URL = fmt.Sprintf("/team-members/%s", team1Member1.Member.ID.String())
				body := apis.UpdateTeamMemberDto{
					Role: apis.TeamMemberRoleGuest,
				}
				scenario.Body = apis.JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, team1Member2.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
		},
		{
			Name:           "fail: does not belong to same team as member",
			Method:         http.MethodPut,
			URL:            "/team-members/{team-member-id}",
			ExpectedStatus: http.StatusForbidden,
			ExpectedContent: []string{
				"team info not found. you are not a member of the team related to this request",
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				team1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1.Team, core.TeamWithRole(models.TeamMemberRoleMember))
				team2 := CreateTeamAndOwner(t, app)
				assert.Equal(t, models.TeamMemberRoleMember, team1Member1.Member.Role)
				assert.Equal(t, models.TeamMemberRoleOwner, team2.Member.Role)
				scenario.Store.Set("team1", team1)
				scenario.Store.Set("team1Member1", team1Member1)
				scenario.URL = fmt.Sprintf("/team-members/%s", team1Member1.Member.ID.String())
				body := apis.UpdateTeamMemberDto{
					Role: apis.TeamMemberRoleGuest,
				}
				scenario.Body = apis.JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, team2.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *apis.TestApi {
				return testApi
			}
			tt.Test(t)
		})
	}
}

func TestApi_DeactivateTeamMember(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "fail: unknown error from payment client. rollback everything.",
			Method:         http.MethodDelete,
			URL:            "/team-members/{team-member-id}",
			ExpectedStatus: http.StatusInternalServerError,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				core.CreateProductsAndPrices(t, app)
				team1Owner1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team)
				sub := CreateTeamSubscription(t, app, team1Owner1, core.SubscriptionWithQuantity(2))
				assert.Equal(t, int64(2), sub.Quantity)
				scenario.Store.Set("team1Owner1", team1Owner1)
				scenario.Store.Set("team1Member1", team1Member1)
				scenario.Store.Set("subscription", sub)
				scenario.URL = fmt.Sprintf("/team-members/%s", team1Member1.Member.ID.String())
				header := core.CreateTokenHeader(t, app, team1Owner1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
				paymentClient := core.ExtractTestPaymentClient(t, app)
				paymentClient.UpdateItemQuantityFunc = func(itemId, priceId string, count int64) (*stripe.SubscriptionItem, error) {
					return nil, errors.New("unknown error")
				}
			},
			ExpectedContent: []string{
				"unknown error",
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				ctx := t.Context()
				paymentClient := core.ExtractTestPaymentClient(t, app)
				sub := scenario.Store.Get("subscription").(*models.StripeSubscription)
				team1Member1 := scenario.Store.Get("team1Member1").(*models.TeamInfoModel)
				item := paymentClient.GetUpdateSubscriptionInput(func(si *stripe.SubscriptionItem) bool {
					if si.ID == sub.ItemID {
						return true
					}
					return false
				})
				assert.Nil(t, item)
				count := repository.MustCountAllCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"team_id": map[string]any{
						"_eq": team1Member1.Team.ID,
					},
					"active": map[string]any{
						"_eq": true,
					},
				})
				assert.Equal(t, int64(2), count)
				member := repository.MustFindOneCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"id": map[string]any{
						"_eq": team1Member1.Member.ID,
					},
				})
				assert.Equal(t, true, member.Active)
				assert.NotNil(t, member.UserID)
			},
		},
		{
			Name:           "success: owner deactivates team member",
			Method:         http.MethodDelete,
			URL:            "/team-members/{team-member-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				core.CreateProductsAndPrices(t, app)
				team1Owner1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team)
				sub := CreateTeamSubscription(t, app, team1Owner1, core.SubscriptionWithQuantity(2))
				assert.Equal(t, int64(2), sub.Quantity)
				scenario.Store.Set("team1Owner1", team1Owner1)
				scenario.Store.Set("team1Member1", team1Member1)
				scenario.Store.Set("subscription", sub)
				scenario.URL = fmt.Sprintf("/team-members/%s", team1Member1.Member.ID.String())
				header := core.CreateTokenHeader(t, app, team1Owner1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				ctx := t.Context()
				paymentClient := core.ExtractTestPaymentClient(t, app)
				sub := scenario.Store.Get("subscription").(*models.StripeSubscription)
				team1Member1 := scenario.Store.Get("team1Member1").(*models.TeamInfoModel)
				item := paymentClient.GetUpdateSubscriptionInput(func(si *stripe.SubscriptionItem) bool {
					if si.ID == sub.ItemID {
						return true
					}
					return false
				})
				assert.NotNil(t, item)
				assert.Equal(t, sub.PriceID, item.Price.ID)
				count := repository.MustCountAllCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"team_id": map[string]any{
						"_eq": team1Member1.Team.ID,
					},
					"active": map[string]any{
						"_eq": true,
					},
				})
				assert.Equal(t, count, item.Quantity)
				member := repository.MustFindOneCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"id": map[string]any{
						"_eq": team1Member1.Member.ID,
					},
				})
				assert.Equal(t, false, member.Active)
				assert.Equal(t, team1Member1.Member.ID, member.ID)
				assert.Nil(t, member.UserID)
			},
		},
		{
			Name:           "fail: owner deactivates deactivated team member",
			Method:         http.MethodDelete,
			URL:            "/team-members/{team-member-id}",
			ExpectedStatus: http.StatusNotFound,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				core.CreateProductsAndPrices(t, app)
				team1Owner1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team, core.TeamWithActive(false))
				sub := CreateTeamSubscription(t, app, team1Owner1, core.SubscriptionWithQuantity(1))
				assert.Equal(t, int64(1), sub.Quantity)
				scenario.URL = fmt.Sprintf("/team-members/%s", team1Member1.Member.ID.String())
				header := core.CreateTokenHeader(t, app, team1Owner1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			ExpectedContent: []string{
				"team member not found",
			},
		},
		{
			Name:           "fail: non-owner deactivates team member",
			Method:         http.MethodDelete,
			URL:            "/team-members/{team-member-id}",
			ExpectedStatus: http.StatusForbidden,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				core.CreateProductsAndPrices(t, app)
				team1Owner1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team)
				team1Member2 := CreateTeamMember(t, app, &team1Owner1.Team)
				sub := CreateTeamSubscription(t, app, team1Owner1, core.SubscriptionWithQuantity(3))
				assert.Equal(t, int64(3), sub.Quantity)
				scenario.Store.Set("team1Owner1", team1Owner1)
				scenario.Store.Set("team1Member1", team1Member1)
				scenario.Store.Set("team1Member2", team1Member2)
				scenario.Store.Set("subscription", sub)
				scenario.URL = fmt.Sprintf("/team-members/%s", team1Member1.Member.ID.String())
				header := core.CreateTokenHeader(t, app, team1Member2.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			ExpectedContent: []string{
				"You do not have the required team permission: team.members.manage",
			},
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *apis.TestApi {
				return testApi
			}
			tt.Test(t)
		})
	}
}
func TestApi_LeaveTeam(t *testing.T) {
	tests := []apis.ApiScenario{
		{
			Name:           "fail: unknown error from payment client. rollback everything.",
			Method:         http.MethodDelete,
			URL:            "/team/{team-id}/leave",
			ExpectedStatus: http.StatusInternalServerError,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				core.CreateProductsAndPrices(t, app)
				team1Owner1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team)
				sub := CreateTeamSubscription(t, app, team1Owner1, core.SubscriptionWithQuantity(2))
				assert.Equal(t, int64(2), sub.Quantity)
				scenario.Store.Set("team1Owner1", team1Owner1)
				scenario.Store.Set("team1Member1", team1Member1)
				scenario.Store.Set("subscription", sub)
				scenario.URL = fmt.Sprintf("/team/%s/leave", team1Member1.Member.TeamID.String())
				header := core.CreateTokenHeader(t, app, team1Member1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
				paymentClient := core.ExtractTestPaymentClient(t, app)
				paymentClient.UpdateItemQuantityFunc = func(itemId, priceId string, count int64) (*stripe.SubscriptionItem, error) {
					return nil, errors.New("unknown error")
				}
			},
			ExpectedContent: []string{
				"unknown error",
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				ctx := t.Context()
				paymentClient := core.ExtractTestPaymentClient(t, app)
				sub := scenario.Store.Get("subscription").(*models.StripeSubscription)
				team1Member1 := scenario.Store.Get("team1Member1").(*models.TeamInfoModel)
				item := paymentClient.GetUpdateSubscriptionInput(func(si *stripe.SubscriptionItem) bool {
					if si.ID == sub.ItemID {
						return true
					}
					return false
				})
				assert.Nil(t, item)
				count := repository.MustCountAllCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"team_id": map[string]any{
						"_eq": team1Member1.Team.ID,
					},
					"active": map[string]any{
						"_eq": true,
					},
				})
				assert.Equal(t, int64(2), count)
				member := repository.MustFindOneCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"id": map[string]any{
						"_eq": team1Member1.Member.ID,
					},
				})
				assert.Equal(t, true, member.Active)
				assert.NotNil(t, member.UserID)
			},
		},
		{
			Name:           "success: member1 leaves team1",
			Method:         http.MethodDelete,
			URL:            "/team/{team-id}/leave",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				core.CreateProductsAndPrices(t, app)
				team1Owner1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team)
				sub := CreateTeamSubscription(t, app, team1Owner1, core.SubscriptionWithQuantity(2))
				assert.Equal(t, int64(2), sub.Quantity)
				scenario.Store.Set("team1Owner1", team1Owner1)
				scenario.Store.Set("team1Member1", team1Member1)
				scenario.Store.Set("subscription", sub)
				scenario.URL = fmt.Sprintf("/team/%s/leave", team1Member1.Member.TeamID.String())
				header := core.CreateTokenHeader(t, app, team1Member1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				ctx := t.Context()
				paymentClient := core.ExtractTestPaymentClient(t, app)
				sub := scenario.Store.Get("subscription").(*models.StripeSubscription)
				team1Member1 := scenario.Store.Get("team1Member1").(*models.TeamInfoModel)
				item := paymentClient.GetUpdateSubscriptionInput(func(si *stripe.SubscriptionItem) bool {
					if si.ID == sub.ItemID {
						return true
					}
					return false
				})
				assert.NotNil(t, item)
				assert.Equal(t, sub.PriceID, item.Price.ID)
				count := repository.MustCountAllCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"team_id": map[string]any{
						"_eq": team1Member1.Team.ID,
					},
					"active": map[string]any{
						"_eq": true,
					},
				})
				assert.Equal(t, count, item.Quantity)
				member := repository.MustFindOneCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"id": map[string]any{
						"_eq": team1Member1.Member.ID,
					},
				})
				assert.Equal(t, false, member.Active)
				assert.Equal(t, team1Member1.Member.ID, member.ID)
				assert.Nil(t, member.UserID)
			},
		},
		{
			Name:           "fail: deactivated team member leaves again",
			Method:         http.MethodDelete,
			URL:            "/team/{team-id}/leave",
			ExpectedStatus: http.StatusForbidden,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				core.CreateProductsAndPrices(t, app)
				team1Owner1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team, core.TeamWithActive(false))
				sub := CreateTeamSubscription(t, app, team1Owner1, core.SubscriptionWithQuantity(1))
				assert.Equal(t, int64(1), sub.Quantity)
				scenario.URL = fmt.Sprintf("/team/%s/leave", team1Member1.Member.TeamID.String())
				header := core.CreateTokenHeader(t, app, team1Member1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			ExpectedContent: []string{
				"team info not found. you are not a member of the team related to this request",
			},
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *apis.TestApi {
				return testApi
			}
			tt.Test(t)
		})
	}
}

func TestApi_ReassignBillingAccess_Fail_NoTeamInfoFound(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		appApi := apis.SetupApi(t, ctx, db)
		testScenario := &apis.ApiScenario{
			Name:           "Fail_NoTeamInfoFound",
			Method:         http.MethodPut,
			URL:            "/team-members/{team-member-id}/reassign-billing-access",
			ExpectedStatus: http.StatusForbidden,
			TestAppFactory: func(t testing.TB) *apis.TestApi {
				return appApi
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				// init stripe
				core.CreateProductsAndPrices(t, app)
				// team1 owner1
				team1Owner1 := CreateTeamAndOwner(t, app)
				// team1 member1
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team)
				// team2 owner2
				team2Owner1 := CreateTeamAndOwner(t, app)
				// set team1 Member1 id into url
				scenario.URL = strings.ReplaceAll(scenario.URL, "{team-member-id}", team1Member1.Member.ID.String())
				// set team2 owner1 to make request
				header := core.CreateTokenHeader(t, app, team2Owner1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			ExpectedContent: []string{
				"team info not found. you are not a member of the team related to this request",
			},
		}
		testScenario.Test(t)
	})
}
func TestApi_ReassignBillingAccess_Fail_NoBillingAccess(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		appApi := apis.SetupApi(t, ctx, db)
		testScenario := &apis.ApiScenario{
			Name:           "Fail_NoBillingAccess",
			Method:         http.MethodPut,
			URL:            "/team-members/{team-member-id}/reassign-billing-access",
			ExpectedStatus: http.StatusForbidden,
			TestAppFactory: func(t testing.TB) *apis.TestApi {
				return appApi
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				// init stripe
				core.CreateProductsAndPrices(t, app)
				// team1 owner1
				team1Owner1 := CreateTeamAndOwner(t, app)
				// team1 member1
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team)
				// team1 owner2
				team1Owner2 := CreateTeamMember(t, app, &team1Owner1.Team, core.TeamWithRole(models.TeamMemberRoleOwner), core.TeamWithBilling(false))
				// set team1 Member1 id into url
				scenario.URL = strings.ReplaceAll(scenario.URL, "{team-member-id}", team1Member1.Member.ID.String())
				// set team2 owner1 to make request
				header := core.CreateTokenHeader(t, app, team1Owner2.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			ExpectedContent: []string{
				"You do not have the required billing access",
			},
		}
		testScenario.Test(t)
	})
}
func TestApi_ReassignBillingAccess_Fail_AssignToNonOwner(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		appApi := apis.SetupApi(t, ctx, db)
		testScenario := &apis.ApiScenario{
			Name:           "Fail_AssignToNonOwner",
			Method:         http.MethodPut,
			URL:            "/team-members/{team-member-id}/reassign-billing-access",
			ExpectedStatus: http.StatusBadRequest,
			TestAppFactory: func(t testing.TB) *apis.TestApi {
				return appApi
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				// init stripe
				core.CreateProductsAndPrices(t, app)
				// team1 owner1
				team1Owner1 := CreateTeamAndOwner(t, app)
				// team1 member1
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team)
				// set team1 Member1 id into url
				scenario.URL = strings.ReplaceAll(scenario.URL, "{team-member-id}", team1Member1.Member.ID.String())
				// set team2 owner1 to make request
				header := core.CreateTokenHeader(t, app, team1Owner1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			ExpectedContent: []string{
				"member to assign is not an owner",
			},
		}
		testScenario.Test(t)
	})
}
func TestApi_ReassignBillingAccess_Fail_AssignToDeactivated(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		appApi := apis.SetupApi(t, ctx, db)
		testScenario := &apis.ApiScenario{
			Name:           "Fail_AssignToDeactivated",
			Method:         http.MethodPut,
			URL:            "/team-members/{team-member-id}/reassign-billing-access",
			ExpectedStatus: http.StatusBadRequest,
			TestAppFactory: func(t testing.TB) *apis.TestApi {
				return appApi
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				// init stripe
				core.CreateProductsAndPrices(t, app)
				// team1 owner1
				team1Owner1 := CreateTeamAndOwner(t, app)
				// team1 member1
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team, core.TeamWithActive(false), core.TeamWithRole(models.TeamMemberRoleOwner), core.TeamWithBilling(false))
				// set team1 Member1 id into url
				scenario.URL = strings.ReplaceAll(scenario.URL, "{team-member-id}", team1Member1.Member.ID.String())
				// set team2 owner1 to make request
				header := core.CreateTokenHeader(t, app, team1Owner1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			ExpectedContent: []string{
				"member to assign is not active",
			},
		}
		testScenario.Test(t)
	})
}
func TestApi_ReassignBillingAccess_Success_AssignToOwner(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		appApi := apis.SetupApi(t, ctx, db)
		testScenario := &apis.ApiScenario{
			Name:           "Success_AssignToOwner",
			Method:         http.MethodPut,
			URL:            "/team-members/{team-member-id}/reassign-billing-access",
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi {
				return appApi
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				// init stripe
				core.CreateProductsAndPrices(t, app)
				// team1 owner1
				team1Owner1 := CreateTeamAndOwner(t, app)
				scenario.Store.Set("prev_owner", team1Owner1)
				// team1 member1
				team1Member1 := CreateTeamMember(t, app, &team1Owner1.Team, core.TeamWithRole(models.TeamMemberRoleOwner), core.TeamWithBilling(false))
				scenario.Store.Set("new_owner", team1Member1)
				// sub := CreateTeamSubscription(t, app, team1Owner1, core.SubscriptionWithQuantity(2))
				// assert.Equal(t, int64(2), sub.Quantity)

				// set team1 Member1 id into url
				scenario.URL = strings.ReplaceAll(scenario.URL, "{team-member-id}", team1Member1.Member.ID.String())
				// set team2 owner1 to make request
				header := core.CreateTokenHeader(t, app, team1Owner1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				team1Owner1, ok := scenario.Store.Get("prev_owner").(*models.TeamInfoModel)
				if !ok || team1Owner1 == nil {
					t.Fail()
				}
				team1Member1, ok := scenario.Store.Get("new_owner").(*models.TeamInfoModel)
				if !ok || team1Member1 == nil {
					t.Fail()
				}
				prevOwner := repository.MustFindOneCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"id": map[string]any{
						"_eq": team1Owner1.Member.ID,
					},
				})
				nextOwner := repository.MustFindOneCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"id": map[string]any{
						"_eq": team1Member1.Member.ID,
					},
				})

				assert.False(t, prevOwner.HasBillingAccess)
				assert.True(t, nextOwner.HasBillingAccess)

				dbCustomer := repository.MustFindOneCtx(t, ctx, repository.StripeCustomer, app.Db(), &map[string]any{
					"team_id": map[string]any{
						"_eq": team1Member1.Team.ID,
					},
				})
				require.NotNil(t, dbCustomer)
				require.Equal(t, *dbCustomer.TeamID, team1Member1.Team.ID)
				require.Equal(t, dbCustomer.Email, team1Member1.User.Email)
				paymentClient := core.ExtractTestPaymentClient(t, app)
				var customer *stripe.Customer
				for _, c := range paymentClient.Customers {
					if c.ID == dbCustomer.ID {
						customer = c
					}
				}
				require.NotNil(t, customer)
				require.Equal(t, team1Member1.User.Email, customer.Email)
			},
		}
		testScenario.Test(t)
	})
}
