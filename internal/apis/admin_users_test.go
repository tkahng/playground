package apis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	apphttp "github.com/tkahng/playground/internal/tools/http"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestApi_AdminUsersList(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("admin@k2dv.io"), core.UserWithPermission(shared.PermissionNameAdmin))
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		user1 := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("user1@example.com"))
		user2 := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("user2@example.com"))
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
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
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
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("admin@k2dv.io"), core.UserWithPermission(shared.PermissionNameAdmin))
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		scenarios := []ApiScenario{
			{
				Name:           "admin users create",
				Method:         http.MethodPost,
				URL:            "/admin/users",
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					input := &apis.UserCreateInput{
						Password: "Password123!",
						UserMutationInput: &apis.UserMutationInput{
							Email: "tkahng+01@gmail.com",
							Name:  types.Pointer("John Doe"),
						},
					}
					data, err := json.Marshal(input)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiUser
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if body.Email != "tkahng+01@gmail.com" {
						t.Errorf("Expected user email to be tkahng+01@gmail.com, got %s", body.Email)
					}
					if body.Name != nil {
						name := *body.Name
						if name != "John Doe" {
							t.Errorf("Expected user name to be John Doe, got %s", *body.Name)
						}
					}
				},
			},
			{
				Name:           "admin users create email exists",
				Method:         http.MethodPost,
				URL:            "/admin/users",
				ExpectedStatus: http.StatusConflict,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					input := &apis.UserCreateInput{
						Password: "Password123!",
						UserMutationInput: &apis.UserMutationInput{
							Email: "tkahng+01@gmail.com",
							Name:  types.Pointer("John Doe"),
						},
					}
					data, err := json.Marshal(input)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apphttp.ErrorModel
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if body.Detail != "User already exists" {
						t.Errorf("Expected detail to be User already exists, got %s", body.Detail)
					}
					if body.Status != http.StatusConflict {
						t.Errorf("Expected status to be %d, got %d", http.StatusConflict, body.Status)
					}
					if body.Title != "Conflict" {
						t.Errorf("Expected title to be Conflict, got %s", body.Title)
					}
				},
			},
		}
		for _, tt := range scenarios {
			tt.Test(t)
		}
	})
}

