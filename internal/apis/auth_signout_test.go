//go:build integration

package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_Signout(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		// testMailer := ExtractTestMailer(t, testApi.App)
		tests := []apis.ApiScenario{
			{
				Name:           "Test signup success",
				Method:         http.MethodPost,
				URL:            "/auth/signout",
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					user := core.CreateUserWithOptions(t, app)
					tokenHeader, refreshToken := core.CreateAccessHeaderAndRefreshToken(t, app, user.User.Email)
					dto := apis.SignoutDto{
						RefreshToken: refreshToken,
					}
					data, err := json.Marshal(dto)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Headers = []string{tokenHeader}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					count := repository.MustCountAllCtx(t, t.Context(), repository.Token, app.Db(), nil)
					if count != 0 {
						t.Errorf("Expected token count to be 0, got %v", count)
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
