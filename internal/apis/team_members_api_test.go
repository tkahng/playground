package apis_test

import (
	"context"
	"errors"
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
	tests := []ApiScenario{
		{
			Name:           "success: team owner update member to guest",
			Method:         http.MethodPut,
			URL:            "/team-members/{team-member-id}",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				team1 := CreateTeamAndOwner(t, app)
				team1Member1 := CreateTeamMember(t, app, &team1.Team, core.TeamWithRole(models.TeamMemberRoleMember))
				assert.Equal(t, models.TeamMemberRoleMember, team1Member1.Member.Role)
				scenario.Store.Set("team1", team1)
				scenario.Store.Set("team1Member1", team1Member1)
				scenario.URL = fmt.Sprintf("/team-members/%s", team1Member1.Member.ID.String())
				body := apis.UpdateTeamMemberDto{
					Role: apis.TeamMemberRoleGuest,
				}
				scenario.Body = JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, team1.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
				"You do not have the required team member roles: [owner]",
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
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
				scenario.Body = JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, team1Member2.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
		},
		{
			Name:           "fail: does not belong to same team as member",
			Method:         http.MethodPut,
			URL:            "/team-members/{team-member-id}",
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedContent: []string{
				"team info not found. you are not a member of the team related to this request",
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
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
				scenario.Body = JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, team2.User.Email)
				scenario.Headers = append(scenario.Headers, header)
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

func TestApi_DeactivateTeamMember(t *testing.T) {
	tests := []ApiScenario{
		{
			Name:           "fail: unknown error from payment client. rollback everything.",
			Method:         http.MethodDelete,
			URL:            "/team-members/{team-member-id}",
			ExpectedStatus: http.StatusInternalServerError,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
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
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
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
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
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
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
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
				"You do not have the required team member roles:",
				"[owner]",
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
