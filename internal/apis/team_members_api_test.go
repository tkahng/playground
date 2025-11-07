package apis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/notification"
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
func TestApi_FindTeamTeamMembers(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		team1 := CreateTeamAndOwner(t, testApi.App)
		team1Member1 := CreateTeamMember(t, testApi.App, &team1.Team)
		team1Member2 := CreateTeamMember(t, testApi.App, &team1.Team, core.TeamWithActive(false))
		team1Member3 := CreateTeamMember(t, testApi.App, &team1.Team, core.TeamWithRole(models.TeamMemberRoleGuest))
		// testMailer := ExtractTestMailer(t, testApi.App)
		tests := []ApiScenario{
			{
				Name:           "success: team1 owner find everyone",
				Method:         http.MethodGet,
				URL:            "/teams/{team-id}/members",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members", team1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members?active=true", team1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members?active=true&roles=guest,member", team1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1Member1.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members", team1Member1.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1Member3.User.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = fmt.Sprintf("/teams/%s/members", team1Member3.Team.ID.String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
					scenario.URL = fmt.Sprintf("/teams/%s/members", team1Member2.Team.ID.String())
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_UpdateTeamMember(t *testing.T) {
	inviteeEmail := "VY7o1@example.com"
	tests := []ApiScenario{
		{
			Name:           "success: accept invitation",
			Method:         http.MethodPost,
			URL:            "/team-invitations/accept",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				ctx := t.Context()
				core.CreateProductsAndPrices(t, app)
				// init team
				teamInfo := CreateTeamAndOwner(t, app)
				sub := CreateTeamSubscription(t, app, teamInfo)
				scenario.Store.Set("teamInfo", teamInfo)
				scenario.Store.Set("subscription", sub)
				// send invitation and get token
				err := app.TeamInvitation().CreateInvitation(
					ctx,
					teamInfo.Team.ID,
					teamInfo.User.ID,
					inviteeEmail,
					models.TeamMemberRoleMember,
					true,
				)
				assert.NoError(t, err)
				if err := app.JobManager().PollOnce(ctx); err != nil {
					t.Fatal(err)
				}
				token := ExtractFistMessageTokenFromMailer(t, app)

				// create invitee user
				inviteeUserInfo := core.CreateUserWithOptions(t, app, core.UserWithEmail(inviteeEmail), core.UserWithVerifiedNow())
				body := apis.CheckValidInvitationDto{
					Token: token,
				}
				scenario.Body = JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, inviteeUserInfo.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
				ctx := t.Context()
				for range 5 {
					err := app.JobManager().PollOnce(ctx)
					assert.NoError(t, err)
				}
				paymentClient := core.ExtractTestPaymentClient(t, app)
				sub := scenario.Store.Get("subscription").(*models.StripeSubscription)
				item := paymentClient.GetUpdateSubscriptionInput(func(si *stripe.SubscriptionItem) bool {
					if si.ID == sub.ItemID {
						return true
					}
					return false
				})
				count := repository.MustCountAllCtx(t, ctx, repository.TeamMember, app.Db(), nil)
				assert.NotNil(t, item)
				assert.Equal(t, count, item.Quantity)
				assert.Equal(t, sub.PriceID, item.Price.ID)
				// notifications
				teamInfo := scenario.Store.Get("teamInfo").(*models.TeamInfoModel)
				notifications := repository.MustFindWithOptionsCtx(t, ctx, repository.Notification, app.Db())
				assert.Len(t, notifications, 1)
				noti := notifications[0]

				payloadString := noti.Payload
				var payload notification.NotificationPayload[notification.NewTeamMemberNotificationData]
				err := json.Unmarshal(payloadString, &payload)

				assert.NoError(t, err)
				assert.Equal(t, teamInfo.Team.ID, payload.Data.TeamID)
				assert.Equal(t, &teamInfo.Member.ID, noti.TeamMemberID)

				member := repository.MustFindOneCtx(t, ctx, repository.TeamMember, app.Db(), &map[string]any{
					"user": map[string]any{
						"email": map[string]any{
							"_eq": inviteeEmail,
						},
					},
				})
				assert.Equal(t, member.ID, payload.Data.TeamMemberID)
			},
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *TestApi {
				return testApi
			}
			tt.Test(t)
		})
	}
}
