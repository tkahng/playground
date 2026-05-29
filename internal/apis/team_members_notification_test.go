package apis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_UnreadNotificationsCount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team1 := CreateTeamAndOwner(t, testApi.App)
		team2 := CreateTeamAndOwner(t, testApi.App)
		adapter := testApi.App.Adapter()

		readAt := time.Now()
		for range 2 {
			_, err := adapter.Notification().CreateNotification(ctx, &models.Notification{
				TeamMemberID: &team1.Member.ID,
				Channel:      "team_member_id:" + team1.Member.ID.String(),
				Type:         "test",
				Payload:      json.RawMessage(`{}`),
				Metadata:     map[string]any{},
			})
			require.NoError(t, err)
		}
		_, err := adapter.Notification().CreateNotification(ctx, &models.Notification{
			TeamMemberID: &team1.Member.ID,
			Channel:      "team_member_id:" + team1.Member.ID.String(),
			Type:         "test",
			Payload:      json.RawMessage(`{}`),
			Metadata:     map[string]any{},
			ReadAt:       &readAt,
		})
		require.NoError(t, err)

		type countBody struct {
			Count int64 `json:"count"`
		}

		tests := []apis.ApiScenario{
			{
				Name:           "success: count is 2 unread",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/team-members/%s/notifications/unread-count", team1.Member.ID),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{header}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					body := test.MustUnMarshal[countBody](t, res.Body.Bytes())
					assert.Equal(t, int64(2), body.Count)
				},
			},
			{
				Name:            "fail: cross-member access denied",
				Method:          http.MethodGet,
				URL:             fmt.Sprintf("/team-members/%s/notifications/unread-count", team1.Member.ID),
				ExpectedStatus:  http.StatusForbidden,
				ExpectedContent: []string{"team info not found"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team2.User.Email)
					scenario.Headers = []string{header}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_MarkAllNotificationsRead(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team1 := CreateTeamAndOwner(t, testApi.App)
		team2 := CreateTeamAndOwner(t, testApi.App)
		adapter := testApi.App.Adapter()

		for range 3 {
			_, err := adapter.Notification().CreateNotification(ctx, &models.Notification{
				TeamMemberID: &team1.Member.ID,
				Channel:      "team_member_id:" + team1.Member.ID.String(),
				Type:         "test",
				Payload:      json.RawMessage(`{}`),
				Metadata:     map[string]any{},
			})
			require.NoError(t, err)
		}

		tests := []apis.ApiScenario{
			{
				Name:            "fail: cross-member access denied",
				Method:          http.MethodPost,
				URL:             fmt.Sprintf("/team-members/%s/notifications/read-all", team1.Member.ID),
				ExpectedStatus:  http.StatusForbidden,
				ExpectedContent: []string{"team info not found"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team2.User.Email)
					scenario.Headers = []string{header}
				},
			},
			{
				Name:           "success: unread count becomes 0 after mark-all-read",
				Method:         http.MethodPost,
				URL:            fmt.Sprintf("/team-members/%s/notifications/read-all", team1.Member.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{header}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, _ *httptest.ResponseRecorder) {
					count, err := adapter.Notification().CountNotification(ctx, &stores.NotificationFilter{
						TeamMemberIds: []uuid.UUID{team1.Member.ID},
						Unread:        true,
					})
					require.NoError(t, err)
					assert.Equal(t, int64(0), count)
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_GetNotificationPreferences(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team1 := CreateTeamAndOwner(t, testApi.App)
		team2 := CreateTeamAndOwner(t, testApi.App)

		type prefsBody struct {
			Preferences []struct {
				Type    string `json:"type"`
				Enabled bool   `json:"enabled"`
			} `json:"preferences"`
		}

		tests := []apis.ApiScenario{
			{
				Name:           "success: returns all known types defaulting to enabled",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/team-members/%s/notification-preferences", team1.Member.ID),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{header}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					body := test.MustUnMarshal[prefsBody](t, res.Body.Bytes())
					assert.NotEmpty(t, body.Preferences)
					for _, p := range body.Preferences {
						assert.True(t, p.Enabled, "all types should default to enabled before any opt-out")
					}
				},
			},
			{
				Name:            "fail: cross-member access denied",
				Method:          http.MethodGet,
				URL:             fmt.Sprintf("/team-members/%s/notification-preferences", team1.Member.ID),
				ExpectedStatus:  http.StatusForbidden,
				ExpectedContent: []string{"team info not found"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team2.User.Email)
					scenario.Headers = []string{header}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_UpsertNotificationPreference(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		team1 := CreateTeamAndOwner(t, testApi.App)
		team2 := CreateTeamAndOwner(t, testApi.App)

		tests := []apis.ApiScenario{
			{
				Name:           "success: opt out of task_completed",
				Method:         http.MethodPut,
				URL:            fmt.Sprintf("/team-members/%s/notification-preferences/task_completed", team1.Member.ID),
				Body:           strings.NewReader(`{"enabled":false}`),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{header}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					prefs, err := app.Adapter().Notification().FindNotificationPreferences(ctx, team1.Member.ID)
					require.NoError(t, err)
					found := false
					for _, p := range prefs {
						if p.Type == "task_completed" {
							found = true
							assert.False(t, p.Enabled)
						}
					}
					assert.True(t, found, "task_completed preference row should exist after upsert")
				},
			},
			{
				Name:            "fail: unknown notification type",
				Method:          http.MethodPut,
				URL:             fmt.Sprintf("/team-members/%s/notification-preferences/nonexistent_type", team1.Member.ID),
				Body:            strings.NewReader(`{"enabled":false}`),
				ExpectedStatus:  http.StatusBadRequest,
				ExpectedContent: []string{"nonexistent_type"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team1.User.Email)
					scenario.Headers = []string{header}
				},
			},
			{
				Name:            "fail: cross-member access denied",
				Method:          http.MethodPut,
				URL:             fmt.Sprintf("/team-members/%s/notification-preferences/task_completed", team1.Member.ID),
				Body:            strings.NewReader(`{"enabled":false}`),
				ExpectedStatus:  http.StatusForbidden,
				ExpectedContent: []string{"team info not found"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, team2.User.Email)
					scenario.Headers = []string{header}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
