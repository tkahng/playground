package apis_test

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v82"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/store"
)

func TestApi_CreateInvitation(t *testing.T) {
	tests := []ApiScenario{
		{
			Name:           "success: create invitation by owner member",
			Method:         http.MethodPost,
			URL:            "/teams/{team-id}/invitations",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				user := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &user.User)
				scenario.URL = fmt.Sprintf("/teams/%s/invitations", team.Team.ID)
				body := apis.InviteTeamMemberDto{
					Email: "VY7o1@example.com",
					Role:  string(apis.TeamMemberRoleMember),
				}
				scenario.Body = JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, user.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
				ctx := t.Context()
				err := app.JobManager().PollOnce(ctx)
				assert.NoError(t, err)
				stoken := ExtractFistMessageTokenFromMailer(t, app)
				assert.NoError(t, err)
				team := repository.MustFindOneCtx(t, ctx, repository.Team, app.Db(), nil)
				assert.NotNil(t, team)
				member := repository.MustFindOneCtx(t, ctx, repository.TeamMember, app.Db(), nil)
				assert.NotNil(t, member)
				invitation := repository.MustFindOneCtx(t, ctx, repository.TeamInvitation, app.Db(), nil)
				assert.NotNil(t, invitation)
				assert.Equal(t, invitation.Token, stoken)
				assert.Equal(t, invitation.Email, "VY7o1@example.com")
				assert.Equal(t, invitation.Status, models.TeamInvitationStatusPending)
				assert.Equal(t, invitation.Role, models.TeamMemberRoleMember)
				assert.Equal(t, invitation.TeamID, team.ID)
				assert.Equal(t, invitation.InviterMemberID, member.ID)
			},
		},
		{
			Name:           "fail: create invitation by non-owner member",
			Method:         http.MethodPost,
			URL:            "/teams/{team-id}/invitations",
			ExpectedStatus: http.StatusForbidden,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				user := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				team := core.CreateTeamAndMemberWithOptions(t, app, &user.User, core.TeamWithRole(models.TeamMemberRoleMember))
				scenario.URL = fmt.Sprintf("/teams/%s/invitations", team.Team.ID)
				body := apis.InviteTeamMemberDto{
					Email: "VY7o1@example.com",
					Role:  string(apis.TeamMemberRoleMember),
				}
				scenario.Body = JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, user.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			ExpectedContent: []string{
				"Forbidden",
				"You do not have the required team member role",
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

func ExtractFistMessageTokenFromMailer(t testing.TB, app *core.BaseApp) string {
	t.Helper()
	mailer := core.ExtractTestMailer(t, app)
	assert.Len(t, mailer.Messages, 1)
	raw := html.UnescapeString(mailer.Messages[0].Body)
	fmt.Println(raw)
	stoken, err := test.GetLinkParam(raw, "token")
	assert.NoError(t, err)
	return stoken
}
func TestApi_AcceptInvitation(t *testing.T) {
	inviteeEmail := "VY7o1@example.com"
	tests := []ApiScenario{
		{
			Name:           "success: accept invitation",
			Method:         http.MethodPost,
			URL:            "/team-invitations/accept",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				ctx := t.Context()
				if err := app.Payment().FindAndUpsertAllProducts(ctx); err != nil {
					t.Fatal(err)
				}
				if err := app.Payment().FindAndUpsertAllPrices(ctx); err != nil {
					t.Fatal(err)
				}
				// init team
				ownerUserInfo := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				teamInfo := core.CreateTeamAndMemberWithOptions(t, app, &ownerUserInfo.User)
				teamCustomer, err := app.Payment().FindCustomerByTeamId(ctx, teamInfo.Team.ID)
				assert.NoError(t, err)
				sub := core.CreateStripeSubscriptionWithOptions(
					t,
					app,
					teamCustomer.ID,
					core.SubscriptionWithID("sub_1"),
					core.SubscriptionWithItemID("item_1"),
					core.SubscriptionWithPriceID("price_pro_month_usd_5000"),
				)
				scenario.Store.Set("subscription", sub)
				// send invitation and get token
				err = app.TeamInvitation().CreateInvitation(
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
			},
		},
		{
			Name:           "fail: user mismatch",
			Method:         http.MethodPost,
			URL:            "/team-invitations/accept",
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedContent: []string{
				"user does not match invitation",
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				ctx := t.Context()
				if err := app.Payment().FindAndUpsertAllProducts(ctx); err != nil {
					t.Fatal(err)
				}
				if err := app.Payment().FindAndUpsertAllPrices(ctx); err != nil {
					t.Fatal(err)
				}
				// init team
				ownerUserInfo := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				teamInfo := core.CreateTeamAndMemberWithOptions(t, app, &ownerUserInfo.User)
				teamCustomer, err := app.Payment().FindCustomerByTeamId(ctx, teamInfo.Team.ID)
				assert.NoError(t, err)
				sub := core.CreateStripeSubscriptionWithOptions(
					t,
					app,
					teamCustomer.ID,
					core.SubscriptionWithID("sub_1"),
					core.SubscriptionWithItemID("item_1"),
					core.SubscriptionWithPriceID("price_pro_month_usd_5000"),
				)
				scenario.Store.Set("subscription", sub)
				// send invitation and get token
				err = app.TeamInvitation().CreateInvitation(
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
				_ = core.CreateUserWithOptions(t, app, core.UserWithEmail(inviteeEmail), core.UserWithVerifiedNow())
				body := apis.CheckValidInvitationDto{
					Token: token,
				}
				otherUserInfo := core.CreateUserWithOptions(t, app, core.UserWithEmail("other@example"), core.UserWithVerifiedNow())
				scenario.Body = JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, otherUserInfo.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
				ctx := t.Context()
				for range 5 {
					err := app.JobManager().PollOnce(ctx)
					assert.NoError(t, err)
				}
				paymentClient := core.ExtractTestPaymentClient(t, app)
				assert.Len(t, paymentClient.SubscriptionItems, 0)
			},
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *TestApi {
				return testApi
			}
			tt.Store = store.New[string, any](nil)
			tt.Test(t)
		})
	}
}
func TestApi_RejectInvitation(t *testing.T) {
	inviteeEmail := "VY7o1@example.com"
	tests := []ApiScenario{
		{
			Name:           "success: decline invitation",
			Method:         http.MethodPost,
			URL:            "/team-invitations/decline",
			ExpectedStatus: http.StatusNoContent,
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				ctx := t.Context()
				if err := app.Payment().FindAndUpsertAllProducts(ctx); err != nil {
					t.Fatal(err)
				}
				if err := app.Payment().FindAndUpsertAllPrices(ctx); err != nil {
					t.Fatal(err)
				}
				// init team
				ownerUserInfo := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				teamInfo := core.CreateTeamAndMemberWithOptions(t, app, &ownerUserInfo.User)
				teamCustomer, err := app.Payment().FindCustomerByTeamId(ctx, teamInfo.Team.ID)
				assert.NoError(t, err)
				sub := core.CreateStripeSubscriptionWithOptions(
					t,
					app,
					teamCustomer.ID,
					core.SubscriptionWithID("sub_1"),
					core.SubscriptionWithItemID("item_1"),
					core.SubscriptionWithPriceID("price_pro_month_usd_5000"),
				)
				scenario.Store.Set("subscription", sub)
				// send invitation and get token
				err = app.TeamInvitation().CreateInvitation(
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
				assert.Len(t, paymentClient.SubscriptionItems, 0)
				count := repository.MustCountAllCtx(t, ctx, repository.TeamMember, app.Db(), nil)
				assert.Equal(t, int64(1), count)
				invitationCount := repository.MustFindOneCtx(t, ctx, repository.TeamInvitation, app.Db(), nil)
				assert.Equal(t, models.TeamInvitationStatusDeclined, invitationCount.Status)

			},
		},
		{
			Name:           "fail: user mismatch",
			Method:         http.MethodPost,
			URL:            "/team-invitations/decline",
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedContent: []string{
				"user does not match invitation",
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				ctx := t.Context()
				if err := app.Payment().FindAndUpsertAllProducts(ctx); err != nil {
					t.Fatal(err)
				}
				if err := app.Payment().FindAndUpsertAllPrices(ctx); err != nil {
					t.Fatal(err)
				}
				// init team
				ownerUserInfo := core.CreateUserWithOptions(t, app, core.UserWithVerifiedNow())
				teamInfo := core.CreateTeamAndMemberWithOptions(t, app, &ownerUserInfo.User)
				teamCustomer, err := app.Payment().FindCustomerByTeamId(ctx, teamInfo.Team.ID)
				assert.NoError(t, err)
				sub := core.CreateStripeSubscriptionWithOptions(
					t,
					app,
					teamCustomer.ID,
					core.SubscriptionWithID("sub_1"),
					core.SubscriptionWithItemID("item_1"),
					core.SubscriptionWithPriceID("price_pro_month_usd_5000"),
				)
				scenario.Store.Set("subscription", sub)
				// send invitation and get token
				err = app.TeamInvitation().CreateInvitation(
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
				_ = core.CreateUserWithOptions(t, app, core.UserWithEmail(inviteeEmail), core.UserWithVerifiedNow())
				body := apis.CheckValidInvitationDto{
					Token: token,
				}
				otherUserInfo := core.CreateUserWithOptions(t, app, core.UserWithEmail("other@example"), core.UserWithVerifiedNow())
				scenario.Body = JsonToReader(t, body)
				header := core.CreateTokenHeader(t, app, otherUserInfo.User.Email)
				scenario.Headers = append(scenario.Headers, header)
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
				ctx := t.Context()
				for range 5 {
					err := app.JobManager().PollOnce(ctx)
					assert.NoError(t, err)
				}
				paymentClient := core.ExtractTestPaymentClient(t, app)
				assert.Len(t, paymentClient.SubscriptionItems, 0)
				invitationCount := repository.MustFindOneCtx(t, ctx, repository.TeamInvitation, app.Db(), nil)
				assert.Equal(t, models.TeamInvitationStatusPending, invitationCount.Status)
			},
		},
	}
	for _, tt := range tests {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := SetupApi(t, ctx, db)
			tt.TestAppFactory = func(t testing.TB) *TestApi {
				return testApi
			}
			tt.Store = store.New[string, any](nil)
			tt.Test(t)
		})
	}
}
