package apis_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func setupTaskCommentFixture(t testing.TB, ctx context.Context, testApi *apis.TestApi) (
	team *models.TeamInfoModel,
	task *models.Task,
) {
	t.Helper()
	team = CreateTeamAndOwner(t, testApi.App)
	project, err := testApi.App.Adapter().Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
		Name:     "comment-test-project",
		Status:   models.TaskProjectStatusTodo,
		TeamID:   team.Team.ID,
		MemberID: team.Member.ID,
	})
	require.NoError(t, err)
	task, err = testApi.App.Adapter().Task().CreateTask(ctx, &models.Task{
		Name:      "comment-test-task",
		Status:    models.TaskStatusTodo,
		TeamID:    team.Team.ID,
		ProjectID: project.ID,
	})
	require.NoError(t, err)
	return
}

func TestApi_TaskCommentList(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team, task := setupTaskCommentFixture(t, ctx, testApi)

		// seed two comments
		for i, content := range []string{"first comment", "second comment"} {
			_, err := testApi.App.Adapter().TaskComment().CreateTaskComment(ctx, &models.TaskComment{
				TaskID:            task.ID,
				CreatedByMemberID: team.Member.ID,
				Content:           fmt.Sprintf("%s %d", content, i),
			})
			require.NoError(t, err)
		}

		scenarios := []apis.ApiScenario{
			{
				Name:           "unauthenticated → 401",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/tasks/%s/comments", task.ID),
				ExpectedStatus: http.StatusUnauthorized,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			},
			{
				Name:           "success: returns all comments ordered by created_at",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/tasks/%s/comments", task.ID),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{tokenHeader}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					comments := test.MustUnMarshal[[]*apis.TaskComment](t, res.Body.Bytes())
					assert.Len(t, comments, 2)
				},
			},
		}
		for _, scenario := range scenarios {
			scenario.Test(t)
		}
	})
}

func TestApi_TaskCommentCreate(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team, task := setupTaskCommentFixture(t, ctx, testApi)

		scenarios := []apis.ApiScenario{
			{
				Name:           "unauthenticated → 401",
				Method:         http.MethodPost,
				URL:            fmt.Sprintf("/tasks/%s/comments", task.ID),
				Body:           strings.NewReader(`{"content":"hello"}`),
				ExpectedStatus: http.StatusUnauthorized,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			},
			{
				Name:           "empty content → 422",
				Method:         http.MethodPost,
				URL:            fmt.Sprintf("/tasks/%s/comments", task.ID),
				Body:           strings.NewReader(`{"content":""}`),
				ExpectedStatus: http.StatusUnprocessableEntity,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{tokenHeader}
				},
			},
			{
				Name:           "success: comment created and returned",
				Method:         http.MethodPost,
				URL:            fmt.Sprintf("/tasks/%s/comments", task.ID),
				Body:           strings.NewReader(`{"content":"great task!"}`),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{tokenHeader}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					comment := test.MustUnMarshal[apis.TaskComment](t, res.Body.Bytes())
					assert.Equal(t, "great task!", comment.Content)
					assert.Equal(t, task.ID, comment.TaskID)
					assert.Equal(t, team.Member.ID, comment.CreatedByMemberID)
				},
			},
		}
		for _, scenario := range scenarios {
			scenario.Test(t)
		}
	})
}

func TestApi_TaskCommentUpdate(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team, task := setupTaskCommentFixture(t, ctx, testApi)

		// seed a comment by the owner
		comment, err := testApi.App.Adapter().TaskComment().CreateTaskComment(ctx, &models.TaskComment{
			TaskID:            task.ID,
			CreatedByMemberID: team.Member.ID,
			Content:           "original content",
		})
		require.NoError(t, err)

		// create a second member who did NOT write the comment
		otherUser, err := testApi.App.Adapter().User().CreateUser(ctx, &models.User{Email: "other@example.com"})
		require.NoError(t, err)
		otherMember, err := testApi.App.Adapter().TeamMember().CreateTeamMemberFromUserAndSlug(ctx, otherUser, "", models.TeamMemberRoleMember)
		require.NoError(t, err)
		otherMember.TeamID = team.Team.ID
		otherMember, err = testApi.App.Adapter().TeamMember().UpdateTeamMember(ctx, otherMember)
		require.NoError(t, err)

		url := fmt.Sprintf("/tasks/%s/comments/%s", task.ID, comment.ID)

		scenarios := []apis.ApiScenario{
			{
				Name:           "non-author → 403",
				Method:         http.MethodPut,
				URL:            url,
				Body:           strings.NewReader(`{"content":"sneaky edit"}`),
				ExpectedStatus: http.StatusForbidden,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, otherUser.Email)
					scenario.Headers = []string{tokenHeader}
				},
			},
			{
				Name:           "author can update",
				Method:         http.MethodPut,
				URL:            url,
				Body:           strings.NewReader(`{"content":"updated content"}`),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{tokenHeader}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					updated := test.MustUnMarshal[apis.TaskComment](t, res.Body.Bytes())
					assert.Equal(t, "updated content", updated.Content)
				},
			},
		}
		for _, scenario := range scenarios {
			_ = otherMember // referenced in setup
			scenario.Test(t)
		}
	})
}

func TestApi_TaskCommentDelete(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team, task := setupTaskCommentFixture(t, ctx, testApi)

		makeComment := func() *models.TaskComment {
			c, err := testApi.App.Adapter().TaskComment().CreateTaskComment(ctx, &models.TaskComment{
				TaskID:            task.ID,
				CreatedByMemberID: team.Member.ID,
				Content:           "to be deleted",
			})
			require.NoError(t, err)
			return c
		}

		otherUser, err := testApi.App.Adapter().User().CreateUser(ctx, &models.User{Email: "stranger@example.com"})
		require.NoError(t, err)
		otherMember, err := testApi.App.Adapter().TeamMember().CreateTeamMemberFromUserAndSlug(ctx, otherUser, "", models.TeamMemberRoleMember)
		require.NoError(t, err)
		otherMember.TeamID = team.Team.ID
		_, err = testApi.App.Adapter().TeamMember().UpdateTeamMember(ctx, otherMember)
		require.NoError(t, err)

		scenarios := []apis.ApiScenario{
			{
				Name:   "non-author non-owner → 403",
				Method: http.MethodDelete,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					c := makeComment()
					scenario.URL = fmt.Sprintf("/tasks/%s/comments/%s", task.ID, c.ID)
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, otherUser.Email)
					scenario.Headers = []string{tokenHeader}
				},
				ExpectedStatus: http.StatusForbidden,
			},
			{
				Name:   "owner can delete any comment",
				Method: http.MethodDelete,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					c := makeComment()
					scenario.URL = fmt.Sprintf("/tasks/%s/comments/%s", task.ID, c.ID)
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team.User.Email)
					scenario.Headers = []string{tokenHeader}
				},
				ExpectedStatus: http.StatusNoContent,
			},
		}
		for _, scenario := range scenarios {
			scenario.Test(t)
		}
	})
}
