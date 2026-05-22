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
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_TeamAiUsageStatus(t *testing.T) {
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		owner := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
		teamInfo := core.CreateTeamAndMemberWithOptions(t, testApi.App, &owner.User)
		header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, owner.User.Email)

		tests := []apis.ApiScenario{
			{
				Name:           "status - ok with zero usage",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/teams/%s/ai-usage", teamInfo.Team.ID),
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.AiUsageStatus
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if body.Consumed != 0 {
						t.Errorf("Consumed = %d, want 0", body.Consumed)
					}
					if body.Limit <= 0 {
						t.Errorf("Limit = %d, want > 0", body.Limit)
					}
					if body.Remaining != body.Limit {
						t.Errorf("Remaining = %d, want %d (= Limit)", body.Remaining, body.Limit)
					}
				},
			},
			{
				Name:           "status - reflects recorded usage",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/teams/%s/ai-usage", teamInfo.Team.ID),
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, _ *apis.ApiScenario) {
					if _, err := app.Adapter().AiUsage().CreateAiUsage(ctx, &models.AiUsage{
						UserID:           owner.User.ID,
						TeamMemberID:     &teamInfo.Member.ID,
						TeamID:           &teamInfo.Team.ID,
						PromptTokens:     1_000,
						CompletionTokens: 500,
						TotalTokens:      1_500,
					}); err != nil {
						t.Fatalf("CreateAiUsage() error = %v", err)
					}
				},
				AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.AiUsageStatus
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if body.Consumed != 1_500 {
						t.Errorf("Consumed = %d, want 1500", body.Consumed)
					}
					if body.Remaining != body.Limit-1_500 {
						t.Errorf("Remaining = %d, want %d", body.Remaining, body.Limit-1_500)
					}
				},
			},
			{
				Name:           "status - remaining floors at zero when over limit",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/teams/%s/ai-usage", teamInfo.Team.ID),
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, _ *apis.ApiScenario) {
					// Record usage exceeding the free-tier limit.
					over := services.FreeTierDailyTokenLimit + 1
					if _, err := app.Adapter().AiUsage().CreateAiUsage(ctx, &models.AiUsage{
						UserID:           owner.User.ID,
						TeamMemberID:     &teamInfo.Member.ID,
						TeamID:           &teamInfo.Team.ID,
						PromptTokens:     over,
						CompletionTokens: 0,
						TotalTokens:      over,
					}); err != nil {
						t.Fatalf("CreateAiUsage() error = %v", err)
					}
				},
				AfterTestFunc: func(t testing.TB, _ *core.BaseApp, _ *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.AiUsageStatus
					if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
						t.Fatalf("decode error: %v", err)
					}
					if body.Remaining != 0 {
						t.Errorf("Remaining = %d, want 0 when over limit", body.Remaining)
					}
				},
			},
			{
				Name:            "status - unauthorized",
				Method:          http.MethodGet,
				URL:             fmt.Sprintf("/teams/%s/ai-usage", teamInfo.Team.ID),
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
