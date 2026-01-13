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
	"github.com/tkahng/playground/internal/test"
)

func TestApi_SignIn(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		tests := []apis.ApiScenario{
			{
				Name:           "Test signin success",
				Method:         http.MethodPost,
				URL:            "/auth/signin",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					_ = core.CreateUserWithOptions(
						t,
						testApi.App,
						core.UserWithPassword("Password123!"),
						core.UserWithEmail("test@example.com"),
					)
					dto := apis.SigninDto{
						Email:    "test@example.com",
						Password: "Password123!",
					}
					data, err := json.Marshal(dto)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiOutput[*apis.ApiUserInfoTokens]
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
