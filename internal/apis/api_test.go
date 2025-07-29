package apis_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/test"
)

func createTokenHeader(t testing.TB, app core.App, email string) string {
	t.Helper()
	tokensVerifiedTokens, err := app.Auth().CreateAuthTokensFromEmail(context.Background(), email)
	if err != nil {
		t.Errorf("Error creating auth tokens: %v", err)
	}
	VerifiedHeader := fmt.Sprintf("Authorization: Bearer %s", tokensVerifiedTokens.Tokens.AccessToken)
	return VerifiedHeader

}

func createTeamAndMember(app core.App, user *models.User, teamName string) (*models.TeamInfoModel, error) {

	team, err := app.Adapter().TeamGroup().CreateTeam(context.Background(), teamName, strings.TrimSpace(teamName))
	if err != nil {
		return nil, err
	}
	member, err := app.Adapter().TeamMember().CreateTeamMember(context.Background(), team.ID, user.ID, models.TeamMemberRoleOwner, true)
	if err != nil {
		return nil, err
	}
	return &models.TeamInfoModel{
		Team: *team,
		User: models.User{
			ID:              user.ID,
			Name:            user.Name,
			EmailVerifiedAt: user.EmailVerifiedAt,
		},
		Member: *member,
	}, nil
}

func findOrCreateRolePermission(t testing.TB, app core.App, permissionName string) *models.RolePermission {
	ctx := context.Background()
	perm, err := app.Adapter().Rbac().FindOrCreatePermission(ctx, permissionName)
	if err != nil {
		t.Fatalf("FindOrCreatePermission() error = %v", err)
	}
	role, err := app.Adapter().Rbac().FindOrCreateRole(ctx, permissionName)
	if err != nil {
		t.Fatalf("FindOrCreateRole() error = %v", err)
	}
	err = app.Adapter().Rbac().CreateRolePermissions(ctx, role.ID, perm.ID)
	if err != nil {
		t.Fatalf("CreateRolePermissions() error = %v, roleID: %v, permissionID: %v", err, role.ID, perm.ID)
	}
	return &models.RolePermission{
		RoleID:       perm.ID,
		PermissionID: role.ID,
	}
}