func TestApi_AdminUsersDelete(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("admin@k2dv.io"), core.UserWithPermission(shared.PermissionNameAdmin))
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		scenarios := []ApiScenario{
			{
				Name:           "admin users delete user",
				Method:         http.MethodDelete,
				URL:            "/admin/users/",
				ExpectedStatus: http.StatusNoContent,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					user1 := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("user1@example.com"))
					scenario.URL = fmt.Sprintf("/admin/users/%s", user1.User.ID)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					user, err := app.Adapter().User().FindUser(ctx, &stores.UserFilter{
						Emails: []string{"user1@example.com"},
					})
					if err != nil {
						t.Errorf("Error getting user: %v", err)
					}
					if user != nil {
						t.Errorf("Expected user to be nil, got %v", user)
					}
				},
			},
			{
				Name:           "admin users delete non existing user",
				Method:         http.MethodDelete,
				URL:            "/admin/users/",
				ExpectedStatus: http.StatusNotFound,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					scenario.URL = fmt.Sprintf("/admin/users/%s", uuid.New().String())
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apphttp.ErrorModel
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if body.Detail != "User not found" {
						t.Errorf("Expected detail to be User not found, got %s", body.Detail)
					}
					if body.Status != http.StatusNotFound {
						t.Errorf("Expected status to be %d, got %d", http.StatusNotFound, body.Status)
					}
					if body.Title != "Not Found" {
						t.Errorf("Expected title to be Not Found, got %s", body.Title)
					}
				},
			},
		}
		for _, tt := range scenarios {
			tt.Test(t)
		}
	})
}
func TestApi_AdminUsersUpdate(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("admin@k2dv.io"), core.UserWithPermission(shared.PermissionNameAdmin))
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		scenarios := []ApiScenario{
			{
				Name:           "admin users update user",
				Method:         http.MethodPut,
				URL:            "/admin/users/",
				ExpectedStatus: http.StatusNoContent,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					user1 := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("user1@example.com"), core.UserWithName("User 1"))
					scenario.URL = fmt.Sprintf("/admin/users/%s", user1.User.ID)
					input := &apis.UserMutationInput{
						Email: "user1@example.com",
						Name:  types.Pointer("User 2"),
					}
					scenario.Body = JsonToReader(t, input)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					user, err := app.Adapter().User().FindUser(ctx, &stores.UserFilter{
						Emails: []string{"user1@example.com"},
					})
					if err != nil {
						t.Errorf("Error getting user: %v", err)
					}
					if user == nil {
						t.Errorf("Expected user to not be nil, got %v", user)
					}
					if user.Name != nil && *user.Name != "User 2" {
						t.Errorf("Expected user name to be User 2, got %s", *user.Name)
					}
				},
			},
			{
				Name:           "admin users update non existing user",
				Method:         http.MethodPut,
				URL:            "/admin/users/",
				ExpectedStatus: http.StatusNotFound,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					scenario.URL = fmt.Sprintf("/admin/users/%s", uuid.New().String())
					input := &apis.UserMutationInput{
						Email: "user1@example.com",
						Name:  types.Pointer("User 2"),
					}
					scenario.Body = JsonToReader(t, input)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apphttp.ErrorModel
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if body.Detail != "User not found" {
						t.Errorf("Expected detail to be User not found, got %s", body.Detail)
					}
					if body.Status != http.StatusNotFound {
						t.Errorf("Expected status to be %d, got %d", http.StatusNotFound, body.Status)
					}
					if body.Title != "Not Found" {
						t.Errorf("Expected title to be Not Found, got %s", body.Title)
					}
				},
			},
		}
		for _, tt := range scenarios {
			tt.Test(t)
		}
	})
}
func TestApi_AdminUsersUpdatePassword(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("admin@k2dv.io"), core.UserWithPermission(shared.PermissionNameAdmin))
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		scenarios := []ApiScenario{
			{
				Name:           "admin users update user password",
				Method:         http.MethodPut,
				URL:            "/admin/users/",
				ExpectedStatus: http.StatusNoContent,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					user1 := core.CreateUserWithOptions(
						t,
						testApi.App,
						core.UserWithEmail("user1@example.com"),
						core.UserWithName("User 1"),
						core.UserWithPassword("password"),
						core.UserWithProvider(models.ProvidersCredentials),
						core.UserWithProviderType(models.ProviderTypeCredentials),
					)
					scenario.URL = fmt.Sprintf("/admin/users/%s/password", user1.User.ID)
					input := &apis.UpdateUserPasswordInput{
						Password: "password2",
					}
					scenario.Body = JsonToReader(t, input)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					user, err := app.Adapter().User().FindUser(ctx, &stores.UserFilter{
						Emails: []string{"user1@example.com"},
					})
					if err != nil {
						t.Errorf("Error getting user: %v", err)
					}
					if user == nil {
						t.Errorf("Expected user to not be nil, got %v", user)
					}
					account, err := app.Adapter().UserAccount().FindUserAccount(ctx, &stores.UserAccountFilter{
						UserIds: []uuid.UUID{
							user.ID,
						},
					})
					if err != nil {
						t.Errorf("Error getting user account: %v", err)
					}
					if account == nil {
						t.Errorf("Expected user account to not be nil, got %v", account)
					}
					if account.Password == nil {
						t.Errorf("Expected user account password to not be nil, got %v", account.Password)
					}
					match, err := app.Hash().Verify("password2", *account.Password)
					if err != nil {
						t.Errorf("Error verifying password: %v", err)
					}
					if !match {
						t.Errorf("Expected user account password to be password2, got %s", *account.Password)
					}
				},
			},
		}
		for _, tt := range scenarios {
			tt.Test(t)
		}
	})
}

func TestApi_AdminUsersGet(t *testing.T) {
	// t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		adminUser := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("admin@k2dv.io"), core.UserWithPermission(shared.PermissionNameAdmin))
		header := core.CreateTokenHeader(t, testApi.App, adminUser.User.Email)
		user1 := core.CreateUserWithOptions(t, testApi.App, core.UserWithEmail("user1@example.com"))
		tests := []ApiScenario{
			{
				Name:           "admin users get user1",
				Method:         http.MethodGet,
				URL:            "/admin/users/" + user1.User.ID.String(),
				ExpectedStatus: http.StatusOK,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiUser
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if body.Email != user1.User.Email {
						t.Errorf("Expected user email to be user1@example.com, got %s", body.Email)
					}
				},
			},
			{
				Name:           "admin users get not existing user",
				Method:         http.MethodGet,
				URL:            "/admin/users/" + uuid.NewString(),
				ExpectedStatus: http.StatusNotFound,
				Headers:        []string{header},
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					var body apphttp.ErrorModel
					err := json.NewDecoder(res.Body).Decode(&body)
					if err != nil {
						t.Errorf("Error decoding response: %v", err)
					}
					if body.Detail != "User not found" {
						t.Errorf("Expected error detail to be User not found, got %s", body.Detail)
					}
					if body.Title != "Not Found" {
						t.Errorf("Expected error title to be Not Found, got %s", body.Title)
					}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}
