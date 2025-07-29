package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_AdminUsersList(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		type args struct {
			url  string
			args []any
		}
		tests := []ApiScenario{
			{
				Name:               "admin users list",
				Method:             http.MethodGet,
				URL:                "/api/admin/users",
				ExpectedStatus:     http.StatusOK,
				ExpectedContent:    []string{"total: 1"},
				NotExpectedContent: []string{},
				TestAppFactory: func(t testing.TB) *TestApi {
					return SetupApi(t, ctx, db)
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseAppDecorator, scenario *ApiScenario) {
					adminUser := createAdminUser(t, app)
					header := createTokenHeader(t, app, adminUser.User.Email)
					scenario.Headers = append(scenario.Headers, header)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseAppDecorator, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.ApiUser]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != 1 {
						t.Errorf("Expected 1 user, got %d", len(body.Data))
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func TestApi_AdminUsersCreate(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		// testApi := SetupApi(t, ctx, db)
		// api := testApi.TestApi
		// user, err := createAdminUser(testApi.App)
		// if err != nil {
		// 	t.Errorf("Error creating user: %v", err)
		// 	return
		// }
		// tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(context.Background(), user.User.Email)
		// if err != nil {
		// 	t.Errorf("Error creating auth tokens: %v", err)
		// 	return
		// }
		// VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
		// resp := api.Post("/admin/users", VerifiedHeader, &apis.CreateAdminUserInput{
		// 	Email: "user1@example",
		// })
		// if resp.Code == 200 {
		// 	t.Fatalf("Unexpected response: %s", resp.Body.String())
		// }
	})
}
