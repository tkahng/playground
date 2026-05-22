//go:build integration

package apis_test

import (
	"context"
	"encoding/json"
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
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/workers"
)

// TestApi_NotificationFlow_TaskCompleted exercises the full pipeline:
// task marked done via API → job enqueued → worker fires → notifications persisted
// → unread count visible → mark-all-read → count drops to 0.
func TestApi_NotificationFlow_TaskCompleted(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("flow-owner@example.com"), core.UserWithVerifiedNow())
		team := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)
		project := core.CreateProjectAndTasks(t, testApi.App, &team.Member)

		other := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("flow-assignee@example.com"), core.UserWithVerifiedNow())
		otherMember := core.CreateTeamMemberWithOptions(t, testApi.App, team.Team.ID, other.User.ID)

		task, err := testApi.App.Adapter().Task().CreateTask(ctx, &models.Task{
			TeamID:            team.Team.ID,
			ProjectID:         project.ID,
			Status:            models.TaskStatusInProgress,
			Name:              "complete me",
			Rank:              100,
			CreatedByMemberID: &team.Member.ID,
			AssigneeID:        &otherMember.ID,
		})
		require.NoError(t, err)

		type countBody struct {
			Count int64 `json:"count"`
		}

		ownerHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, owner.User.Email)
		otherHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, other.User.Email)

		// Step 1: mark task done — enqueues task_completed job
		step1 := apis.ApiScenario{
			Name:           "mark task done",
			Method:         http.MethodPut,
			URL:            fmt.Sprintf("/tasks/%s", task.ID),
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				scenario.Headers = []string{ownerHeader}
				scenario.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
					Name:       "complete me",
					Status:     models.TaskStatusDone,
					AssigneeID: &otherMember.ID,
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
				assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "task_completed"))
			},
		}
		step1.Test(t)

		// Step 2: make pending jobs immediately runnable, then poll
		_, err = db.Exec(ctx, `UPDATE app.jobs SET run_after = clock_timestamp() WHERE status = 'pending'`)
		require.NoError(t, err)
		require.NoError(t, testApi.App.JobManager().PollOnce(ctx))

		// Step 3: both creator and assignee should have a notification
		for _, memberID := range []uuid.UUID{team.Member.ID, otherMember.ID} {
			count, err := testApi.App.Adapter().Notification().CountNotification(ctx, &stores.NotificationFilter{
				TeamMemberIds: []uuid.UUID{memberID},
				Types:         []string{"task_completed"},
			})
			require.NoError(t, err)
			assert.Equal(t, int64(1), count, "member %s should have 1 task_completed notification", memberID)
		}

		// Step 4: unread count for assignee should be 1
		step4 := apis.ApiScenario{
			Name:           "unread count is 1 for assignee",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/team-members/%s/notifications/unread-count", otherMember.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				scenario.Headers = []string{otherHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				body := test.MustUnMarshal[countBody](t, res.Body.Bytes())
				assert.Equal(t, int64(1), body.Count)
			},
		}
		step4.Test(t)

		// Step 5: list notifications — assignee sees the notification
		step5 := apis.ApiScenario{
			Name:            "list notifications shows unread entry",
			Method:          http.MethodGet,
			URL:             fmt.Sprintf("/team-members/%s/notifications", otherMember.ID),
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{"task_completed", "complete me"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				scenario.Headers = []string{otherHeader}
			},
		}
		step5.Test(t)

		// Step 6: mark all read → unread count drops to 0
		step6 := apis.ApiScenario{
			Name:           "mark-all-read clears unread count",
			Method:         http.MethodPost,
			URL:            fmt.Sprintf("/team-members/%s/notifications/read-all", otherMember.ID),
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				scenario.Headers = []string{otherHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
				count, err := app.Adapter().Notification().CountNotification(ctx, &stores.NotificationFilter{
					TeamMemberIds: []uuid.UUID{otherMember.ID},
					Unread:        true,
				})
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		}
		step6.Test(t)
	})
}

// TestApi_NotificationFlow_DeleteNotification verifies delete removes the record
// and is scoped to the owning member.
func TestApi_NotificationFlow_DeleteNotification(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team1 := CreateTeamAndOwner(t, testApi.App)
		team2 := CreateTeamAndOwner(t, testApi.App)

		n, err := testApi.App.Adapter().Notification().CreateNotification(ctx, &models.Notification{
			TeamMemberID: &team1.Member.ID,
			Channel:      "team_member_id:" + team1.Member.ID.String(),
			Type:         "test",
			Payload:      json.RawMessage(`{}`),
			Metadata:     map[string]any{},
		})
		require.NoError(t, err)

		tests := []apis.ApiScenario{
			{
				Name:            "fail: cross-member delete denied",
				Method:          http.MethodDelete,
				URL:             fmt.Sprintf("/team-members/%s/notifications/%s", team1.Member.ID, n.ID),
				ExpectedStatus:  http.StatusForbidden,
				ExpectedContent: []string{"team info not found"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team2.User.Email)
					scenario.Headers = []string{header}
				},
			},
			{
				Name:           "success: owner can delete own notification",
				Method:         http.MethodDelete,
				URL:            fmt.Sprintf("/team-members/%s/notifications/%s", team1.Member.ID, n.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{header}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					found, err := app.Adapter().Notification().FindNotification(ctx, &stores.NotificationFilter{
						Ids: []uuid.UUID{n.ID},
					})
					require.NoError(t, err)
					assert.Nil(t, found, "notification should be gone after delete")
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// TestApi_NotificationFlow_NewTeamMember verifies the new-member notification pipeline:
// new member added → job enqueued → worker fires → existing members notified, new member is not.
func TestApi_NotificationFlow_NewTeamMember(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("newmember-owner@example.com"), core.UserWithVerifiedNow())
		team := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)

		newUser := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("newmember-user@example.com"), core.UserWithVerifiedNow())

		newMember, err := testApi.App.Adapter().TeamMember().CreateTeamMember(ctx, &models.TeamMember{
			TeamID: team.Team.ID,
			UserID: &newUser.User.ID,
			Role:   models.TeamMemberRoleMember,
			Active: true,
		})
		require.NoError(t, err)

		require.NoError(t, testApi.App.JobService().EnqueueTeamMemberAddedJob(ctx, &workers.NewMemberNotificationJobArgs{
			TeamMemberID: newMember.ID,
		}))

		_, err = db.Exec(ctx, `UPDATE app.jobs SET run_after = clock_timestamp() WHERE status = 'pending'`)
		require.NoError(t, err)
		require.NoError(t, testApi.App.JobManager().PollOnce(ctx))

		ownerCount, err := testApi.App.Adapter().Notification().CountNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{team.Member.ID},
			Types:         []string{"new_team_member"},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), ownerCount, "owner should get new_team_member notification")

		selfCount, err := testApi.App.Adapter().Notification().CountNotification(ctx, &stores.NotificationFilter{
			TeamMemberIds: []uuid.UUID{newMember.ID},
			Types:         []string{"new_team_member"},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), selfCount, "new member should not notify themselves")
	})
}
