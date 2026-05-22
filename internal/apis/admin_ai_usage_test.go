//go:build integration

package apis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_AdminAiUsageList(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(
			t, testApi.App,
			core.UserWithEmail("admin@k2dv.io"),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)

		// Seed two usage rows for a real team.
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
		teamInfo := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)

		for range 2 {
			if _, err := testApi.App.Adapter().AiUsage().CreateAiUsage(ctx, &models.AiUsage{
				UserID:           owner.User.ID,
				TeamMemberID:     &teamInfo.Member.ID,
				TeamID:           &teamInfo.Team.ID,
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			}); err != nil {
				t.Fatalf("CreateAiUsage() error = %v", err)
			}
		}

		tests := []apis.ApiScenario{
			{
				Name:           "list - ok returns records",
				Method:         http.MethodGet,
				URL:            "/admin/ai-usage",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.AdminAiUsageRecord]
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if body.Meta.Total < 2 {
						t.Errorf("Total = %d, want at least 2", body.Meta.Total)
					}
				},
			},
			{
				Name:           "list - filter by team_id",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/admin/ai-usage?team_id=%s", teamInfo.Team.ID),
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.AdminAiUsageRecord]
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if body.Meta.Total != 2 {
						t.Errorf("Total filtered by team = %d, want 2", body.Meta.Total)
					}
					for _, r := range body.Data {
						if r.TeamID == nil || *r.TeamID != teamInfo.Team.ID {
							t.Errorf("record has unexpected team_id")
						}
					}
				},
			},
			{
				Name:            "list - unauthorized",
				Method:          http.MethodGet,
				URL:             "/admin/ai-usage",
				ExpectedStatus:  http.StatusUnauthorized,
				Headers:         []string{},
				ExpectedContent: []string{"Unauthorized"},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
