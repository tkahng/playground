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

func TestApi_RefreshToken(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithNewTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)

		userInfo := CreateUserWithOptions(t, testApi.App, UserWithPassword("Password123!"))

		tests := []ApiScenario{
			{
				Name:           "Test refresh token success",
				Method:         http.MethodPost,
				URL:            "/auth/refresh-token",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokens, err := app.Auth().CreateAuthTokensFromEmail(ctx, userInfo.User.Email)
					if err != nil {
						t.Errorf("Error creating auth tokens: %v", err)
					}
					dto := apis.RefreshTokenInput{
						RefreshToken: tokens.Tokens.RefreshToken,
					}
					data, err := json.Marshal(dto)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiOutput[*apis.ApiUserInfoTokens]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					// tokens, err := app.Adapter().Token().GetToken(ctx, )
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
