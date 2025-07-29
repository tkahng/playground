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
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/test"
)

func TestApi_AdminUsersList(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		adminUser := CreateUserWithOptions(t, testApi.App, UserWithEmail("admin@k2dv.io"), UserWithPermission(shared.PermissionNameAdmin))
		header := createTokenHeader(t, testApi.App, adminUser.User.Email)
		user1 := CreateUserWithOptions(t, testApi.App, UserWithEmail("user1@example.com"))
		user2 := CreateUserWithOptions(t, testApi.App, UserWithEmail("user2@example.com"))
		tests := []ApiScenario{
			{
				Name:           "admin users list all",
				Method:         http.MethodGet,
				URL:            "/admin/users",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseAppDecorator, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.ApiUser]
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if len(body.Data) != 3 {
						t.Errorf("Expected 3 user, got %d", len(body.Data))
					}
				},
			},
			{
				Name:           "admin users list filter email",
				Method:         http.MethodGet,
				URL:            "/admin/users?emails=" + user1.User.Email,
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
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
					if body.Data[0].Email != user1.User.Email {
						t.Errorf("Expected user email to be user1@example.com, got %s", body.Data[0].Email)
					}
				},
			},
			{
				Name:           "admin users list filter email 2",
				Method:         http.MethodGet,
				URL:            "/admin/users?emails=" + user2.User.Email,
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
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
					if body.Data[0].Email != user2.User.Email {
						t.Errorf("Expected user email to be user1@example.com, got %s", body.Data[0].Email)
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