func createAdminUser(t testing.TB, app core.App) *models.UserInfo {
	ctx := context.Background()
	nw := time.Now()
	user, err := app.Adapter().User().CreateUser(ctx, &models.User{
		Email:           "admin@k2dv.io",
		EmailVerifiedAt: &nw,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	_, err = app.Adapter().UserAccount().CreateUserAccount(ctx, &models.UserAccount{
		UserID:            user.ID,
		Provider:          models.ProvidersCredentials,
		Type:              models.ProviderTypeCredentials,
		ProviderAccountID: "admin@k2dv.io",
	})
	if err != nil {
		t.Fatalf("CreateUserAccount() error = %v", err)
	}
	err = app.Adapter().Rbac().EnsureRoleAndPermissions(ctx, shared.PermissionNameAdmin, shared.PermissionNameAdmin)
	if err != nil {
		t.Fatalf("EnsureRoleAndPermissions() error = %v", err)
	}
	perm, err := app.Adapter().Rbac().FindOrCreatePermission(ctx, shared.PermissionNameAdmin)
	if err != nil {
		t.Fatalf("FindOrCreatePermission() error = %v", err)
	}
	err = app.Adapter().Rbac().CreateUserPermissions(ctx, user.ID, perm.ID)
	if err != nil {
		t.Fatalf("CreateUserAccount() error = %v", err)
	}
	return &models.UserInfo{
		User: *user,
	}
}
func createVerifiedUser(app core.App) (*models.UserInfo, error) {
	nw := time.Now()
	user, err := app.Adapter().User().CreateUser(context.Background(), &models.User{
		Email:           "authenticated@example.com",
		EmailVerifiedAt: &nw,
	})
	if err != nil {
		return nil, err
	}
	_, err = app.Adapter().UserAccount().CreateUserAccount(context.Background(), &models.UserAccount{
		UserID:            user.ID,
		Provider:          models.ProvidersGoogle,
		Type:              "oauth",
		ProviderAccountID: "google-123",
	})
	if err != nil {
		return nil, err
	}
	return &models.UserInfo{
		User: *user,
	}, nil
}
func createUnverifiedUser(app *core.BaseAppDecorator) (*models.UserInfo, error) {
	user, err := app.Adapter().User().CreateUser(context.Background(), &models.User{
		Email: "authenticated@example.com",
	})
	if err != nil {
		return nil, err
	}
	_, err = app.Adapter().UserAccount().CreateUserAccount(context.Background(), &models.UserAccount{
		UserID:            user.ID,
		Provider:          models.ProvidersGoogle,
		Type:              "oauth",
		ProviderAccountID: "google-123",
	})
	if err != nil {
		return nil, err
	}
	return &models.UserInfo{
		User: *user,
	}, nil
}

type TestApi struct {
	TestApi humatest.TestAPI
	Api     apis.Api
	App     *core.BaseAppDecorator
	Cfg     conf.EnvConfig
	Router  http.Handler
}

func SetupApi(t testing.TB, ctx context.Context, db database.Dbx) *TestApi {
	t.Helper()
	cfg := conf.ZeroEnvConfig()
	app := core.NewAppDecorator(ctx, cfg, db)
	appApi := apis.NewAppApi(app)
	router, api := test.NewHumaApi(t)
	apis.AddRoutes(api, appApi)
	testApi := &TestApi{
		TestApi: api,
		Api:     *appApi,
		App:     app,
		Cfg:     cfg,
		Router:  router,
	}
	return testApi
}

// ApiScenario defines a single api request test case/scenario.
type ApiScenario struct {
	// Name is the test name.
	Name string

	// Method is the HTTP method of the test request to use.
	Method string

	// URL is the url/path of the endpoint you want to test.
	URL string

	// Body specifies the body to send with the request.
	//
	// For example:
	//
	//	strings.NewReader(`{"title":"abc"}`)
	Body io.Reader

	// Headers specifies the headers to send with the request (e.g. "Authorization": "abc")
	Headers []string

	// Delay adds a delay before checking the expectations usually
	// to ensure that all fired non-awaited go routines have finished
	Delay time.Duration

	// Timeout specifies how long to wait before cancelling the request context.
	//
	// A zero or negative value means that there will be no timeout.
	Timeout time.Duration

	// expectations
	// ---------------------------------------------------------------

	// ExpectedStatus specifies the expected response HTTP status code.
	ExpectedStatus int

	// List of keywords that MUST exist in the response body.
	//
	// Either ExpectedContent or NotExpectedContent must be set if the response body is non-empty.
	// Leave both fields empty if you want to ensure that the response didn't have any body (e.g. 204).
	ExpectedContent []string

	// List of keywords that MUST NOT exist in the response body.
	//
	// Either ExpectedContent or NotExpectedContent must be set if the response body is non-empty.
	// Leave both fields empty if you want to ensure that the response didn't have any body (e.g. 204).
	NotExpectedContent []string

	// List of hook events to check whether they were fired or not.
	//
	// You can use the wildcard "*" event key if you want to ensure
	// that no other hook events except those listed have been fired.
	//
	// For example:
	//
	//	map[string]int{ "*": 0 } // no hook events were fired
	//	map[string]int{ "*": 0, "EventA": 2 } // no hook events, except EventA were fired
	//	map[string]int{ EventA": 2, "EventB": 0 } // ensures that EventA was fired exactly 2 times and EventB exactly 0 times.
	// ExpectedEvents map[string]int

	// test hooks
	// ---------------------------------------------------------------

	TestAppFactory func(t testing.TB) *TestApi
	BeforeTestFunc func(t testing.TB, app *core.BaseAppDecorator, scenario *ApiScenario)
	AfterTestFunc  func(t testing.TB, app *core.BaseAppDecorator, res *httptest.ResponseRecorder)
}

// Test executes the test scenario.
//
// Example:
//
//	func TestListExample(t *testing.T) {
//	    scenario := tests.ApiScenario{
//	        Name:           "list example collection",
//	        Method:         http.MethodGet,
//	        URL:            "/api/collections/example/records",
//	        ExpectedStatus: 200,
//	        ExpectedContent: []string{
//	            `"totalItems":3`,
//	            `"id":"0yxhwia2amd8gec"`,
//	            `"id":"achvryl401bhse3"`,
//	            `"id":"llvuca81nly1qls"`,
//	        },
//	        ExpectedEvents: map[string]int{
//	            "OnRecordsListRequest": 1,
//	            "OnRecordEnrich":       3,
//	        },
//	    }
//
//	    scenario.Test(t)
//	}
func (scenario *ApiScenario) Test(t *testing.T) {
	t.Run(scenario.normalizedName(), func(t *testing.T) {
		scenario.test(t)
	})
}

// Benchmark benchmarks the test scenario.
//
// Example:
//
//	func BenchmarkListExample(b *testing.B) {
//	    scenario := tests.ApiScenario{
//	        Name:           "list example collection",
//	        Method:         http.MethodGet,
//	        URL:            "/api/collections/example/records",
//	        ExpectedStatus: 200,
//	        ExpectedContent: []string{
//	            `"totalItems":3`,
//	            `"id":"0yxhwia2amd8gec"`,
//	            `"id":"achvryl401bhse3"`,
//	            `"id":"llvuca81nly1qls"`,
//	        },
//	        ExpectedEvents: map[string]int{
//	            "OnRecordsListRequest": 1,
//	            "OnRecordEnrich":       3,
//	        },
//	    }
//
//	    scenario.Benchmark(b)
//	}
func (scenario *ApiScenario) Benchmark(b *testing.B) {
	b.Run(scenario.normalizedName(), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			scenario.test(b)
		}
	})
}

