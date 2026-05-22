//go:build integration

package apis_test

// Tests verifying that TeamTaskUpdateBind reads the CURRENT committed DB state
// (via SELECT … FOR UPDATE inside a transaction) before deciding which notification
// jobs to enqueue.  The key invariant: previousState must reflect what is actually
// in the database at lock time, not a value captured before a concurrent writer
// committed.

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
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

// TestApi_TaskUpdate_ReadsCurrentAssigneeFromDB verifies that the update handler
// detects an assignee change relative to what is committed in the database at the
// time of the request, not relative to any earlier in-process snapshot.
//
// Scenario simulating a committed concurrent write before our request:
//  1. Task created with assignee = memberA.
//  2. DB row updated directly (bypassing the handler) to assignee = nil —
//     models a concurrent request that already committed.
//  3. PUT /tasks/:id with assignee = memberA.
//  4. Expected: the handler sees previousAssignee = nil (current DB value),
//     detects a change nil → memberA, and enqueues exactly 1 assigned_to_task job.
func TestApi_TaskUpdate_ReadsCurrentAssigneeFromDB(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
		team := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)
		project := core.CreateProjectAndTasks(t, testApi.App, &team.Member)

		// Create task initially assigned to the team owner.
		task := repository.MustCreateOneCtx(t, ctx, repository.Task, db, &models.Task{
			TeamID:     team.Team.ID,
			ProjectID:  project.ID,
			Status:     models.TaskStatusTodo,
			Rank:       0,
			AssigneeID: &team.Member.ID,
		})

		// Simulate a concurrent committed write: clear the assignee directly in the DB.
		// The handler must read this state, not the original creation state.
		_, err := db.Exec(ctx, "UPDATE task.tasks SET assignee_id = NULL WHERE id = $1", task.ID)
		require.NoError(t, err)

		tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, team.User.Email)

		scenario := apis.ApiScenario{
			Name:           "detects nil→member change from current DB state",
			Method:         http.MethodPut,
			URL:            fmt.Sprintf("/tasks/%s", task.ID),
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				sc.Headers = []string{tokenHeader}
				sc.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
					Name:       task.Name,
					Status:     models.TaskStatusTodo,
					AssigneeID: &team.Member.ID, // re-assigning; current DB state is nil
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario, _ *httptest.ResponseRecorder) {
				// previousAssignee was nil in the DB → new assignment → job enqueued
				assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "assigned_to_task"))
			},
		}
		scenario.Test(t)
	})
}

// TestApi_TaskUpdate_NoJobWhenAssigneeUnchangedInDB verifies that when the assignee
// in the DB already matches the value being set, no spurious notification is fired.
func TestApi_TaskUpdate_NoJobWhenAssigneeUnchangedInDB(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
		team := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)
		project := core.CreateProjectAndTasks(t, testApi.App, &team.Member)

		task := repository.MustCreateOneCtx(t, ctx, repository.Task, db, &models.Task{
			TeamID:     team.Team.ID,
			ProjectID:  project.ID,
			Status:     models.TaskStatusTodo,
			Rank:       0,
			AssigneeID: &team.Member.ID,
		})

		tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, team.User.Email)

		scenario := apis.ApiScenario{
			Name:           "no job when assignee matches current DB value",
			Method:         http.MethodPut,
			URL:            fmt.Sprintf("/tasks/%s", task.ID),
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				sc.Headers = []string{tokenHeader}
				sc.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
					Name:       task.Name,
					Status:     models.TaskStatusTodo,
					AssigneeID: &team.Member.ID, // same as in DB
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario, _ *httptest.ResponseRecorder) {
				assert.Equal(t, int64(0), countJobsByKind(t, ctx, db, "assigned_to_task"))
			},
		}
		scenario.Test(t)
	})
}

// TestApi_TaskUpdate_DetectsAssigneeChangeInDB verifies that when the DB assignee
// differs from the new value, a notification job is enqueued.
func TestApi_TaskUpdate_DetectsAssigneeChangeInDB(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
		team := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)

		// Create a second user and add them to the team as another member.
		otherUser := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow(), core.UserWithEmail("tkahng+02@gmail.com"))
		otherTeamInfo := core.CreateTeamAndMemberWithOptions(t, testApi.App, &otherUser.User)
		otherMember := repository.MustCreateOneCtx(t, ctx, repository.TeamMember, db, &models.TeamMember{
			TeamID: team.Team.ID,
			UserID: types.Pointer(otherUser.User.ID),
			Role:   models.TeamMemberRoleMember,
			Active: true,
		})
		_ = otherTeamInfo

		project := core.CreateProjectAndTasks(t, testApi.App, &team.Member)

		// Task starts assigned to the owner.
		task := repository.MustCreateOneCtx(t, ctx, repository.Task, db, &models.Task{
			TeamID:     team.Team.ID,
			ProjectID:  project.ID,
			Status:     models.TaskStatusTodo,
			Rank:       0,
			AssigneeID: &team.Member.ID,
		})

		tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, team.User.Email)

		scenario := apis.ApiScenario{
			Name:           "detects memberA→memberB change in DB",
			Method:         http.MethodPut,
			URL:            fmt.Sprintf("/tasks/%s", task.ID),
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario) {
				sc.Headers = []string{tokenHeader}
				sc.Body = apis.JsonToReader(t, stores.UpdateTaskDto{
					Name:       task.Name,
					Status:     models.TaskStatusTodo,
					AssigneeID: &otherMember.ID, // change from owner to otherMember
				})
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, sc *apis.ApiScenario, _ *httptest.ResponseRecorder) {
				assert.Equal(t, int64(1), countJobsByKind(t, ctx, db, "assigned_to_task"))
			},
		}
		scenario.Test(t)
	})
}
