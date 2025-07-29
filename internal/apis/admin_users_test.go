package apis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	apphttp "github.com/tkahng/playground/internal/tools/http"
	"github.com/tkahng/playground/internal/tools/types"
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
		testApi := SetupApi(t, ctx, db)
		adminUser := CreateUserWithOptions(t, testApi.App, UserWithEmail("admin@k2dv.io"), UserWithPermission(shared.PermissionNameAdmin))
		header := createTokenHeader(t, testApi.App, adminUser.User.Email)
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
				BeforeTestFunc: func(t testing.TB, app *core.BaseAppDecorator, scenario *ApiScenario) {
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
				AfterTestFunc: func(t testing.TB, app *core.BaseAppDecorator, res *httptest.ResponseRecorder) {
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
				BeforeTestFunc: func(t testing.TB, app *core.BaseAppDecorator, scenario *ApiScenario) {
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
				AfterTestFunc: func(t testing.TB, app *core.BaseAppDecorator, res *httptest.ResponseRecorder) {
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
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		adminUser := CreateUserWithOptions(t, testApi.App, UserWithEmail("admin@k2dv.io"), UserWithPermission(shared.PermissionNameAdmin))
		header := createTokenHeader(t, testApi.App, adminUser.User.Email)
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
				BeforeTestFunc: func(t testing.TB, app *core.BaseAppDecorator, scenario *ApiScenario) {
					user1 := CreateUserWithOptions(t, testApi.App, UserWithEmail("user1@example.com"))
					scenario.URL = fmt.Sprintf("/admin/users/%s", user1.User.ID)
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseAppDecorator, res *httptest.ResponseRecorder) {
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
		}
		for _, tt := range scenarios {
			tt.Test(t)
		}
	})
}
