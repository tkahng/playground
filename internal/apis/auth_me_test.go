package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
)

func TestApi_Me(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		tests := []ApiScenario{
			{
				Name:           "success: authorized",
				Method:         http.MethodGet,
				URL:            "/auth/me",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					user := core.CreateUserWithOptions(t, app, core.UserWithEmail("me@me.com"))
					authToken := core.CreateTokenHeader(t, app, user.User.Email)
					scenario.Headers = []string{authToken}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.UserWithAccounts
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					assert.Equal(t, "me@me.com", body.ApiUser.Email)
				},
			},
			{
				Name:           "fail: unauthorized",
				Method:         http.MethodGet,
				URL:            "/auth/me",
				ExpectedStatus: http.StatusUnauthorized,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				ExpectedContent: []string{
					"you are not authenticated.",
					"Unauthorized",
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body huma.ErrorModel
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