func (scenario *ApiScenario) normalizedName() string {
	var name = scenario.Name

	if name == "" {
		name = fmt.Sprintf("%s:%s", scenario.Method, scenario.URL)
	}

	return name
}

func (scenario *ApiScenario) test(t testing.TB) {
	testApi := scenario.TestAppFactory(t)
	if scenario.BeforeTestFunc != nil {
		scenario.BeforeTestFunc(t, testApi.App, scenario)
	}
	var args []any
	if len(scenario.Headers) != 0 {
		for _, header := range scenario.Headers {
			args = append(args, header)
		}
	}
	if scenario.Body != nil {
		args = append(args, scenario.Body)
	}
	recorder := testApi.TestApi.Do(scenario.Method, scenario.URL, args...)

	if recorder.Code != scenario.ExpectedStatus {
		t.Errorf("Expected status code %d, got %d", scenario.ExpectedStatus, recorder.Code)
	}
	if len(scenario.ExpectedContent) == 0 && len(scenario.NotExpectedContent) == 0 && scenario.AfterTestFunc == nil {
		if len(recorder.Body.String()) != 0 {
			t.Errorf("Expected empty body, got \n%v", recorder.Body.String())
		}
	} else {
		// normalize json response format
		buffer := new(bytes.Buffer)
		err := json.Compact(buffer, recorder.Body.Bytes())
		var normalizedBody string
		if err != nil {
			// not a json...
			normalizedBody = recorder.Body.String()
		} else {
			normalizedBody = buffer.String()
		}

		for _, item := range scenario.ExpectedContent {
			if !strings.Contains(normalizedBody, item) {
				t.Errorf("Cannot find %v in response body \n%v", item, normalizedBody)
				break
			}
		}

		for _, item := range scenario.NotExpectedContent {
			if strings.Contains(normalizedBody, item) {
				t.Errorf("Didn't expect %v in response body \n%v", item, normalizedBody)
				break
			}
		}
	}
	if scenario.AfterTestFunc != nil {
		scenario.AfterTestFunc(t, testApi.App, recorder)
	}
}
